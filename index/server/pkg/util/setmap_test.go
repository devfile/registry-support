//
// Copyright Red Hat
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
	"reflect"
	"testing"
)

func TestStrArrayToSetMap(t *testing.T) {
	tests := []struct {
		name       string
		strArray   []string
		wantSetMap SetMap
	}{
		{
			name:     "convert single element string array into setmap",
			strArray: []string{"go"},
			wantSetMap: SetMap{
				"go": true,
			},
		},
		{
			name:     "convert multi-element string array into setmap",
			strArray: []string{"go", "python", "typescript"},
			wantSetMap: SetMap{
				"go":         true,
				"python":     true,
				"typescript": true,
			},
		},
		{
			name:       "empty string array into setmap",
			strArray:   []string{},
			wantSetMap: SetMap{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotSetMap := StrArrayToSetMap(test.strArray)
			if !reflect.DeepEqual(gotSetMap, test.wantSetMap) {
				t.Errorf("Got: %v, Expected: %v", gotSetMap, test.wantSetMap)
			}
		})
	}
}

func TestSetMapHas(t *testing.T) {
	tests := []struct {
		name   string
		setMap SetMap
		value  string
		want   bool
	}{
		{
			name: "value is in setmap",
			setMap: SetMap{
				"go":         true,
				"python":     true,
				"typescript": true,
			},
			value: "python",
			want:  true,
		},
		{
			name: "value is not in setmap",
			setMap: SetMap{
				"go":         true,
				"python":     true,
				"typescript": true,
			},
			value: "java",
			want:  false,
		},
		{
			name:   "empty setmap",
			setMap: SetMap{},
			value:  "python",
			want:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.setMap.SetMapHas(test.value)
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Got: %v, Expected: %v", got, test.want)
			}
		})
	}
}
