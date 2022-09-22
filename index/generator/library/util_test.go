//
// Copyright 2022 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package library

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/devfile/registry-support/index/generator/schema"
)

const (
	zipType string = "application/zip"
)

func TestCloneRemoteStack(t *testing.T) {
	tests := []struct {
		name       string
		git        *schema.Git
		path       string
		wantErr    bool
		wantErrStr string
	}{
		{
			name: "Case 1: Maven Java",
			git: &schema.Git{
				Url:        "https://github.com/odo-devfiles/springboot-ex.git",
				RemoteName: "origin",
			},
			path:       filepath.Join(os.TempDir(), "springboot-ex"),
			wantErr:    false,
			wantErrStr: "",
		},
		{
			name: "Case 2: Maven Java (With subDir)",
			git: &schema.Git{
				Url:        "https://github.com/odo-devfiles/springboot-ex.git",
				RemoteName: "origin",
				SubDir:     "src/main",
			},
			path:       filepath.Join(os.TempDir(), "springboot-ex"),
			wantErr:    false,
			wantErrStr: "",
		},
		{
			name: "Case 3: Wildfly Java - microprofile-config subdirectory",
			git: &schema.Git{
				Url:        "https://github.com/wildfly/quickstart.git",
				RemoteName: "wildfly-quickstart",
				Revision:   "22.0.1.Final",
				SubDir:     "microprofile-config",
			},
			path:       filepath.Join(os.TempDir(), "quickstart"),
			wantErr:    false,
			wantErrStr: "",
		},
		{
			name: "Case 4: Maven Java - Cloning with Hash Revision",
			git: &schema.Git{
				Url:        "https://github.com/odo-devfiles/springboot-ex.git",
				RemoteName: "origin",
				Revision:   "694e96286ffdc3a9990d0041637d32cecba38181",
			},
			path:       filepath.Join(os.TempDir(), "springboot-ex"),
			wantErr:    true,
			wantErrStr: "specifying commit in 'revision' is not yet supported",
		},
		{
			name: "Case 5: Cloning a non-existent repo",
			git: &schema.Git{
				Url:        "https://github.com/odo-devfiles/nonexist.git",
				RemoteName: "origin",
			},
			path:       filepath.Join(os.TempDir(), "nonexist"),
			wantErr:    true,
			wantErrStr: "",
		},
		{
			name: "Case 6: Maven Java - Cloning with Invalid Revision",
			git: &schema.Git{
				Url:        "https://github.com/odo-devfiles/springboot-ex.git",
				RemoteName: "origin",
				Revision:   "invalid",
			},
			path:       filepath.Join(os.TempDir(), "springboot-ex"),
			wantErr:    true,
			wantErrStr: "couldn't find remote ref \"refs/tags/invalid\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hiddenGitPath := filepath.Join(tt.path, ".git")

			if gotErr := CloneRemoteStack(tt.git, tt.path, false); gotErr != nil {
				if !tt.wantErr || (tt.wantErrStr != "" && gotErr.Error() != tt.wantErrStr) {
					t.Errorf("Git download to bytes failed: %v", gotErr)
				}
				return
			}

			if _, gotErr := os.Stat(tt.path); os.IsNotExist(gotErr) {
				t.Errorf("%s does not exist but is suppose to", tt.path)
			} else if _, gotErr := os.Stat(hiddenGitPath); os.IsExist(gotErr) {
				t.Errorf(".git exist but isn't suppose to within %s", tt.path)
			}

			if err := os.RemoveAll(tt.path); err != nil {
				t.Logf("Deleting %s failed.", tt.path)
			}
		})
	}
}

func TestDownloadStackFromZipUrl(t *testing.T) {
	tests := []struct {
		name       string
		params     map[string]string
		wantErr    bool
		wantErrStr string
	}{
		{
			name: "Case 1: Java Quarkus",
			params: map[string]string{
				"Name":   "quarkus",
				"ZipUrl": "https://code.quarkus.io/d?e=io.quarkus%3Aquarkus-resteasy&e=io.quarkus%3Aquarkus-micrometer&e=io.quarkus%3Aquarkus-smallrye-health&e=io.quarkus%3Aquarkus-openshift&cn=devfile",
				"SubDir": "",
			},
			wantErr:    false,
			wantErrStr: "",
		},
		{
			name: "Case 2: Java Quarkus (With subDir)",
			params: map[string]string{
				"Name":   "quarkus",
				"ZipUrl": "https://code.quarkus.io/d?e=io.quarkus%3Aquarkus-resteasy&e=io.quarkus%3Aquarkus-micrometer&e=io.quarkus%3Aquarkus-smallrye-health&e=io.quarkus%3Aquarkus-openshift&cn=devfile",
				"SubDir": "src",
			},
			wantErr:    false,
			wantErrStr: "",
		},
		{
			name: "Case 3: Download error",
			params: map[string]string{
				"Name":   "quarkus",
				"ZipUrl": "https://code.quarkus.io/d?e=io.quarkus",
				"SubDir": "",
			},
			wantErr:    true,
			wantErrStr: "failed to retrieve https://code.quarkus.io/d?e=io.quarkus, 400: Bad Request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(os.TempDir(), tt.params["Name"])
			zipPath := fmt.Sprintf("%s.zip", path)
			bytes, err := DownloadStackFromZipUrl(tt.params["ZipUrl"], tt.params["SubDir"], path)

			if err != nil {
				if !tt.wantErr || (tt.wantErrStr != "" && err.Error() != tt.wantErrStr) {
					t.Errorf("Zip download to bytes failed: %v", err)
				}
				return
			}

			resultantType := http.DetectContentType(bytes)

			if resultantType != zipType {
				t.Errorf("Content type of download not matching expected. Expected: %s, Actual: %s",
					zipType, resultantType)
			}

			if err := os.RemoveAll(path); err != nil {
				t.Logf("Deleting %s failed.", path)
			} else if err := os.Remove(zipPath); err != nil {
				t.Logf("Deleting %s failed.", zipPath)
			}
		})
	}
}

func TestDownloadStackFromGit(t *testing.T) {
	tests := []struct {
		name       string
		git        *schema.Git
		path       string
		wantErr    bool
		wantErrStr string
	}{
		{
			name: "Case 1: Maven Java",
			git: &schema.Git{
				Url:        "https://github.com/odo-devfiles/springboot-ex.git",
				RemoteName: "origin",
			},
			path:       filepath.Join(os.TempDir(), "springboot-ex"),
			wantErr:    false,
			wantErrStr: "",
		},
		{
			name: "Case 2: Maven Java (With subDir)",
			git: &schema.Git{
				Url:        "https://github.com/odo-devfiles/springboot-ex.git",
				RemoteName: "origin",
				SubDir:     "src/main",
			},
			path:       filepath.Join(os.TempDir(), "springboot-ex"),
			wantErr:    false,
			wantErrStr: "",
		},
		{
			name: "Case 3: Wildfly Java - microprofile-config subdirectory",
			git: &schema.Git{
				Url:        "https://github.com/wildfly/quickstart.git",
				RemoteName: "wildfly-quickstart",
				Revision:   "22.0.1.Final",
				SubDir:     "microprofile-config",
			},
			path:       filepath.Join(os.TempDir(), "quickstart"),
			wantErr:    false,
			wantErrStr: "",
		},
		{
			name: "Case 4: Maven Java - Cloning with Hash Revision",
			git: &schema.Git{
				Url:        "https://github.com/odo-devfiles/springboot-ex.git",
				RemoteName: "origin",
				Revision:   "694e96286ffdc3a9990d0041637d32cecba38181",
			},
			path:       filepath.Join(os.TempDir(), "springboot-ex"),
			wantErr:    true,
			wantErrStr: "specifying commit in 'revision' is not yet supported",
		},
		{
			name: "Case 5: Cloning a non-existent repo",
			git: &schema.Git{
				Url:        "https://github.com/odo-devfiles/nonexist.git",
				RemoteName: "origin",
			},
			path:       filepath.Join(os.TempDir(), "nonexist"),
			wantErr:    true,
			wantErrStr: "",
		},
		{
			name: "Case 6: Maven Java - Cloning with Invalid Revision",
			git: &schema.Git{
				Url:        "https://github.com/odo-devfiles/springboot-ex.git",
				RemoteName: "origin",
				Revision:   "invalid",
			},
			path:       filepath.Join(os.TempDir(), "springboot-ex"),
			wantErr:    true,
			wantErrStr: "couldn't find remote ref \"refs/tags/invalid\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hiddenGitPath := filepath.Join(tt.path, ".git")
			zipPath := fmt.Sprintf("%s.zip", tt.path)
			bytes, err := DownloadStackFromGit(tt.git, tt.path, false)

			if err != nil {
				if !tt.wantErr || (tt.wantErrStr != "" && err.Error() != tt.wantErrStr) {
					t.Errorf("Git download to bytes failed: %v", err)
				}
				return
			} else if _, err := os.Stat(hiddenGitPath); os.IsExist(err) {
				t.Errorf(".git exist but isn't suppose to within %s", hiddenGitPath)
			}

			resultantType := http.DetectContentType(bytes)

			if resultantType != zipType {
				t.Errorf("Content type of download not matching expected. Expected: %s, Actual: %s",
					zipType, resultantType)
			}

			if err := os.RemoveAll(tt.path); err != nil {
				t.Logf("Deleting %s failed.", tt.path)
			} else if err := os.Remove(zipPath); err != nil {
				t.Logf("Deleting %s failed.", zipPath)
			}
		})
	}
}

func TestZipDir(t *testing.T) {
	dirPath := filepath.Join(os.TempDir(), "TestZipDir")
	filePath := filepath.Join(dirPath, "test.txt")
	zipPath := fmt.Sprintf("%s.zip", dirPath)

	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		t.Errorf("Failed to create directory '%s': %v", dirPath, err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		t.Errorf("Failed to create file '%s': %v", filePath, err)
	}

	if _, err = file.WriteString("Hello World!"); err != nil {
		t.Errorf("Failed to write to file '%s': %v", filePath, err)
	}

	file.Close()

	if err = ZipDir(dirPath, zipPath); err != nil {
		t.Errorf("Failed to zip directory '%s': %v", dirPath, err)
	}

	bytes, err := ioutil.ReadFile(zipPath)
	if err != nil {
		t.Errorf("Unable to read zip file '%s': %v", zipPath, err)
	}

	resultantType := http.DetectContentType(bytes)

	if resultantType != zipType {
		t.Errorf("Content type of download not matching expected. Expected: %s, Actual: %s",
			zipType, resultantType)
	}

	if err := os.RemoveAll(dirPath); err != nil {
		t.Logf("Deleting %s failed.", dirPath)
	} else if err := os.Remove(zipPath); err != nil {
		t.Logf("Deleting %s failed.", zipPath)
	}
}
