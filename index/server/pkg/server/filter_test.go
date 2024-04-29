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

package server

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	indexSchema "github.com/devfile/registry-support/index/generator/schema"
	"github.com/devfile/registry-support/index/server/pkg/util"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

var testIndexSchema = []indexSchema.Schema{
	{
		Name:        "devfileA",
		DisplayName: "Python",
		Description: "A python stack.",
		Attributes: map[string]apiext.JSON{
			"attributeA": {},
			"attributeB": {},
			"attributeC": {},
			"attributeD": {},
		},
		Architectures: []string{"amd64", "arm64"},
		Tags:          []string{"Python", "Django", "Flask"},
		Versions: []indexSchema.Version{
			{
				Version:         "v1.0.0",
				Description:     "A python stack.",
				Architectures:   []string{"amd64"},
				Tags:            []string{"Python"},
				Resources:       []string{"devfile.yaml"},
				StarterProjects: []string{"starterA"},
			},
			{
				Version:         "v1.1.0",
				Description:     "A python stack.",
				Architectures:   []string{"amd64", "arm64"},
				Tags:            []string{"Python", "Django"},
				Resources:       []string{"devfile.yaml"},
				StarterProjects: []string{"starterA", "starterB"},
				Links: map[string]string{
					"linkA": "git.test.com",
					"linkB": "https://http.test.com",
					"linkC": "https://another.testlink.ca",
				},
				CommandGroups: map[indexSchema.CommandGroupKind]bool{
					indexSchema.BuildCommandGroupKind: true,
					indexSchema.RunCommandGroupKind:   true,
				},
				DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
					indexSchema.InnerloopKind: true,
					indexSchema.OuterloopKind: false,
				},
			},
			{
				Version:         "v2.0.0",
				Description:     "A python stack.",
				Icon:            "devfileA.png",
				Default:         true,
				Tags:            []string{"Python", "Flask"},
				Resources:       []string{"devfile.yaml", "archive.tar"},
				StarterProjects: []string{"starterA", "starterB"},
				CommandGroups: map[indexSchema.CommandGroupKind]bool{
					indexSchema.DebugCommandGroupKind:  false,
					indexSchema.DeployCommandGroupKind: false,
					indexSchema.BuildCommandGroupKind:  true,
					indexSchema.RunCommandGroupKind:    true,
					indexSchema.TestCommandGroupKind:   false,
				},
				DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
					indexSchema.InnerloopKind: true,
					indexSchema.OuterloopKind: false,
				},
			},
		},
	},
	{
		Name:        "devfileB",
		DisplayName: "Python",
		Description: "A python sample.",
		Icon:        "devfileB.ico",
		Attributes: map[string]apiext.JSON{
			"attributeA": {},
			"attributeC": {},
			"attributeD": {},
			"attributeE": {},
		},
		Architectures:   []string{"amd64"},
		Tags:            []string{"Python"},
		Resources:       []string{"devfile.yaml"},
		StarterProjects: []string{"starterB"},
		Links: map[string]string{
			"linkA": "git.test.com",
			"linkC": "https://another.testlink.ca",
		},
		CommandGroups: map[indexSchema.CommandGroupKind]bool{
			indexSchema.BuildCommandGroupKind: true,
			indexSchema.RunCommandGroupKind:   true,
		},
		DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
			indexSchema.InnerloopKind: true,
			indexSchema.OuterloopKind: false,
		},
	},
	{
		Name:        "devfileC",
		DisplayName: "Flask",
		Icon:        "devfileC.jpg",
		Attributes: map[string]apiext.JSON{
			"attributeB": {},
			"attributeC": {},
			"attributeD": {},
			"attributeE": {},
		},
		Resources: []string{"devfile.yaml", "archive.tar"},
		Links: map[string]string{
			"linkA": "git.test.com",
		},
		CommandGroups: map[indexSchema.CommandGroupKind]bool{
			indexSchema.DeployCommandGroupKind: false,
			indexSchema.RunCommandGroupKind:    true,
		},
		DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
			indexSchema.InnerloopKind: true,
			indexSchema.OuterloopKind: false,
		},
		Versions: []indexSchema.Version{
			{
				Version:         "v1.0.0",
				Icon:            "devfileC.png",
				Resources:       []string{"devfile.yaml", "archive.tar"},
				StarterProjects: []string{"starterA", "starterC"},
				Default:         true,
				Links: map[string]string{
					"linkA": "git.test.com",
				},
				CommandGroups: map[indexSchema.CommandGroupKind]bool{
					indexSchema.DeployCommandGroupKind: false,
					indexSchema.RunCommandGroupKind:    true,
				},
				DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
					indexSchema.InnerloopKind: true,
					indexSchema.OuterloopKind: false,
				},
			},
			{
				Version:         "v2.0.0",
				Icon:            "devfileC.ico",
				StarterProjects: []string{"starterA", "starterB"},
				Links: map[string]string{
					"linkA": "git.test.com",
					"linkC": "https://another.testlink.ca",
				},
				CommandGroups: map[indexSchema.CommandGroupKind]bool{
					indexSchema.DebugCommandGroupKind:  false,
					indexSchema.DeployCommandGroupKind: false,
					indexSchema.BuildCommandGroupKind:  true,
					indexSchema.RunCommandGroupKind:    true,
					indexSchema.TestCommandGroupKind:   false,
				},
				DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
					indexSchema.InnerloopKind: true,
					indexSchema.OuterloopKind: false,
				},
			},
		},
	},
	{
		Name: "devfileD",
	},
}

func TestFilterFieldbyParam(t *testing.T) {
	tests := []struct {
		name       string
		index      []indexSchema.Schema
		v1Index    bool
		paramName  string
		paramValue any
		wantIndex  []indexSchema.Schema
	}{
		{
			name:       "Case 1: string parameter",
			index:      testIndexSchema,
			v1Index:    true,
			paramName:  util.ParamIcon,
			paramValue: ".jpg",
			wantIndex: []indexSchema.Schema{
				{
					Name:        "devfileC",
					DisplayName: "Flask",
					Icon:        "devfileC.jpg",
					Attributes: map[string]apiext.JSON{
						"attributeB": {},
						"attributeC": {},
						"attributeD": {},
						"attributeE": {},
					},
					Resources: []string{"devfile.yaml", "archive.tar"},
					Links: map[string]string{
						"linkA": "git.test.com",
					},
					CommandGroups: map[indexSchema.CommandGroupKind]bool{
						indexSchema.DeployCommandGroupKind: false,
						indexSchema.RunCommandGroupKind:    true,
					},
					DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
						indexSchema.InnerloopKind: true,
						indexSchema.OuterloopKind: false,
					},
					Versions: []indexSchema.Version{
						{
							Version:         "v1.0.0",
							Icon:            "devfileC.png",
							Resources:       []string{"devfile.yaml", "archive.tar"},
							StarterProjects: []string{"starterA", "starterC"},
							Default:         true,
							Links: map[string]string{
								"linkA": "git.test.com",
							},
							CommandGroups: map[indexSchema.CommandGroupKind]bool{
								indexSchema.DeployCommandGroupKind: false,
								indexSchema.RunCommandGroupKind:    true,
							},
							DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
								indexSchema.InnerloopKind: true,
								indexSchema.OuterloopKind: false,
							},
						},
						{
							Version:         "v2.0.0",
							Icon:            "devfileC.ico",
							StarterProjects: []string{"starterA", "starterB"},
							Links: map[string]string{
								"linkA": "git.test.com",
								"linkC": "https://another.testlink.ca",
							},
							CommandGroups: map[indexSchema.CommandGroupKind]bool{
								indexSchema.DebugCommandGroupKind:  false,
								indexSchema.DeployCommandGroupKind: false,
								indexSchema.BuildCommandGroupKind:  true,
								indexSchema.RunCommandGroupKind:    true,
								indexSchema.TestCommandGroupKind:   false,
							},
							DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
								indexSchema.InnerloopKind: true,
								indexSchema.OuterloopKind: false,
							},
						},
					},
				},
			},
		},
		{
			name:       "Case 2: string parameter v2",
			index:      testIndexSchema,
			paramName:  util.ParamIcon,
			paramValue: ".png",
			wantIndex: []indexSchema.Schema{
				{
					Name:        "devfileA",
					DisplayName: "Python",
					Description: "A python stack.",
					Attributes: map[string]apiext.JSON{
						"attributeA": {},
						"attributeB": {},
						"attributeC": {},
						"attributeD": {},
					},
					Architectures: []string{"amd64", "arm64"},
					Tags:          []string{"Python", "Django", "Flask"},
					Versions: []indexSchema.Version{
						{
							Version:         "v2.0.0",
							Description:     "A python stack.",
							Icon:            "devfileA.png",
							Default:         true,
							Tags:            []string{"Python", "Flask"},
							Resources:       []string{"devfile.yaml", "archive.tar"},
							StarterProjects: []string{"starterA", "starterB"},
							CommandGroups: map[indexSchema.CommandGroupKind]bool{
								indexSchema.DebugCommandGroupKind:  false,
								indexSchema.DeployCommandGroupKind: false,
								indexSchema.BuildCommandGroupKind:  true,
								indexSchema.RunCommandGroupKind:    true,
								indexSchema.TestCommandGroupKind:   false,
							},
							DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
								indexSchema.InnerloopKind: true,
								indexSchema.OuterloopKind: false,
							},
						},
					},
				},
				{
					Name:        "devfileC",
					DisplayName: "Flask",
					Icon:        "devfileC.jpg",
					Attributes: map[string]apiext.JSON{
						"attributeB": {},
						"attributeC": {},
						"attributeD": {},
						"attributeE": {},
					},
					Resources: []string{"devfile.yaml", "archive.tar"},
					Links: map[string]string{
						"linkA": "git.test.com",
					},
					CommandGroups: map[indexSchema.CommandGroupKind]bool{
						indexSchema.DeployCommandGroupKind: false,
						indexSchema.RunCommandGroupKind:    true,
					},
					DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
						indexSchema.InnerloopKind: true,
						indexSchema.OuterloopKind: false,
					},
					Versions: []indexSchema.Version{
						{
							Version:         "v1.0.0",
							Icon:            "devfileC.png",
							Resources:       []string{"devfile.yaml", "archive.tar"},
							StarterProjects: []string{"starterA", "starterC"},
							Default:         true,
							Links: map[string]string{
								"linkA": "git.test.com",
							},
							CommandGroups: map[indexSchema.CommandGroupKind]bool{
								indexSchema.DeployCommandGroupKind: false,
								indexSchema.RunCommandGroupKind:    true,
							},
							DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
								indexSchema.InnerloopKind: true,
								indexSchema.OuterloopKind: false,
							},
						},
					},
				},
			},
		},
		{
			name:       "Case 3: Non-string parameter",
			index:      testIndexSchema,
			paramName:  util.ParamDefault,
			paramValue: true,
			wantIndex: []indexSchema.Schema{
				{
					Name:        "devfileA",
					DisplayName: "Python",
					Description: "A python stack.",
					Attributes: map[string]apiext.JSON{
						"attributeA": {},
						"attributeB": {},
						"attributeC": {},
						"attributeD": {},
					},
					Architectures: []string{"amd64", "arm64"},
					Tags:          []string{"Python", "Django", "Flask"},
					Versions: []indexSchema.Version{
						{
							Version:         "v2.0.0",
							Description:     "A python stack.",
							Icon:            "devfileA.png",
							Default:         true,
							Tags:            []string{"Python", "Flask"},
							Resources:       []string{"devfile.yaml", "archive.tar"},
							StarterProjects: []string{"starterA", "starterB"},
							CommandGroups: map[indexSchema.CommandGroupKind]bool{
								indexSchema.DebugCommandGroupKind:  false,
								indexSchema.DeployCommandGroupKind: false,
								indexSchema.BuildCommandGroupKind:  true,
								indexSchema.RunCommandGroupKind:    true,
								indexSchema.TestCommandGroupKind:   false,
							},
							DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
								indexSchema.InnerloopKind: true,
								indexSchema.OuterloopKind: false,
							},
						},
					},
				},
				{
					Name:        "devfileC",
					DisplayName: "Flask",
					Icon:        "devfileC.jpg",
					Attributes: map[string]apiext.JSON{
						"attributeB": {},
						"attributeC": {},
						"attributeD": {},
						"attributeE": {},
					},
					Resources: []string{"devfile.yaml", "archive.tar"},
					Links: map[string]string{
						"linkA": "git.test.com",
					},
					CommandGroups: map[indexSchema.CommandGroupKind]bool{
						indexSchema.DeployCommandGroupKind: false,
						indexSchema.RunCommandGroupKind:    true,
					},
					DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
						indexSchema.InnerloopKind: true,
						indexSchema.OuterloopKind: false,
					},
					Versions: []indexSchema.Version{
						{
							Version:         "v1.0.0",
							Icon:            "devfileC.png",
							Resources:       []string{"devfile.yaml", "archive.tar"},
							StarterProjects: []string{"starterA", "starterC"},
							Default:         true,
							Links: map[string]string{
								"linkA": "git.test.com",
							},
							CommandGroups: map[indexSchema.CommandGroupKind]bool{
								indexSchema.DeployCommandGroupKind: false,
								indexSchema.RunCommandGroupKind:    true,
							},
							DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
								indexSchema.InnerloopKind: true,
								indexSchema.OuterloopKind: false,
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotResult := filterFieldbyParam(test.index, test.v1Index, test.paramName, test.paramValue)

			if gotResult.Error != nil {
				t.Errorf("unexpected error: %v", gotResult.Error)
				return
			}

			sort.Slice(gotResult.Index, func(i, j int) bool {
				return gotResult.Index[i].Name < gotResult.Index[j].Name
			})
			sort.Slice(test.wantIndex, func(i, j int) bool {
				return test.wantIndex[i].Name < test.wantIndex[j].Name
			})

			if !reflect.DeepEqual(gotResult.Index, test.wantIndex) {
				t.Errorf("expected: %v, got: %v", test.wantIndex, gotResult.Index)
			}
		})
	}
}

func TestFilterFieldsByParams(t *testing.T) {
	tests := []struct {
		name       string
		index      []indexSchema.Schema
		v1Index    bool
		params     IndexParams
		wantIndex  []indexSchema.Schema
		wantErr    bool
		wantErrStr string
	}{
		{
			name:  "Case 1: Single filter",
			index: testIndexSchema,
			params: IndexParams{
				Arch: &[]string{"arm64"},
			},
			v1Index: true,
			wantIndex: []indexSchema.Schema{
				{
					Name:        "devfileA",
					DisplayName: "Python",
					Description: "A python stack.",
					Attributes: map[string]apiext.JSON{
						"attributeA": {},
						"attributeB": {},
						"attributeC": {},
						"attributeD": {},
					},
					Architectures: []string{"amd64", "arm64"},
					Tags:          []string{"Python", "Django", "Flask"},
					Versions: []indexSchema.Version{
						{
							Version:         "v1.0.0",
							Description:     "A python stack.",
							Architectures:   []string{"amd64"},
							Tags:            []string{"Python"},
							Resources:       []string{"devfile.yaml"},
							StarterProjects: []string{"starterA"},
						},
						{
							Version:         "v1.1.0",
							Description:     "A python stack.",
							Architectures:   []string{"amd64", "arm64"},
							Tags:            []string{"Python", "Django"},
							Resources:       []string{"devfile.yaml"},
							StarterProjects: []string{"starterA", "starterB"},
							Links: map[string]string{
								"linkA": "git.test.com",
								"linkB": "https://http.test.com",
								"linkC": "https://another.testlink.ca",
							},
							CommandGroups: map[indexSchema.CommandGroupKind]bool{
								indexSchema.BuildCommandGroupKind: true,
								indexSchema.RunCommandGroupKind:   true,
							},
							DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
								indexSchema.InnerloopKind: true,
								indexSchema.OuterloopKind: false,
							},
						},
						{
							Version:         "v2.0.0",
							Description:     "A python stack.",
							Icon:            "devfileA.png",
							Default:         true,
							Tags:            []string{"Python", "Flask"},
							Resources:       []string{"devfile.yaml", "archive.tar"},
							StarterProjects: []string{"starterA", "starterB"},
							CommandGroups: map[indexSchema.CommandGroupKind]bool{
								indexSchema.DebugCommandGroupKind:  false,
								indexSchema.DeployCommandGroupKind: false,
								indexSchema.BuildCommandGroupKind:  true,
								indexSchema.RunCommandGroupKind:    true,
								indexSchema.TestCommandGroupKind:   false,
							},
							DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
								indexSchema.InnerloopKind: true,
								indexSchema.OuterloopKind: false,
							},
						},
					},
				},
				{
					Name:        "devfileC",
					DisplayName: "Flask",
					Icon:        "devfileC.jpg",
					Attributes: map[string]apiext.JSON{
						"attributeB": {},
						"attributeC": {},
						"attributeD": {},
						"attributeE": {},
					},
					Resources: []string{"devfile.yaml", "archive.tar"},
					Links: map[string]string{
						"linkA": "git.test.com",
					},
					CommandGroups: map[indexSchema.CommandGroupKind]bool{
						indexSchema.DeployCommandGroupKind: false,
						indexSchema.RunCommandGroupKind:    true,
					},
					DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
						indexSchema.InnerloopKind: true,
						indexSchema.OuterloopKind: false,
					},
					Versions: []indexSchema.Version{
						{
							Version:         "v1.0.0",
							Icon:            "devfileC.png",
							Resources:       []string{"devfile.yaml", "archive.tar"},
							StarterProjects: []string{"starterA", "starterC"},
							Default:         true,
							Links: map[string]string{
								"linkA": "git.test.com",
							},
							CommandGroups: map[indexSchema.CommandGroupKind]bool{
								indexSchema.DeployCommandGroupKind: false,
								indexSchema.RunCommandGroupKind:    true,
							},
							DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
								indexSchema.InnerloopKind: true,
								indexSchema.OuterloopKind: false,
							},
						},
						{
							Version:         "v2.0.0",
							Icon:            "devfileC.ico",
							StarterProjects: []string{"starterA", "starterB"},
							Links: map[string]string{
								"linkA": "git.test.com",
								"linkC": "https://another.testlink.ca",
							},
							CommandGroups: map[indexSchema.CommandGroupKind]bool{
								indexSchema.DebugCommandGroupKind:  false,
								indexSchema.DeployCommandGroupKind: false,
								indexSchema.BuildCommandGroupKind:  true,
								indexSchema.RunCommandGroupKind:    true,
								indexSchema.TestCommandGroupKind:   false,
							},
							DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
								indexSchema.InnerloopKind: true,
								indexSchema.OuterloopKind: false,
							},
						},
					},
				},
				{
					Name: "devfileD",
				},
			},
		},
		{
			name:  "Case 2: Single filter v2",
			index: testIndexSchema,
			params: IndexParams{
				Arch: &[]string{"arm64"},
			},
			wantIndex: []indexSchema.Schema{
				{
					Name:        "devfileA",
					DisplayName: "Python",
					Description: "A python stack.",
					Attributes: map[string]apiext.JSON{
						"attributeA": {},
						"attributeB": {},
						"attributeC": {},
						"attributeD": {},
					},
					Architectures: []string{"amd64", "arm64"},
					Tags:          []string{"Python", "Django", "Flask"},
					Versions: []indexSchema.Version{
						{
							Version:         "v1.1.0",
							Description:     "A python stack.",
							Architectures:   []string{"amd64", "arm64"},
							Tags:            []string{"Python", "Django"},
							Resources:       []string{"devfile.yaml"},
							StarterProjects: []string{"starterA", "starterB"},
							Links: map[string]string{
								"linkA": "git.test.com",
								"linkB": "https://http.test.com",
								"linkC": "https://another.testlink.ca",
							},
							CommandGroups: map[indexSchema.CommandGroupKind]bool{
								indexSchema.BuildCommandGroupKind: true,
								indexSchema.RunCommandGroupKind:   true,
							},
							DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
								indexSchema.InnerloopKind: true,
								indexSchema.OuterloopKind: false,
							},
						},
						{
							Version:         "v2.0.0",
							Description:     "A python stack.",
							Icon:            "devfileA.png",
							Default:         true,
							Tags:            []string{"Python", "Flask"},
							Resources:       []string{"devfile.yaml", "archive.tar"},
							StarterProjects: []string{"starterA", "starterB"},
							CommandGroups: map[indexSchema.CommandGroupKind]bool{
								indexSchema.DebugCommandGroupKind:  false,
								indexSchema.DeployCommandGroupKind: false,
								indexSchema.BuildCommandGroupKind:  true,
								indexSchema.RunCommandGroupKind:    true,
								indexSchema.TestCommandGroupKind:   false,
							},
							DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
								indexSchema.InnerloopKind: true,
								indexSchema.OuterloopKind: false,
							},
						},
					},
				},
				{
					Name:        "devfileC",
					DisplayName: "Flask",
					Icon:        "devfileC.jpg",
					Attributes: map[string]apiext.JSON{
						"attributeB": {},
						"attributeC": {},
						"attributeD": {},
						"attributeE": {},
					},
					Resources: []string{"devfile.yaml", "archive.tar"},
					Links: map[string]string{
						"linkA": "git.test.com",
					},
					CommandGroups: map[indexSchema.CommandGroupKind]bool{
						indexSchema.DeployCommandGroupKind: false,
						indexSchema.RunCommandGroupKind:    true,
					},
					DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
						indexSchema.InnerloopKind: true,
						indexSchema.OuterloopKind: false,
					},
					Versions: []indexSchema.Version{
						{
							Version:         "v1.0.0",
							Icon:            "devfileC.png",
							Resources:       []string{"devfile.yaml", "archive.tar"},
							StarterProjects: []string{"starterA", "starterC"},
							Default:         true,
							Links: map[string]string{
								"linkA": "git.test.com",
							},
							CommandGroups: map[indexSchema.CommandGroupKind]bool{
								indexSchema.DeployCommandGroupKind: false,
								indexSchema.RunCommandGroupKind:    true,
							},
							DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
								indexSchema.InnerloopKind: true,
								indexSchema.OuterloopKind: false,
							},
						},
						{
							Version:         "v2.0.0",
							Icon:            "devfileC.ico",
							StarterProjects: []string{"starterA", "starterB"},
							Links: map[string]string{
								"linkA": "git.test.com",
								"linkC": "https://another.testlink.ca",
							},
							CommandGroups: map[indexSchema.CommandGroupKind]bool{
								indexSchema.DebugCommandGroupKind:  false,
								indexSchema.DeployCommandGroupKind: false,
								indexSchema.BuildCommandGroupKind:  true,
								indexSchema.RunCommandGroupKind:    true,
								indexSchema.TestCommandGroupKind:   false,
							},
							DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
								indexSchema.InnerloopKind: true,
								indexSchema.OuterloopKind: false,
							},
						},
					},
				},
				{
					Name: "devfileD",
				},
			},
		},
		{
			name:  "Case 3: Multi filter",
			index: testIndexSchema,
			params: IndexParams{
				Arch: &[]string{"arm64"},
				CommandGroups: &[]string{
					string(indexSchema.RunCommandGroupKind),
				},
			},
			v1Index: true,
			wantIndex: []indexSchema.Schema{
				{
					Name:        "devfileC",
					DisplayName: "Flask",
					Icon:        "devfileC.jpg",
					Attributes: map[string]apiext.JSON{
						"attributeB": {},
						"attributeC": {},
						"attributeD": {},
						"attributeE": {},
					},
					Resources: []string{"devfile.yaml", "archive.tar"},
					Links: map[string]string{
						"linkA": "git.test.com",
					},
					CommandGroups: map[indexSchema.CommandGroupKind]bool{
						indexSchema.DeployCommandGroupKind: false,
						indexSchema.RunCommandGroupKind:    true,
					},
					DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
						indexSchema.InnerloopKind: true,
						indexSchema.OuterloopKind: false,
					},
					Versions: []indexSchema.Version{
						{
							Version:         "v1.0.0",
							Icon:            "devfileC.png",
							Resources:       []string{"devfile.yaml", "archive.tar"},
							StarterProjects: []string{"starterA", "starterC"},
							Default:         true,
							Links: map[string]string{
								"linkA": "git.test.com",
							},
							CommandGroups: map[indexSchema.CommandGroupKind]bool{
								indexSchema.DeployCommandGroupKind: false,
								indexSchema.RunCommandGroupKind:    true,
							},
							DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
								indexSchema.InnerloopKind: true,
								indexSchema.OuterloopKind: false,
							},
						},
						{
							Version:         "v2.0.0",
							Icon:            "devfileC.ico",
							StarterProjects: []string{"starterA", "starterB"},
							Links: map[string]string{
								"linkA": "git.test.com",
								"linkC": "https://another.testlink.ca",
							},
							CommandGroups: map[indexSchema.CommandGroupKind]bool{
								indexSchema.DebugCommandGroupKind:  false,
								indexSchema.DeployCommandGroupKind: false,
								indexSchema.BuildCommandGroupKind:  true,
								indexSchema.RunCommandGroupKind:    true,
								indexSchema.TestCommandGroupKind:   false,
							},
							DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
								indexSchema.InnerloopKind: true,
								indexSchema.OuterloopKind: false,
							},
						},
					},
				},
			},
		},
		{
			name:  "Case 4: Multi filter v2",
			index: testIndexSchema,
			params: IndexParams{
				Arch: &[]string{"arm64"},
				CommandGroups: &[]string{
					string(indexSchema.BuildCommandGroupKind),
					string(indexSchema.RunCommandGroupKind),
				},
				DeploymentScopes: &[]string{
					string(indexSchema.InnerloopKind),
				},
			},
			wantIndex: []indexSchema.Schema{
				{
					Name:        "devfileA",
					DisplayName: "Python",
					Description: "A python stack.",
					Attributes: map[string]apiext.JSON{
						"attributeA": {},
						"attributeB": {},
						"attributeC": {},
						"attributeD": {},
					},
					Architectures: []string{"amd64", "arm64"},
					Tags:          []string{"Python", "Django", "Flask"},
					Versions: []indexSchema.Version{
						{
							Version:         "v1.1.0",
							Description:     "A python stack.",
							Architectures:   []string{"amd64", "arm64"},
							Tags:            []string{"Python", "Django"},
							Resources:       []string{"devfile.yaml"},
							StarterProjects: []string{"starterA", "starterB"},
							Links: map[string]string{
								"linkA": "git.test.com",
								"linkB": "https://http.test.com",
								"linkC": "https://another.testlink.ca",
							},
							CommandGroups: map[indexSchema.CommandGroupKind]bool{
								indexSchema.BuildCommandGroupKind: true,
								indexSchema.RunCommandGroupKind:   true,
							},
							DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
								indexSchema.InnerloopKind: true,
								indexSchema.OuterloopKind: false,
							},
						},
						{
							Version:         "v2.0.0",
							Description:     "A python stack.",
							Icon:            "devfileA.png",
							Default:         true,
							Tags:            []string{"Python", "Flask"},
							Resources:       []string{"devfile.yaml", "archive.tar"},
							StarterProjects: []string{"starterA", "starterB"},
							CommandGroups: map[indexSchema.CommandGroupKind]bool{
								indexSchema.DebugCommandGroupKind:  false,
								indexSchema.DeployCommandGroupKind: false,
								indexSchema.BuildCommandGroupKind:  true,
								indexSchema.RunCommandGroupKind:    true,
								indexSchema.TestCommandGroupKind:   false,
							},
							DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
								indexSchema.InnerloopKind: true,
								indexSchema.OuterloopKind: false,
							},
						},
					},
				},
				{
					Name:        "devfileC",
					DisplayName: "Flask",
					Icon:        "devfileC.jpg",
					Attributes: map[string]apiext.JSON{
						"attributeB": {},
						"attributeC": {},
						"attributeD": {},
						"attributeE": {},
					},
					Resources: []string{"devfile.yaml", "archive.tar"},
					Links: map[string]string{
						"linkA": "git.test.com",
					},
					CommandGroups: map[indexSchema.CommandGroupKind]bool{
						indexSchema.DeployCommandGroupKind: false,
						indexSchema.RunCommandGroupKind:    true,
					},
					DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
						indexSchema.InnerloopKind: true,
						indexSchema.OuterloopKind: false,
					},
					Versions: []indexSchema.Version{
						{
							Version:         "v2.0.0",
							Icon:            "devfileC.ico",
							StarterProjects: []string{"starterA", "starterB"},
							Links: map[string]string{
								"linkA": "git.test.com",
								"linkC": "https://another.testlink.ca",
							},
							CommandGroups: map[indexSchema.CommandGroupKind]bool{
								indexSchema.DebugCommandGroupKind:  false,
								indexSchema.DeployCommandGroupKind: false,
								indexSchema.BuildCommandGroupKind:  true,
								indexSchema.RunCommandGroupKind:    true,
								indexSchema.TestCommandGroupKind:   false,
							},
							DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
								indexSchema.InnerloopKind: true,
								indexSchema.OuterloopKind: false,
							},
						},
					},
				},
			},
		},
		{
			name:  "Case 5: Blank result",
			index: testIndexSchema,
			params: IndexParams{
				Arch:           &[]string{"arm64"},
				AttributeNames: &[]string{"attributeE"},
				CommandGroups: &[]string{
					string(indexSchema.DeployCommandGroupKind),
					string(indexSchema.RunCommandGroupKind),
				},
			},
			wantIndex: []indexSchema.Schema{},
		},
		{
			name:      "Case 6: Blank filter",
			index:     testIndexSchema,
			params:    IndexParams{},
			wantIndex: testIndexSchema,
		},
		{
			name:    "Case 7: Error",
			index:   testIndexSchema,
			params:  IndexParams{},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotIndex, gotErr := filterFieldsByParams(test.index, test.v1Index, test.params)

			// sorting is not consistent in output
			sort.Slice(gotIndex, func(i, j int) bool {
				return gotIndex[i].Name < gotIndex[j].Name
			})
			sort.Slice(test.wantIndex, func(i, j int) bool {
				return test.wantIndex[i].Name < test.wantIndex[j].Name
			})

			if test.wantErr && gotErr != nil && !strings.Contains(gotErr.Error(), test.wantErrStr) {
				t.Errorf("unexpected error %v", gotErr)
			} else if !test.wantErr && gotErr != nil {
				t.Errorf("unexpected error %v", gotErr)
			} else if !test.wantErr && !reflect.DeepEqual(gotIndex, test.wantIndex) {
				t.Errorf("expected: %v, got: %v", test.wantIndex, gotIndex)
			}
		})
	}
}
