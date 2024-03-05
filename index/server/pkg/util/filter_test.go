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
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	indexSchema "github.com/devfile/registry-support/index/generator/schema"
	"github.com/mohae/deepcopy"
)

func TestFilterOut(t *testing.T) {
	tests := []struct {
		name        string
		index       []indexSchema.Schema
		filterAtIdx int
		wantIndex   []indexSchema.Schema
		expectPanic bool
	}{
		{
			name: "filter out index 2",
			index: []indexSchema.Schema{
				{
					Name:          "devfileA",
					Architectures: []string{"amd64", "arm64"},
				},
				{
					Name:          "devfileB",
					Architectures: []string{"amd64"},
				},
				{
					Name: "devfileC",
				},
			},
			filterAtIdx: 2,
			wantIndex: []indexSchema.Schema{
				{
					Name:          "devfileA",
					Architectures: []string{"amd64", "arm64"},
				},
				{
					Name:          "devfileB",
					Architectures: []string{"amd64"},
				},
			},
		},
		{
			name: "filter out index 1",
			index: []indexSchema.Schema{
				{
					Name:          "devfileA",
					Architectures: []string{"amd64", "arm64"},
				},
				{
					Name:          "devfileB",
					Architectures: []string{"amd64"},
				},
				{
					Name: "devfileC",
				},
			},
			filterAtIdx: 1,
			wantIndex: []indexSchema.Schema{
				{
					Name:          "devfileA",
					Architectures: []string{"amd64", "arm64"},
				},
				{
					Name: "devfileC",
				},
			},
		},
		{
			name: "filter out non-existent index 3",
			index: []indexSchema.Schema{
				{
					Name:          "devfileA",
					Architectures: []string{"amd64", "arm64"},
				},
				{
					Name:          "devfileB",
					Architectures: []string{"amd64"},
				},
				{
					Name: "devfileC",
				},
			},
			filterAtIdx: 3,
			wantIndex: []indexSchema.Schema{
				{
					Name:          "devfileA",
					Architectures: []string{"amd64", "arm64"},
				},
				{
					Name:          "devfileB",
					Architectures: []string{"amd64"},
				},
				{
					Name: "devfileC",
				},
			},
		},
		{
			name:        "filter out index 0 on empty array",
			index:       []indexSchema.Schema{},
			filterAtIdx: 0,
			wantIndex:   []indexSchema.Schema{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotIndex := deepcopy.Copy(test.index).([]indexSchema.Schema)
			filterOut(&gotIndex, &test.filterAtIdx)
			if !reflect.DeepEqual(gotIndex, test.wantIndex) {
				t.Errorf("Got: %v, Expected: %v", gotIndex, test.wantIndex)
			}
		})
	}
}

func TestFuzzyMatch(t *testing.T) {
	tests := []struct {
		name   string
		valueA string
		valueB string
		want   bool
	}{
		{
			name:   "Exact match",
			valueA: "Java Springboot",
			valueB: "Java Springboot",
			want:   true,
		},
		{
			name:   "Match with period",
			valueA: "Java Springboot",
			valueB: "Java Springboot.",
			want:   true,
		},
		{
			name:   "Match with question mark",
			valueA: "Java Springboot",
			valueB: "Java Springboot?",
			want:   true,
		},
		{
			name:   "Match with exclamation mark",
			valueA: "Java Springboot",
			valueB: "Java Springboot!",
			want:   true,
		},
		{
			name:   "Match using a dash",
			valueA: "Java Springboot",
			valueB: "java-springboot",
			want:   true,
		},
		{
			name:   "Extra space match",
			valueA: "Java Springboot",
			valueB: " Java  Springboot ",
			want:   true,
		},
		{
			name:   "Extra space with newlines match",
			valueA: "Java Springboot",
			valueB: `  Java
			Springboot  

			`,
			want: true,
		},
		{
			name:   "Partial match",
			valueA: "Java Springboot",
			valueB: "spring",
			want:   true,
		},
		{
			name:   "One word match",
			valueA: "Java Springboot",
			valueB: "java",
			want:   true,
		},
		{
			name:   "Mismatches",
			valueA: "Java Springboot",
			valueB: "python",
			want:   false,
		},
		{
			name:   "Mismatches with dash",
			valueA: "Java Springboot",
			valueB: "python-flask",
			want:   false,
		},
		{
			name:   "Mismatches with whitespace",
			valueA: "Java Springboot",
			valueB: `   python 
			flask `,
			want: false,
		},
		{
			name:   "Blank A",
			valueA: "",
			valueB: "python",
			want:   false,
		},
		{
			name:   "Blank B",
			valueA: "Java Springboot",
			valueB: "",
			want:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := fuzzyMatch(test.valueA, test.valueB)
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Got: %v, Expected: %v", got, test.want)
			}
		})
	}
}

func TestFilterDevfileArchitectures(t *testing.T) {

	tests := []struct {
		name      string
		index     []indexSchema.Schema
		archs     []string
		v1Index   bool
		wantIndex []indexSchema.Schema
	}{
		{
			name: "one arch filter",
			index: []indexSchema.Schema{
				{
					Name:          "devfileA",
					Architectures: []string{"amd64", "arm64"},
				},
				{
					Name:          "devfileB",
					Architectures: []string{"amd64"},
				},
				{
					Name: "devfileC",
				},
			},
			v1Index: true,
			archs:   []string{"amd64"},
			wantIndex: []indexSchema.Schema{
				{
					Name:          "devfileA",
					Architectures: []string{"amd64", "arm64"},
				},
				{
					Name:          "devfileB",
					Architectures: []string{"amd64"},
				},
				{
					Name: "devfileC",
				},
			},
		},
		{
			name: "one arch filter with v2 index",
			index: []indexSchema.Schema{
				{
					Name:          "devfileA",
					Architectures: []string{"amd64", "arm64"},
					Versions: []indexSchema.Version{
						{
							Version:       "1.0.0",
							Architectures: []string{"amd64"},
						},
						{
							Version:       "1.1.0",
							Architectures: []string{"arm64"},
						},
					},
				},
				{
					Name:          "devfileB",
					Architectures: []string{"amd64"},
				},
				{
					Name: "devfileC",
				},
			},
			v1Index: false,
			archs:   []string{"amd64"},
			wantIndex: []indexSchema.Schema{
				{
					Name:          "devfileA",
					Architectures: []string{"amd64", "arm64"},
					Versions: []indexSchema.Version{
						{
							Version:       "1.0.0",
							Architectures: []string{"amd64"},
						},
					},
				},
				{
					Name:          "devfileB",
					Architectures: []string{"amd64"},
				},
				{
					Name: "devfileC",
				},
			},
		},
		{
			name: "two arch filters",
			index: []indexSchema.Schema{
				{
					Name:          "devfileA",
					Architectures: []string{"amd64", "arm64"},
				},
				{
					Name:          "devfileB",
					Architectures: []string{"amd64"},
				},
				{
					Name: "devfileC",
				},
			},
			v1Index: true,
			archs:   []string{"amd64", "arm64"},
			wantIndex: []indexSchema.Schema{
				{
					Name:          "devfileA",
					Architectures: []string{"amd64", "arm64"},
				},
				{
					Name: "devfileC",
				},
			},
		},
		{
			name: "two arch filters with v2 index",
			index: []indexSchema.Schema{
				{
					Name:          "devfileA",
					Architectures: []string{"amd64", "arm64"},
					Versions: []indexSchema.Version{
						{
							Version:       "1.0.0",
							Architectures: []string{"amd64"},
						},
						{
							Version:       "1.1.0",
							Architectures: []string{"amd64", "arm64"},
						},
					},
				},
				{
					Name:          "devfileB",
					Architectures: []string{"amd64"},
				},
				{
					Name: "devfileC",
				},
			},
			v1Index: false,
			archs:   []string{"amd64", "arm64"},
			wantIndex: []indexSchema.Schema{
				{
					Name:          "devfileA",
					Architectures: []string{"amd64", "arm64"},
					Versions: []indexSchema.Version{
						{
							Version:       "1.1.0",
							Architectures: []string{"amd64", "arm64"},
						},
					},
				},
				{
					Name: "devfileC",
				},
			},
		},
		{
			name: "empty filters",
			index: []indexSchema.Schema{
				{
					Name:          "devfileA",
					Architectures: []string{"amd64", "arm64"},
				},
				{
					Name:          "devfileB",
					Architectures: []string{"amd64"},
				},
				{
					Name: "devfileC",
				},
			},
			v1Index: true,
			archs:   []string{},
			wantIndex: []indexSchema.Schema{
				{
					Name:          "devfileA",
					Architectures: []string{"amd64", "arm64"},
				},
				{
					Name:          "devfileB",
					Architectures: []string{"amd64"},
				},
				{
					Name: "devfileC",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotResult := filterDevfileArchitectures(test.index, test.archs, test.v1Index)
			gotResult.Eval()
			if !gotResult.IsEval {
				t.Errorf("Got unexpected unevaluated result: %v", gotResult)
			} else if gotResult.Error != nil {
				t.Errorf("Unexpected error: %v", gotResult.Error)
			} else if !reflect.DeepEqual(gotResult.Index, test.wantIndex) {
				t.Errorf("Got: %v, Expected: %v", gotResult.Index, test.wantIndex)
			}
		})
	}
}

func TestFilterDevfileTags(t *testing.T) {

	tests := []struct {
		name      string
		index     []indexSchema.Schema
		tags      []string
		v1Index   bool
		wantIndex []indexSchema.Schema
	}{
		{
			name: "one tag filter",
			index: []indexSchema.Schema{
				{
					Name: "devfileA",
					Tags: []string{"Python", "Django"},
				},
				{
					Name: "devfileB",
					Tags: []string{"Python"},
				},
				{
					Name: "devfileC",
				},
			},
			v1Index: true,
			tags:    []string{"Django"},
			wantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
					Tags: []string{"Python", "Django"},
				},
			},
		},
		{
			name: "one tag filter with v2 index",
			index: []indexSchema.Schema{
				{
					Name: "devfileA",
					Tags: []string{"Python", "Django", "Flask"},
					Versions: []indexSchema.Version{
						{
							Version: "1.0.0",
							Tags:    []string{"Python"},
						},
						{
							Version: "1.1.0",
							Tags:    []string{"Python", "Django"},
						},
						{
							Version: "2.0.0",
							Tags:    []string{"Python", "Flask"},
						},
					},
				},
				{
					Name: "devfileB",
					Tags: []string{"Python"},
				},
				{
					Name: "devfileC",
				},
			},
			v1Index: false,
			tags:    []string{"Django"},
			wantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
					Tags: []string{"Python", "Django", "Flask"},
					Versions: []indexSchema.Version{
						{
							Version: "1.1.0",
							Tags:    []string{"Python", "Django"},
						},
					},
				},
			},
		},
		{
			name: "two tag filters",
			index: []indexSchema.Schema{
				{
					Name: "devfileA",
					Tags: []string{"Python", "Django"},
				},
				{
					Name: "devfileB",
					Tags: []string{"Python"},
				},
				{
					Name: "devfileC",
				},
			},
			v1Index: true,
			tags:    []string{"Python", "Django"},
			wantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
					Tags: []string{"Python", "Django"},
				},
			},
		},
		{
			name: "two tag filters with v2 index",
			index: []indexSchema.Schema{
				{
					Name: "devfileA",
					Tags: []string{"Python", "Django", "Flask"},
					Versions: []indexSchema.Version{
						{
							Version: "1.0.0",
							Tags:    []string{"Python"},
						},
						{
							Version: "1.1.0",
							Tags:    []string{"Python", "Django"},
						},
						{
							Version: "2.0.0",
							Tags:    []string{"Python", "Flask"},
						},
					},
				},
				{
					Name: "devfileB",
					Tags: []string{"Python"},
				},
				{
					Name: "devfileC",
				},
			},
			v1Index: false,
			tags:    []string{"Python", "Django"},
			wantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
					Tags: []string{"Python", "Django", "Flask"},
					Versions: []indexSchema.Version{
						{
							Version: "1.1.0",
							Tags:    []string{"Python", "Django"},
						},
					},
				},
			},
		},
		{
			name: "empty filters",
			index: []indexSchema.Schema{
				{
					Name: "devfileA",
					Tags: []string{"Python", "Django"},
				},
				{
					Name: "devfileB",
					Tags: []string{"Python"},
				},
				{
					Name: "devfileC",
				},
			},
			v1Index: true,
			tags:    []string{},
			wantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
					Tags: []string{"Python", "Django"},
				},
				{
					Name: "devfileB",
					Tags: []string{"Python"},
				},
				{
					Name: "devfileC",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotResult := filterDevfileTags(test.index, test.tags, test.v1Index)
			gotResult.Eval()
			if !gotResult.IsEval {
				t.Errorf("Got unexpected unevaluated result: %v", gotResult)
			} else if gotResult.Error != nil {
				t.Errorf("Unexpected error: %v", gotResult.Error)
			} else if !reflect.DeepEqual(gotResult.Index, test.wantIndex) {
				t.Errorf("Got: %v, Expected: %v", gotResult.Index, test.wantIndex)
			}
		})
	}
}

func TestFilterDevfileSchemaVersion(t *testing.T) {

	tests := []struct {
		name             string
		index            []indexSchema.Schema
		minSchemaVersion string
		maxSchemaVersion string
		wantIndex        []indexSchema.Schema
	}{
		{
			name: "only minSchemaVersion",
			index: []indexSchema.Schema{
				{
					Name: "devfileA",
					Versions: []indexSchema.Version{
						{
							Version:       "1.0.0",
							SchemaVersion: "2.0.0",
						},
						{
							Version:       "1.1.0",
							SchemaVersion: "2.1.0",
						},
						{
							Version:       "1.2.0",
							SchemaVersion: "2.2.0",
						},
					},
				},
				{
					Name: "devfileB",
					Versions: []indexSchema.Version{
						{
							Version:       "1.0.0",
							SchemaVersion: "2.0.0",
						},
					},
				},
			},
			minSchemaVersion: "2.1",
			wantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
					Versions: []indexSchema.Version{
						{
							Version:       "1.1.0",
							SchemaVersion: "2.1.0",
						},
						{
							Version:       "1.2.0",
							SchemaVersion: "2.2.0",
						},
					},
				},
			},
		},
		{
			name: "only maxSchemaVersion",
			index: []indexSchema.Schema{
				{
					Name: "devfileA",
					Versions: []indexSchema.Version{
						{
							Version:       "1.0.0",
							SchemaVersion: "2.0.0",
						},
						{
							Version:       "1.1.0",
							SchemaVersion: "2.1.0",
						},
						{
							Version:       "1.2.0",
							SchemaVersion: "2.2.0",
						},
					},
				},
				{
					Name: "devfileB",
					Versions: []indexSchema.Version{
						{
							Version:       "1.1.0",
							SchemaVersion: "2.1.0",
						},
					},
				},
			},
			maxSchemaVersion: "2.1",
			wantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
					Versions: []indexSchema.Version{
						{
							Version:       "1.0.0",
							SchemaVersion: "2.0.0",
						},
						{
							Version:       "1.1.0",
							SchemaVersion: "2.1.0",
						},
					},
				},
				{
					Name: "devfileB",
					Versions: []indexSchema.Version{
						{
							Version:       "1.1.0",
							SchemaVersion: "2.1.0",
						},
					},
				},
			},
		},
		{
			name: "both minSchemaVersion and maxSchemaVersion",
			index: []indexSchema.Schema{
				{
					Name: "devfileA",
					Versions: []indexSchema.Version{
						{
							Version:       "1.0.0",
							SchemaVersion: "2.0.0",
						},
						{
							Version:       "1.1.0",
							SchemaVersion: "2.1.0",
						},
						{
							Version:       "1.2.0",
							SchemaVersion: "2.2.0",
						},
					},
				},
				{
					Name: "devfileB",
					Versions: []indexSchema.Version{
						{
							Version:       "1.1.0",
							SchemaVersion: "2.1.0",
						},
					},
				},
			},
			minSchemaVersion: "2.1",
			maxSchemaVersion: "2.1",
			wantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
					Versions: []indexSchema.Version{
						{
							Version:       "1.1.0",
							SchemaVersion: "2.1.0",
						},
					},
				},
				{
					Name: "devfileB",
					Versions: []indexSchema.Version{
						{
							Version:       "1.1.0",
							SchemaVersion: "2.1.0",
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotResult := FilterDevfileSchemaVersion(test.index, test.minSchemaVersion, test.maxSchemaVersion)
			gotResult.Eval()
			if !gotResult.IsEval {
				t.Errorf("Got unexpected unevaluated result: %v", gotResult)
			} else if gotResult.Error != nil {
				t.Errorf("Unexpected error: %v", gotResult.Error)
			} else if !reflect.DeepEqual(gotResult.Index, test.wantIndex) {
				t.Errorf("Got: %v, Expected: %v", gotResult.Index, test.wantIndex)
			}
		})
	}
}

func TestFilterDevfileStrArrayField(t *testing.T) {
	tests := []struct {
		name      string
		index     []indexSchema.Schema
		fieldName string
		values    []string
		v1Index   bool
		wantIndex []indexSchema.Schema
	}{
		{
			name:      "two arch filters",
			fieldName: ARRAY_PARAM_ARCHITECTURES,
			index: []indexSchema.Schema{
				{
					Name:          "devfileA",
					Architectures: []string{"amd64", "arm64"},
				},
				{
					Name:          "devfileB",
					Architectures: []string{"amd64"},
				},
				{
					Name: "devfileC",
				},
			},
			v1Index: true,
			values:  []string{"amd64", "arm64"},
			wantIndex: []indexSchema.Schema{
				{
					Name:          "devfileA",
					Architectures: []string{"amd64", "arm64"},
				},
				{
					Name: "devfileC",
				},
			},
		},
		{
			name:      "two arch filters with v2 index",
			fieldName: ARRAY_PARAM_ARCHITECTURES,
			index: []indexSchema.Schema{
				{
					Name:          "devfileA",
					Architectures: []string{"amd64", "arm64"},
					Versions: []indexSchema.Version{
						{
							Version:       "1.0.0",
							Architectures: []string{"amd64"},
						},
						{
							Version:       "1.1.0",
							Architectures: []string{"amd64", "arm64"},
						},
					},
				},
				{
					Name:          "devfileB",
					Architectures: []string{"amd64"},
				},
				{
					Name: "devfileC",
				},
			},
			v1Index: false,
			values:  []string{"amd64", "arm64"},
			wantIndex: []indexSchema.Schema{
				{
					Name:          "devfileA",
					Architectures: []string{"amd64", "arm64"},
					Versions: []indexSchema.Version{
						{
							Version:       "1.1.0",
							Architectures: []string{"amd64", "arm64"},
						},
					},
				},
				{
					Name: "devfileC",
				},
			},
		},
		{
			name:      "two tag filters",
			fieldName: ARRAY_PARAM_TAGS,
			index: []indexSchema.Schema{
				{
					Name: "devfileA",
					Tags: []string{"Python", "Django"},
				},
				{
					Name: "devfileB",
					Tags: []string{"Python"},
				},
				{
					Name: "devfileC",
				},
			},
			v1Index: true,
			values:  []string{"Python", "Django"},
			wantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
					Tags: []string{"Python", "Django"},
				},
			},
		},
		{
			name:      "two tag filters with v2 index",
			fieldName: ARRAY_PARAM_TAGS,
			index: []indexSchema.Schema{
				{
					Name: "devfileA",
					Tags: []string{"Python", "Django", "Flask"},
					Versions: []indexSchema.Version{
						{
							Version: "1.0.0",
							Tags:    []string{"Python"},
						},
						{
							Version: "1.1.0",
							Tags:    []string{"Python", "Django"},
						},
						{
							Version: "2.0.0",
							Tags:    []string{"Python", "Flask"},
						},
					},
				},
				{
					Name: "devfileB",
					Tags: []string{"Python"},
				},
				{
					Name: "devfileC",
				},
			},
			v1Index: false,
			values:  []string{"Python", "Django"},
			wantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
					Tags: []string{"Python", "Django", "Flask"},
					Versions: []indexSchema.Version{
						{
							Version: "1.1.0",
							Tags:    []string{"Python", "Django"},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotResult := FilterDevfileStrArrayField(test.index, test.fieldName, test.values, test.v1Index)
			gotResult.Eval()
			if !gotResult.IsEval {
				t.Errorf("Got unexpected unevaluated result: %v", gotResult)
			} else if gotResult.Error != nil {
				t.Errorf("Unexpected error: %v", gotResult.Error)
			} else if !reflect.DeepEqual(gotResult.Index, test.wantIndex) {
				t.Errorf("Got: %v, Expected: %v", gotResult.Index, test.wantIndex)
			}
		})
	}
}

func TestFilterDevfileStrField(t *testing.T) {
	tests := []struct {
		name       string
		index      []indexSchema.Schema
		fieldName  string
		value      string
		v1Index    bool
		wantIndex  []indexSchema.Schema
		wantErr    bool
		wantErrStr string
	}{
		{
			name:      "name filter",
			fieldName: PARAM_NAME,
			index: []indexSchema.Schema{
				{
					Name: "devfileA",
				},
				{
					Name: "devfileB",
				},
				{
					Name: "devfileC",
				},
				{
					Name: "devfileAA",
				},
			},
			v1Index: true,
			value:   "A",
			wantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
				},
				{
					Name: "devfileAA",
				},
			},
			wantErr: false,
		},
		{
			name:      "name filter v2",
			fieldName: PARAM_NAME,
			index: []indexSchema.Schema{
				{
					Name: "devfileA",
				},
				{
					Name: "devfileB",
				},
				{
					Name: "devfileC",
				},
				{
					Name: "devfileAA",
				},
			},
			value: "A",
			wantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
				},
				{
					Name: "devfileAA",
				},
			},
			wantErr: false,
		},
		{
			name:      "display name filter",
			fieldName: PARAM_DISPLAY_NAME,
			index: []indexSchema.Schema{
				{
					Name:        "devfileA",
					DisplayName: "Python",
				},
				{
					Name:        "devfileB",
					DisplayName: "Python",
				},
				{
					Name:        "devfileC",
					DisplayName: "Flask",
				},
				{
					Name:        "devfileAA",
					DisplayName: "Python - Flask",
				},
			},
			v1Index: true,
			value:   "Flask",
			wantIndex: []indexSchema.Schema{
				{
					Name:        "devfileC",
					DisplayName: "Flask",
				},
				{
					Name:        "devfileAA",
					DisplayName: "Python - Flask",
				},
			},
			wantErr: false,
		},
		{
			name:      "display name filter v2",
			fieldName: PARAM_DISPLAY_NAME,
			index: []indexSchema.Schema{
				{
					Name:        "devfileA",
					DisplayName: "Python",
				},
				{
					Name:        "devfileB",
					DisplayName: "Python",
				},
				{
					Name:        "devfileC",
					DisplayName: "Flask",
				},
				{
					Name:        "devfileAA",
					DisplayName: "Python - Flask",
				},
			},
			value: "Flask",
			wantIndex: []indexSchema.Schema{
				{
					Name:        "devfileC",
					DisplayName: "Flask",
				},
				{
					Name:        "devfileAA",
					DisplayName: "Python - Flask",
				},
			},
			wantErr: false,
		},
		{
			name:      "description filter",
			fieldName: PARAM_DESCRIPTION,
			index: []indexSchema.Schema{
				{
					Name:        "devfileA",
					DisplayName: "Python",
					Description: "A python stack.",
				},
				{
					Name:        "devfileB",
					DisplayName: "Python",
					Description: "A python sample.",
				},
				{
					Name:        "devfileC",
					DisplayName: "Flask",
					Description: "A python flask stack.",
				},
				{
					Name:        "devfileAA",
					DisplayName: "Python - Flask",
					Description: "A python flask sample.",
				},
			},
			v1Index: true,
			value:   "stack",
			wantIndex: []indexSchema.Schema{
				{
					Name:        "devfileA",
					DisplayName: "Python",
					Description: "A python stack.",
				},
				{
					Name:        "devfileC",
					DisplayName: "Flask",
					Description: "A python flask stack.",
				},
			},
			wantErr: false,
		},
		{
			name:      "description filter v2",
			fieldName: PARAM_DESCRIPTION,
			index: []indexSchema.Schema{
				{
					Name:        "devfileA",
					DisplayName: "Python",
					Description: "A python stack.",
					Versions: []indexSchema.Version{
						{
							Version:     "1.0.0",
							Description: "A python stack.",
							Default:     true,
						},
					},
				},
				{
					Name:        "devfileB",
					DisplayName: "Python",
					Description: "A python sample.",
					Versions: []indexSchema.Version{
						{
							Version:     "1.0.0",
							Description: "A python sample.",
						},
						{
							Version:     "2.0.0",
							Description: "A python stack.",
							Default:     true,
						},
					},
				},
				{
					Name:        "devfileC",
					DisplayName: "Flask",
					Description: "A python flask stack.",
				},
				{
					Name:        "devfileAA",
					DisplayName: "Python - Flask",
					Description: "A python flask sample.",
				},
			},
			value: "stack",
			wantIndex: []indexSchema.Schema{
				{
					Name:        "devfileA",
					DisplayName: "Python",
					Description: "A python stack.",
					Versions: []indexSchema.Version{
						{
							Version:     "1.0.0",
							Description: "A python stack.",
							Default:     true,
						},
					},
				},
				{
					Name:        "devfileB",
					DisplayName: "Python",
					Description: "A python sample.",
					Versions: []indexSchema.Version{
						{
							Version:     "2.0.0",
							Description: "A python stack.",
							Default:     true,
						},
					},
				},
				{
					Name:        "devfileC",
					DisplayName: "Flask",
					Description: "A python flask stack.",
				},
			},
			wantErr: false,
		},
		{
			name:      "icon filter",
			fieldName: PARAM_ICON,
			index: []indexSchema.Schema{
				{
					Name:        "devfileA",
					DisplayName: "Python",
					Icon:        "devfileA.png",
				},
				{
					Name:        "devfileB",
					DisplayName: "Python",
					Icon:        "devfileB.png",
				},
				{
					Name:        "devfileC",
					DisplayName: "Flask",
					Icon:        "devfileC.jpg",
				},
				{
					Name:        "devfileAA",
					DisplayName: "Python - Flask",
					Icon:        "devfileAA.ico",
				},
			},
			v1Index: true,
			value:   "png",
			wantIndex: []indexSchema.Schema{
				{
					Name:        "devfileA",
					DisplayName: "Python",
					Icon:        "devfileA.png",
				},
				{
					Name:        "devfileB",
					DisplayName: "Python",
					Icon:        "devfileB.png",
				},
			},
			wantErr: false,
		},
		{
			name:      "icon filter v2",
			fieldName: PARAM_ICON,
			index: []indexSchema.Schema{
				{
					Name:        "devfileA",
					DisplayName: "Python",
					Description: "A python stack.",
					Versions: []indexSchema.Version{
						{
							Version: "1.0.0",
							Icon:    "devfileA.png",
							Default: true,
						},
					},
				},
				{
					Name:        "devfileB",
					DisplayName: "Python",
					Description: "A python sample.",
					Versions: []indexSchema.Version{
						{
							Version: "1.0.0",
							Icon:    "devfileB.png",
						},
						{
							Version: "2.0.0",
							Icon:    "devfileB.ico",
							Default: true,
						},
					},
				},
				{
					Name:        "devfileC",
					DisplayName: "Flask",
					Icon:        "devfileC.jpg",
				},
				{
					Name:        "devfileAA",
					DisplayName: "Python - Flask",
					Icon:        "devfileAA.ico",
				},
			},
			value: "ico",
			wantIndex: []indexSchema.Schema{
				{
					Name:        "devfileB",
					DisplayName: "Python",
					Description: "A python sample.",
					Versions: []indexSchema.Version{
						{
							Version: "2.0.0",
							Icon:    "devfileB.ico",
							Default: true,
						},
					},
				},
				{
					Name:        "devfileAA",
					DisplayName: "Python - Flask",
					Icon:        "devfileAA.ico",
				},
			},
			wantErr: false,
		},
		{
			name:      "project type filter",
			fieldName: PARAM_PROJECT_TYPE,
			index: []indexSchema.Schema{
				{
					Name:        "devfileA",
					DisplayName: "Python",
					ProjectType: "python",
				},
				{
					Name:        "devfileB",
					DisplayName: "Python",
					ProjectType: "python",
				},
				{
					Name:        "devfileC",
					DisplayName: "Flask",
					ProjectType: "python",
				},
				{
					Name:        "devfileD",
					DisplayName: "Java Springboot",
					ProjectType: "java",
				},
				{
					Name:        "devfileAA",
					DisplayName: "Python - Flask",
					ProjectType: "python",
				},
			},
			v1Index: true,
			value:   "java",
			wantIndex: []indexSchema.Schema{
				{
					Name:        "devfileD",
					DisplayName: "Java Springboot",
					ProjectType: "java",
				},
			},
			wantErr: false,
		},
		{
			name:      "project type filter v2",
			fieldName: PARAM_PROJECT_TYPE,
			index: []indexSchema.Schema{
				{
					Name:        "devfileA",
					DisplayName: "Python",
					ProjectType: "python",
				},
				{
					Name:        "devfileB",
					DisplayName: "Python",
					ProjectType: "python",
				},
				{
					Name:        "devfileC",
					DisplayName: "Flask",
					ProjectType: "python",
				},
				{
					Name:        "devfileD",
					DisplayName: "Java Springboot",
					ProjectType: "java",
				},
				{
					Name:        "devfileAA",
					DisplayName: "Python - Flask",
					ProjectType: "python",
				},
			},
			value: "java",
			wantIndex: []indexSchema.Schema{
				{
					Name:        "devfileD",
					DisplayName: "Java Springboot",
					ProjectType: "java",
				},
			},
			wantErr: false,
		},
		{
			name:      "language filter",
			fieldName: PARAM_LANGUAGE,
			index: []indexSchema.Schema{
				{
					Name:        "devfileA",
					DisplayName: "Python",
					Language:    "Python",
				},
				{
					Name:        "devfileB",
					DisplayName: "Python",
					Language:    "Python",
				},
				{
					Name:        "devfileC",
					DisplayName: "Flask",
					Language:    "Python",
				},
				{
					Name:        "devfileD",
					DisplayName: "Java Springboot",
					Language:    "Java",
				},
				{
					Name:        "devfileAA",
					DisplayName: "Python - Flask",
					Language:    "Python",
				},
			},
			v1Index: true,
			value:   "java",
			wantIndex: []indexSchema.Schema{
				{
					Name:        "devfileD",
					DisplayName: "Java Springboot",
					Language:    "Java",
				},
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotResult := FilterDevfileStrField(test.index, test.fieldName, test.value, test.v1Index)
			gotResult.Eval()
			if !gotResult.IsEval {
				t.Errorf("Got unexpected unevaluated result: %v", gotResult)
			} else if !test.wantErr && gotResult.Error != nil {
				t.Errorf("Unexpected error: %v", gotResult.Error)
			} else if !test.wantErr && !reflect.DeepEqual(gotResult.Index, test.wantIndex) {
				t.Errorf("Got: %v, Expected: %v", gotResult.Index, test.wantIndex)
			} else if test.wantErr && !strings.HasPrefix(gotResult.Error.Error(), test.wantErrStr) {
				t.Errorf("Got: %v, Expected: %v", gotResult.Error.Error(), test.wantErrStr)
			}
		})
	}
}

func TestAndFilter(t *testing.T) {
	tests := []struct {
		name        string
		filters     []FilterResult
		wantIndex   []indexSchema.Schema
		wantNotEval bool
		wantErr     bool
		wantErrStr  string
	}{
		{
			name: "Test Valid And with same results",
			filters: []FilterResult{
				{
					Index: []indexSchema.Schema{
						{
							Name:        "go",
							DisplayName: "Go",
							Description: "Go Gin Project",
						},
						{
							Name:        "python",
							DisplayName: "Python Flask",
							Description: "Python Flask project",
						},
					},
					IsEval: true,
				},
				{
					Index: []indexSchema.Schema{
						{
							Name:        "go",
							DisplayName: "Go",
							Description: "Go Gin Project",
						},
						{
							Name:        "python",
							DisplayName: "Python Flask",
							Description: "Python Flask project",
						},
					},
					IsEval: true,
				},
			},
			wantIndex: []indexSchema.Schema{
				{
					Name:        "go",
					DisplayName: "Go",
					Description: "Go Gin Project",
				},
				{
					Name:        "python",
					DisplayName: "Python Flask",
					Description: "Python Flask project",
				},
			},
		},
		{
			name: "Test Valid And with same results v2",
			filters: []FilterResult{
				{
					Index: []indexSchema.Schema{
						{
							Name:        "go",
							DisplayName: "Go",
							Description: "Go Gin Project",
							Versions: []indexSchema.Version{
								{
									Version:       "1.0.0",
									SchemaVersion: "2.2.0",
									Description:   "Go Gin Project",
									Tags: []string{
										"Go",
										"Gin",
									},
								},
							},
						},
						{
							Name:        "python",
							DisplayName: "Python Flask",
							Description: "Python Flask project",
							Versions: []indexSchema.Version{
								{
									Version:       "1.0.0",
									SchemaVersion: "2.2.0",
									Description:   "Python Flask project",
									Tags: []string{
										"Python",
										"Flask",
									},
								},
								{
									Version:       "1.1.0",
									SchemaVersion: "2.2.0",
									Description:   "Python Flask backend-web project",
									Tags: []string{
										"Python",
										"Flask",
									},
								},
							},
						},
					},
					IsEval: true,
				},
				{
					Index: []indexSchema.Schema{
						{
							Name:        "go",
							DisplayName: "Go",
							Description: "Go Gin Project",
							Versions: []indexSchema.Version{
								{
									Version:       "1.0.0",
									SchemaVersion: "2.2.0",
									Description:   "Go Gin Project",
									Tags: []string{
										"Go",
										"Gin",
									},
								},
							},
						},
						{
							Name:        "python",
							DisplayName: "Python Flask",
							Description: "Python Flask project",
							Versions: []indexSchema.Version{
								{
									Version:       "1.0.0",
									SchemaVersion: "2.2.0",
									Description:   "Python Flask project",
									Tags: []string{
										"Python",
										"Flask",
									},
								},
								{
									Version:       "1.1.0",
									SchemaVersion: "2.2.0",
									Description:   "Python Flask backend-web project",
									Tags: []string{
										"Python",
										"Flask",
									},
								},
							},
						},
					},
					IsEval: true,
				},
			},
			wantIndex: []indexSchema.Schema{
				{
					Name:        "go",
					DisplayName: "Go",
					Description: "Go Gin Project",
					Versions: []indexSchema.Version{
						{
							Version:       "1.0.0",
							SchemaVersion: "2.2.0",
							Description:   "Go Gin Project",
							Tags: []string{
								"Go",
								"Gin",
							},
						},
					},
				},
				{
					Name:        "python",
					DisplayName: "Python Flask",
					Description: "Python Flask project",
					Versions: []indexSchema.Version{
						{
							Version:       "1.0.0",
							SchemaVersion: "2.2.0",
							Description:   "Python Flask project",
							Tags: []string{
								"Python",
								"Flask",
							},
						},
						{
							Version:       "1.1.0",
							SchemaVersion: "2.2.0",
							Description:   "Python Flask backend-web project",
							Tags: []string{
								"Python",
								"Flask",
							},
						},
					},
				},
			},
		},
		{
			name: "Test Valid And with overlapping results",
			filters: []FilterResult{
				{
					Index: []indexSchema.Schema{
						{
							Name:        "go",
							DisplayName: "Go",
							Description: "Go Gin Project",
						},
						{
							Name:        "python",
							DisplayName: "Python Flask",
							Description: "Python Flask project",
						},
					},
					IsEval: true,
				},
				{
					Index: []indexSchema.Schema{
						{
							Name:        "go",
							DisplayName: "Go",
							Description: "Go Gin Project",
						},
					},
					IsEval: true,
				},
			},
			wantIndex: []indexSchema.Schema{
				{
					Name:        "go",
					DisplayName: "Go",
					Description: "Go Gin Project",
				},
			},
		},
		{
			name: "Test Valid And with overlapping results v2",
			filters: []FilterResult{
				{
					Index: []indexSchema.Schema{
						{
							Name:        "go",
							DisplayName: "Go",
							Description: "Go Gin Project",
							Versions: []indexSchema.Version{
								{
									Version:       "1.0.0",
									SchemaVersion: "2.2.0",
									Description:   "Go Gin Project",
									Tags: []string{
										"Go",
										"Gin",
									},
								},
							},
						},
					},
					IsEval: true,
				},
				{
					Index: []indexSchema.Schema{
						{
							Name:        "go",
							DisplayName: "Go",
							Description: "Go Gin Project",
							Versions: []indexSchema.Version{
								{
									Version:       "1.0.0",
									SchemaVersion: "2.2.0",
									Description:   "Go Gin Project",
									Tags: []string{
										"Go",
										"Gin",
									},
								},
							},
						},
						{
							Name:        "python",
							DisplayName: "Python Flask",
							Description: "Python Flask project",
							Versions: []indexSchema.Version{
								{
									Version:       "1.0.0",
									SchemaVersion: "2.2.0",
									Description:   "Python Flask project",
									Tags: []string{
										"Python",
										"Flask",
									},
								},
								{
									Version:       "1.1.0",
									SchemaVersion: "2.2.0",
									Description:   "Python Flask backend-web project",
									Tags: []string{
										"Python",
										"Flask",
									},
								},
							},
						},
					},
					IsEval: true,
				},
			},
			wantIndex: []indexSchema.Schema{
				{
					Name:        "go",
					DisplayName: "Go",
					Description: "Go Gin Project",
					Versions: []indexSchema.Version{
						{
							Version:       "1.0.0",
							SchemaVersion: "2.2.0",
							Description:   "Go Gin Project",
							Tags: []string{
								"Go",
								"Gin",
							},
						},
					},
				},
			},
		},
		{
			name: "Test Valid And with overlapping results and versions v2",
			filters: []FilterResult{
				{
					Index: []indexSchema.Schema{
						{
							Name:        "go",
							DisplayName: "Go",
							Description: "Go Gin Project",
							Versions: []indexSchema.Version{
								{
									Version:       "1.0.0",
									SchemaVersion: "2.2.0",
									Description:   "Go Gin Project",
									Tags: []string{
										"Go",
										"Gin",
									},
								},
							},
						},
					},
					IsEval: true,
				},
				{
					Index: []indexSchema.Schema{
						{
							Name:        "go",
							DisplayName: "Go",
							Description: "Go Gin Project",
							Versions: []indexSchema.Version{
								{
									Version:       "1.0.0",
									SchemaVersion: "2.2.0",
									Description:   "Go Gin Project",
									Tags: []string{
										"Go",
										"Gin",
									},
								},
								{
									Version:       "2.0.0",
									SchemaVersion: "2.2.0",
									Description:   "Go Gin Backend Project",
									Tags: []string{
										"Go",
										"Gin",
										"MySQL",
									},
								},
							},
						},
						{
							Name:        "python",
							DisplayName: "Python Flask",
							Description: "Python Flask project",
							Versions: []indexSchema.Version{
								{
									Version:       "1.0.0",
									SchemaVersion: "2.2.0",
									Description:   "Python Flask project",
									Tags: []string{
										"Python",
										"Flask",
									},
								},
								{
									Version:       "1.1.0",
									SchemaVersion: "2.2.0",
									Description:   "Python Flask backend-web project",
									Tags: []string{
										"Python",
										"Flask",
									},
								},
							},
						},
					},
					IsEval: true,
				},
			},
			wantIndex: []indexSchema.Schema{
				{
					Name:        "go",
					DisplayName: "Go",
					Description: "Go Gin Project",
					Versions: []indexSchema.Version{
						{
							Version:       "1.0.0",
							SchemaVersion: "2.2.0",
							Description:   "Go Gin Project",
							Tags: []string{
								"Go",
								"Gin",
							},
						},
					},
				},
			},
		},
		{
			name:      "Test Valid And with no results",
			filters:   []FilterResult{},
			wantIndex: []indexSchema.Schema{},
		},
		{
			name: "Test Invalid And with single error",
			filters: []FilterResult{
				{
					Error:  fmt.Errorf("A test error"),
					IsEval: true,
				},
			},
			wantIndex:  []indexSchema.Schema{},
			wantErr:    true,
			wantErrStr: "A test error",
		},
		{
			name: "Test Invalid And with multiple errors",
			filters: []FilterResult{
				{
					Error:  fmt.Errorf("First test error"),
					IsEval: true,
				},
				{
					Error:  fmt.Errorf("Second test error"),
					IsEval: true,
				},
				{
					Error:  fmt.Errorf("Third test error"),
					IsEval: true,
				},
			},
			wantIndex:  []indexSchema.Schema{},
			wantErr:    true,
			wantErrStr: "First test error",
		},
		{
			name: "Test Invalid And with valid filters and errors",
			filters: []FilterResult{
				{
					Index: []indexSchema.Schema{
						{
							Name:        "go",
							DisplayName: "Go",
							Description: "Go Gin Project",
						},
						{
							Name:        "python",
							DisplayName: "Python Flask",
							Description: "Python Flask project",
						},
					},
					IsEval: true,
				},
				{
					Error:  fmt.Errorf("First test error"),
					IsEval: true,
				},
				{
					Index: []indexSchema.Schema{
						{
							Name:        "go",
							DisplayName: "Go",
							Description: "Go Gin Project",
						},
					},
					IsEval: true,
				},
				{
					Error:  fmt.Errorf("Second test error"),
					IsEval: true,
				},
				{
					Error:  fmt.Errorf("Third test error"),
					IsEval: true,
				},
			},
			wantIndex:  []indexSchema.Schema{},
			wantErr:    true,
			wantErrStr: "First test error",
		},
		{
			name: "Test Invalid And with valid filters and errors v2",
			filters: []FilterResult{
				{
					Index: []indexSchema.Schema{
						{
							Name:        "go",
							DisplayName: "Go",
							Description: "Go Gin Project",
							Versions: []indexSchema.Version{
								{
									Version:       "1.0.0",
									SchemaVersion: "2.2.0",
									Description:   "Go Gin Project",
									Tags: []string{
										"Go",
										"Gin",
									},
								},
								{
									Version:       "2.0.0",
									SchemaVersion: "2.2.0",
									Description:   "Go Gin Backend Project",
									Tags: []string{
										"Go",
										"Gin",
										"MySQL",
									},
								},
							},
						},
						{
							Name:        "python",
							DisplayName: "Python Flask",
							Description: "Python Flask project",
							Versions: []indexSchema.Version{
								{
									Version:       "1.0.0",
									SchemaVersion: "2.2.0",
									Description:   "Python Flask project",
									Tags: []string{
										"Python",
										"Flask",
									},
								},
								{
									Version:       "1.1.0",
									SchemaVersion: "2.2.0",
									Description:   "Python Flask backend-web project",
									Tags: []string{
										"Python",
										"Flask",
									},
								},
							},
						},
					},
					IsEval: true,
				},
				{
					Error:  fmt.Errorf("First test error"),
					IsEval: true,
				},
				{
					Index: []indexSchema.Schema{
						{
							Name:        "go",
							DisplayName: "Go",
							Description: "Go Gin Project",
							Versions: []indexSchema.Version{
								{
									Version:       "1.0.0",
									SchemaVersion: "2.2.0",
									Description:   "Go Gin Project",
									Tags: []string{
										"Go",
										"Gin",
									},
								},
							},
						},
					},
					IsEval: true,
				},
				{
					Error:  fmt.Errorf("Second test error"),
					IsEval: true,
				},
				{
					Error:  fmt.Errorf("Third test error"),
					IsEval: true,
				},
			},
			wantIndex:  []indexSchema.Schema{},
			wantErr:    true,
			wantErrStr: "First test error",
		},
		{
			name: "Test Unevaluated FilterResult entities with And filter",
			filters: []FilterResult{
				{
					filterFn: func(fr *FilterResult) {
						newIndex := []indexSchema.Schema{}
						for _, schema := range fr.Index {
							if schema.Name != "python" {
								newIndex = append(newIndex, schema)
							}
						}
						fr.Index = newIndex
					},
					Index: []indexSchema.Schema{
						{
							Name:        "go",
							DisplayName: "Go",
							Description: "Go Gin Project",
						},
						{
							Name:        "python",
							DisplayName: "Python Flask",
							Description: "Python Flask project",
						},
					},
				},
			},
			wantIndex: []indexSchema.Schema{
				{
					Name:        "go",
					DisplayName: "Go",
					Description: "Go Gin Project",
				},
			},
			wantNotEval: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Fail if a child filter is not evaluated and test case does not expect this
			for _, filter := range test.filters {
				if !test.wantNotEval && !filter.IsEval {
					t.Errorf("Got unexpected unevaluated result: %v", filter)
					return
				}
			}

			gotResult := AndFilter(test.filters...)
			gotResult.Eval()

			if !gotResult.IsEval {
				t.Errorf("Got unexpected unevaluated result: %v", gotResult)
				return
			}

			// sort result and expected as sorting of result is not consistent
			sort.SliceStable(gotResult.Index, func(i, j int) bool {
				return gotResult.Index[i].Name < gotResult.Index[j].Name
			})
			sort.SliceStable(test.wantIndex, func(i, j int) bool {
				return test.wantIndex[i].Name < test.wantIndex[j].Name
			})

			if !test.wantErr && gotResult.Error != nil {
				t.Errorf("Unexpected error: %v", gotResult.Error)
			} else if !test.wantErr && !reflect.DeepEqual(gotResult.Index, test.wantIndex) {
				t.Errorf("Got: %v, Expected: %v", gotResult.Index, test.wantIndex)
			} else if test.wantErr && !strings.HasPrefix(gotResult.Error.Error(), test.wantErrStr) {
				t.Errorf("Got: %v, Expected: %v", gotResult.Error.Error(), test.wantErrStr)
			}
		})
	}
}
