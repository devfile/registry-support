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

package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	indexSchema "github.com/devfile/registry-support/index/generator/schema"
)

func TestIsHtmlRequested(t *testing.T) {
	tests := []struct {
		name   string
		header []string
		want   bool
	}{
		{
			name:   "Case 1: Empty header",
			header: []string{},
			want:   false,
		},
		{
			name:   "Case 2: Single header, no html",
			header: []string{"application/xml"},
			want:   false,
		},
		{
			name:   "Case 3: Single header, html",
			header: []string{"application/xml,text/html"},
			want:   true,
		},
		{
			name:   "Case 4: Multiple headers, no html",
			header: []string{"Header1", "Header2", "Header3"},
			want:   false,
		},
		{
			name:   "Case 5: Multiple headers, html",
			header: []string{"Header1", "Header2", "Header3", "text/html"},
			want:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			htmlRequested := IsHtmlRequested(test.header)
			if htmlRequested != test.want {
				t.Errorf("Got: %v, Expected: %v", htmlRequested, test.want)
			}
		})
	}
}

func TestEncodeToBase64(t *testing.T) {
	tests := []struct {
		name       string
		uri        string
		wantBase64 string
		wantErr    bool
	}{
		{
			"Case 1: test remote uri is valid",
			"https://raw.githubusercontent.com/maysunfaisal/node-bulletin-board-2/main/nodejs-icon.png",
			"data:image/png;base64,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",
			false,
		},
		{
			"Case 2: test local uri is valid",
			"../../tests/resources/nodejs-icon.png",
			"data:image/png;base64,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",
			false,
		},
		{
			"Case 3: test remote uri is invalid",
			"https://invalid",
			"",
			true,
		},
		{
			"Case 4: test local uri is invalid",
			"./invalid.png",
			"",
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotErr := false
			gotBase64, err := encodeToBase64(test.uri)
			if err != nil {
				gotErr = true
			}
			if gotErr != test.wantErr || gotBase64 != test.wantBase64 {
				t.Errorf("Got error: %t, want error: %t, function return error: %v, got base64: %s, want base64: %s", gotErr, test.wantErr, err, gotBase64, test.wantBase64)
			}
		})
	}
}

func TestEncodeIndexIconToBase64(t *testing.T) {
	const wantBase64IndexPath = "../../tests/expectations/base64_index.json"

	tests := []struct {
		name            string
		indexPath       string
		base64IndexPath string
		wantErr         bool
	}{
		{
			"Case 1: test the generation of index.json with base64 format",
			"../../tests/resources/index.json",
			"../../tests/expectations/temp_base64_index.json",
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotErr := false
			_, err := EncodeIndexIconToBase64(test.indexPath, test.base64IndexPath)
			if err != nil {
				gotErr = true
			}
			if gotErr != test.wantErr {
				t.Errorf("Got error: %t, want error %t, function return error: %v", gotErr, test.wantErr, err)
			}
			gotFs, err := os.Stat(test.base64IndexPath)
			if err != nil {
				t.Errorf("Can't get file info of %s", test.indexPath)
			}
			wantFs, err := os.Stat(wantBase64IndexPath)
			if err != nil {
				t.Errorf("Can't get file info of %s", wantBase64IndexPath)
			}
			if gotFs.Mode() != wantFs.Mode() || gotFs.Size() != wantFs.Size() {
				t.Errorf("Got base64 index %s and want base64 index %s are not identical", test.base64IndexPath, wantBase64IndexPath)
			}

			// Clean up the temp index file
			if _, err = os.Stat(test.base64IndexPath); !os.IsNotExist(err) {
				err = os.Remove(test.base64IndexPath)
				if err != nil {
					t.Errorf("Can't remove the temp base64 index: %s", test.base64IndexPath)
				}
			}
		})
	}
}

func TestGetOptionalEnv(t *testing.T) {
	os.Setenv("SET_BOOL", "true")
	os.Setenv("SET_STRING", "test")

	tests := []struct {
		name         string
		key          string
		defaultValue interface{}
		want         interface{}
	}{
		{
			name:         "Test get SET_BOOL environment variable",
			key:          "SET_BOOL",
			defaultValue: false,
			want:         true,
		},
		{
			name:         "Test get unset bool environment variable",
			key:          "UNSET_BOOL",
			defaultValue: false,
			want:         false,
		},
		{
			name:         "Test get SET_STRING environment variable",
			key:          "SET_STRING",
			defaultValue: "anonymous",
			want:         "test",
		},
		{
			name:         "Test get unset string environment variable",
			key:          "UNSET_STRING",
			defaultValue: "anonymous",
			want:         "anonymous",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := GetOptionalEnv(test.key, test.defaultValue)
			if value != test.want {
				t.Errorf("Got: %v, want: %v", value, test.want)
			}
		})
	}
}

func TestConvertToOldIndexFormat(t *testing.T) {
	const inputIndexFilePath = "../../tests/resources/newIndexStruct.json"
	const wantIndexFilePath = "../../tests/resources/oldIndexStruct.json"
	bytes, err := ioutil.ReadFile(inputIndexFilePath)
	if err != nil {
		t.Errorf("Failed to read newIndexStruct.json: %v", err)
	}
	expected, err := ioutil.ReadFile(wantIndexFilePath)
	if err != nil {
		t.Errorf("Failed to oldIndexStruct.json: %v", err)
	}
	var inputIndex []indexSchema.Schema
	err = json.Unmarshal(bytes, &inputIndex)
	if err != nil {
		t.Errorf("Failed to unmarshal inputIndex json")
	}
	var wantIndex []indexSchema.Schema
	err = json.Unmarshal(expected, &wantIndex)
	if err != nil {
		t.Errorf("Failed to unmarshal wantIndex json")
	}

	t.Run("Test generate index", func(t *testing.T) {
		gotIndex := ConvertToOldIndexFormat(inputIndex)

		if !reflect.DeepEqual(wantIndex, gotIndex) {
			t.Errorf("Want index %v, got index %v", wantIndex, gotIndex)
		}
	})
}

func TestMakeVersionMap(t *testing.T) {
	devfileIndex := indexSchema.Schema{
		Name:              "Test Devfile",
		Version:           "2.2.0",
		Attributes:        nil,
		DisplayName:       "",
		Description:       "",
		Type:              "",
		Tags:              []string{},
		Architectures:     []string{},
		Icon:              "",
		GlobalMemoryLimit: "",
		ProjectType:       "",
		Language:          "",
		Links:             map[string]string{},
		Resources:         []string{},
		StarterProjects:   []string{},
		Git:               &indexSchema.Git{},
		Provider:          "",
		SupportUrl:        "",
		Versions: []indexSchema.Version{
			{
				Version: "1.1.0",
				Default: true,
			},
			{Version: "1.2.0"},
		},
	}
	tests := []struct {
		key     string
		wantVal string
	}{
		{
			key:     "default",
			wantVal: "1.1.0",
		},
		{
			key:     "latest",
			wantVal: "1.2.0",
		},
		{
			key:     "1.1.0",
			wantVal: "1.1.0",
		},
		{
			key:     "",
			wantVal: "",
		},
		{
			key:     "1.3.0",
			wantVal: "",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Test generate version map with key %s", test.key), func(t *testing.T) {
			versionMap, err := MakeVersionMap(devfileIndex)
			if err != nil {
				t.Errorf("Was not expecting error with MakeVersionMap: %v", err)
			}

			if !reflect.DeepEqual(test.wantVal, versionMap[test.key].Version) {
				t.Errorf("Was expecting '%s' to map to '%s' not '%s'",
					test.key, test.wantVal, versionMap[test.key].Version)
			}
		})
	}
}

func TestMakeVersionMapOnBadVersion(t *testing.T) {
	devfileIndex := indexSchema.Schema{
		Name:              "Test Devfile",
		Version:           "2.2.0",
		Attributes:        nil,
		DisplayName:       "",
		Description:       "",
		Type:              "",
		Tags:              []string{},
		Architectures:     []string{},
		Icon:              "",
		GlobalMemoryLimit: "",
		ProjectType:       "",
		Language:          "",
		Links:             map[string]string{},
		Resources:         []string{},
		StarterProjects:   []string{},
		Git:               &indexSchema.Git{},
		Provider:          "",
		SupportUrl:        "",
		Versions: []indexSchema.Version{
			{Version: "fsdf-sf3v.-dfg"},
			{Version: "erdgf-v.-dd-,.fdgg"},
		},
	}
	t.Run("Test generate version map with bad versioning", func(t *testing.T) {
		_, err := MakeVersionMap(devfileIndex)
		if err == nil {
			t.Error("Was expecting malformed version error with MakeVersionMap")
		}
	})
}
