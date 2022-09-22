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
	"reflect"
	"testing"

	indexSchema "github.com/devfile/registry-support/index/generator/schema"
)

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
			gotIndex := FilterDevfileArchitectures(test.index, test.archs, test.v1Index)
			if !reflect.DeepEqual(gotIndex, test.wantIndex) {
				t.Errorf("Got: %v, Expected: %v", gotIndex, test.wantIndex)
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
			gotIndex, err := FilterDevfileSchemaVersion(test.index, test.minSchemaVersion, test.maxSchemaVersion)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !reflect.DeepEqual(gotIndex, test.wantIndex) {
				t.Errorf("Got: %v, Expected: %v", gotIndex, test.wantIndex)
			}
		})
	}
}
