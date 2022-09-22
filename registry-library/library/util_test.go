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
	"reflect"
	"testing"
)

func TestValidateStackVersionTag(t *testing.T) {
	tests := []struct {
		name       string
		stackTag   string
		wantResult bool
		wantErr    bool
	}{
		{
			name:       "Stack without version",
			stackTag:   "java-maven",
			wantResult: true,
		},
		{
			name:       "Stack with specific version",
			stackTag:   "java-maven:1.0.0",
			wantResult: true,
		},
		{
			name:       "Stack with labelled version",
			stackTag:   "java-maven:latest",
			wantResult: true,
		},
		{
			name:       "Invalid stack name",
			stackTag:   "134stack",
			wantResult: false,
		},
		{
			name:       "Invalid version",
			stackTag:   "java-maven:.",
			wantResult: false,
		},
		{
			name:       "Malformed stack tag",
			stackTag:   "36g::test:",
			wantResult: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotResult, err := ValidateStackVersionTag(test.stackTag)
			if !test.wantErr && err != nil {
				t.Errorf("Unexpected err: %+v", err)
			} else if test.wantErr && err == nil {
				t.Error("Expected error but got nil")
			} else if !reflect.DeepEqual(gotResult, test.wantResult) {
				t.Errorf("Expected: %+v, \nGot: %+v", test.wantResult, gotResult)
			}
		})
	}
}

func TestSplitVersionFromStack(t *testing.T) {
	tests := []struct {
		name        string
		stackTag    string
		wantStack   string
		wantVersion string
		wantErr     bool
	}{
		{
			name:      "Stack without version",
			stackTag:  "java-maven",
			wantStack: "java-maven",
		},
		{
			name:        "Stack with specific version",
			stackTag:    "java-maven:1.0.0",
			wantStack:   "java-maven",
			wantVersion: "1.0.0",
		},
		{
			name:        "Stack with labelled version",
			stackTag:    "java-maven:latest",
			wantStack:   "java-maven",
			wantVersion: "latest",
		},
		{
			name:     "Invalid stack name",
			stackTag: "134stack",
			wantErr:  true,
		},
		{
			name:     "Invalid version",
			stackTag: "java-maven:.",
			wantErr:  true,
		},
		{
			name:     "Malformed stack tag",
			stackTag: "36g::test:",
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotStack, gotVersion, err := SplitVersionFromStack(test.stackTag)
			if !test.wantErr && err != nil {
				t.Errorf("Unexpected err: %+v", err)
			} else if test.wantErr && err == nil {
				t.Error("Expected error but got nil")
			} else if !reflect.DeepEqual(gotStack, test.wantStack) {
				t.Errorf("Expected Stack Name: %+v, \nGot Stack Name: %+v", test.wantStack, gotStack)
			} else if !reflect.DeepEqual(gotVersion, test.wantVersion) {
				t.Errorf("Expected Version: %+v, \nGot Version: %+v", test.wantVersion, gotVersion)
			}
		})
	}
}
