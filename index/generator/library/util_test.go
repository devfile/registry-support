package library

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/devfile/registry-support/index/generator/schema"
	"github.com/stretchr/testify/assert"
)

const (
	zipType string = "application/zip"
)

func TestDownloadRemoteStack(t *testing.T) {
	tests := []struct {
		name string
		git  *schema.Git
		path string
	}{
		{
			"Case 1: Maven Java (Without subDir)",
			&schema.Git{
				Url:        "https://github.com/odo-devfiles/springboot-ex.git",
				RemoteName: "origin",
			},
			filepath.Join(os.TempDir(), "springboot-ex"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hiddenGitPath := filepath.Join(tt.path, ".git")

			if err := DownloadRemoteStack(tt.git, tt.path, false); err != nil {
				t.Errorf("Git download to bytes failed: %v", err)
			}

			if _, err := os.Stat(tt.path); os.IsNotExist(err) {
				t.Errorf("%s does not exist but is suppose to", tt.path)
			} else if _, err := os.Stat(hiddenGitPath); os.IsExist(err) {
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
		name   string
		params map[string]string
	}{
		{
			"Case 1: Java Quarkus (Without subDir)",
			map[string]string{
				"Name":   "quarkus",
				"ZipUrl": "https://code.quarkus.io/d?e=io.quarkus%3Aquarkus-resteasy&e=io.quarkus%3Aquarkus-micrometer&e=io.quarkus%3Aquarkus-smallrye-health&e=io.quarkus%3Aquarkus-openshift&cn=devfile",
				"SubDir": "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(os.TempDir(), tt.params["Name"])
			zipPath := fmt.Sprintf("%s.zip", path)
			bytes, err := DownloadStackFromZipUrl(tt.params["ZipUrl"], tt.params["SubDir"], path)

			if err != nil {
				t.Errorf("Zip download to bytes failed: %v", err)
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
		name string
		git  *schema.Git
		path string
	}{
		{
			"Case 1: Maven Java (Without subDir)",
			&schema.Git{
				Url:        "https://github.com/odo-devfiles/springboot-ex.git",
				RemoteName: "origin",
			},
			filepath.Join(os.TempDir(), "springboot-ex"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hiddenGitPath := filepath.Join(tt.path, ".git")
			zipPath := fmt.Sprintf("%s.zip", tt.path)
			bytes, err := DownloadStackFromGit(tt.git, tt.path, false)

			if err != nil {
				t.Errorf("Git download to bytes failed: %v", err)
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

func TestGetSubDir(t *testing.T) {
	assert.Fail(t, "Not Implemented.")
}

func TestCopyFileWithFs(t *testing.T) {
	assert.Fail(t, "Not Implemented.")
}

func TestCopyDirWithFS(t *testing.T) {
	assert.Fail(t, "Not Implemented.")
}

func TestCleanDir(t *testing.T) {
	assert.Fail(t, "Not Implemented.")
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
