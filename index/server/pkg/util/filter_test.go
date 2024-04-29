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
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// filterDevfileStrArrayFieldTestCase type of test case to be used with FilterDevfileStrArrayField
type filterDevfileStrArrayFieldTestCase struct {
	Name      string
	Index     []indexSchema.Schema
	FieldName string
	Values    []string
	V1Index   bool
	WantIndex []indexSchema.Schema
}

// filterDevfileStrFieldTestCase type of test case to be used with FilterDevfileStrField
type filterDevfileStrFieldTestCase struct {
	Name       string
	Index      []indexSchema.Schema
	FieldName  string
	Value      string
	V1Index    bool
	WantIndex  []indexSchema.Schema
	WantErr    bool
	WantErrStr string
}

var (
	// ============================================
	// Filter Devfile String Array Field Test Cases
	// ============================================
	filterAttributeNamesTestCases = []filterDevfileStrArrayFieldTestCase{
		{
			Name:      "two attribute filters",
			FieldName: ArrayParamAttributeNames,
			Index: []indexSchema.Schema{
				{
					Name: "devfileA",
					Attributes: map[string]apiext.JSON{
						"attributeA": {},
						"attributeB": {},
						"attributeC": {},
						"attributeD": {},
					},
				},
				{
					Name: "devfileB",
					Attributes: map[string]apiext.JSON{
						"attributeA": {},
						"attributeC": {},
						"attributeD": {},
						"attributeE": {},
					},
				},
				{
					Name: "devfileC",
					Attributes: map[string]apiext.JSON{
						"attributeB": {},
						"attributeC": {},
						"attributeD": {},
						"attributeE": {},
					},
				},
				{
					Name: "devfileD",
				},
			},
			Values:  []string{"attributeB", "attributeD"},
			V1Index: true,
			WantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
					Attributes: map[string]apiext.JSON{
						"attributeA": {},
						"attributeB": {},
						"attributeC": {},
						"attributeD": {},
					},
				},
				{
					Name: "devfileC",
					Attributes: map[string]apiext.JSON{
						"attributeB": {},
						"attributeC": {},
						"attributeD": {},
						"attributeE": {},
					},
				},
			},
		},
		{
			Name:      "two attribute filters v2",
			FieldName: ArrayParamAttributeNames,
			Index: []indexSchema.Schema{
				{
					Name: "devfileA",
					Attributes: map[string]apiext.JSON{
						"attributeA": {},
						"attributeB": {},
						"attributeC": {},
						"attributeD": {},
					},
					Versions: []indexSchema.Version{
						{
							Version: "v1.0.0",
						},
						{
							Version: "v1.1.0",
						},
					},
				},
				{
					Name: "devfileB",
					Attributes: map[string]apiext.JSON{
						"attributeA": {},
						"attributeC": {},
						"attributeD": {},
						"attributeE": {},
					},
				},
				{
					Name: "devfileC",
					Attributes: map[string]apiext.JSON{
						"attributeB": {},
						"attributeC": {},
						"attributeD": {},
						"attributeE": {},
					},
				},
				{
					Name: "devfileD",
				},
			},
			Values: []string{"attributeB", "attributeD"},
			WantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
					Attributes: map[string]apiext.JSON{
						"attributeA": {},
						"attributeB": {},
						"attributeC": {},
						"attributeD": {},
					},
					Versions: []indexSchema.Version{
						{
							Version: "v1.0.0",
						},
						{
							Version: "v1.1.0",
						},
					},
				},
				{
					Name: "devfileC",
					Attributes: map[string]apiext.JSON{
						"attributeB": {},
						"attributeC": {},
						"attributeD": {},
						"attributeE": {},
					},
				},
			},
		},
	}
	filterArchitecturesTestCases = []filterDevfileStrArrayFieldTestCase{
		{
			Name:      "two arch filters",
			FieldName: ArrayParamArchitectures,
			Index: []indexSchema.Schema{
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
			V1Index: true,
			Values:  []string{"amd64", "arm64"},
			WantIndex: []indexSchema.Schema{
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
			Name:      "two arch filters with v2 index",
			FieldName: ArrayParamArchitectures,
			Index: []indexSchema.Schema{
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
			V1Index: false,
			Values:  []string{"amd64", "arm64"},
			WantIndex: []indexSchema.Schema{
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
	}
	filterTagsTestCases = []filterDevfileStrArrayFieldTestCase{
		{
			Name:      "two tag filters",
			FieldName: ArrayParamTags,
			Index: []indexSchema.Schema{
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
			V1Index: true,
			Values:  []string{"Python", "Django"},
			WantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
					Tags: []string{"Python", "Django"},
				},
			},
		},
		{
			Name:      "two tag filters with v2 index",
			FieldName: ArrayParamTags,
			Index: []indexSchema.Schema{
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
			V1Index: false,
			Values:  []string{"Python", "Django"},
			WantIndex: []indexSchema.Schema{
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
	filterResourcesTestCases = []filterDevfileStrArrayFieldTestCase{
		{
			Name:      "two resource filters",
			FieldName: ArrayParamResources,
			Index: []indexSchema.Schema{
				{
					Name:      "devfileA",
					Resources: []string{"devfile.yaml", "archive.tar"},
				},
				{
					Name:      "devfileB",
					Resources: []string{"devfile.yaml"},
				},
				{
					Name:      "devfileC",
					Resources: []string{"devfile.yaml", "archive.tar"},
				},
				{
					Name: "devfileD",
				},
			},
			V1Index: true,
			Values:  []string{"devfile.yaml", "archive.tar"},
			WantIndex: []indexSchema.Schema{
				{
					Name:      "devfileA",
					Resources: []string{"devfile.yaml", "archive.tar"},
				},
				{
					Name:      "devfileC",
					Resources: []string{"devfile.yaml", "archive.tar"},
				},
			},
		},
		{
			Name:      "two resource filters with v2 index",
			FieldName: ArrayParamResources,
			Index: []indexSchema.Schema{
				{
					Name: "devfileA",
					Versions: []indexSchema.Version{
						{
							Version:   "1.0.0",
							Resources: []string{"devfile.yaml"},
						},
						{
							Version:   "1.1.0",
							Resources: []string{"devfile.yaml"},
						},
						{
							Version:   "2.0.0",
							Resources: []string{"devfile.yaml", "archive.tar"},
						},
					},
				},
				{
					Name:      "devfileB",
					Resources: []string{"devfile.yaml"},
				},
				{
					Name:      "devfileC",
					Resources: []string{"devfile.yaml", "archive.tar"},
					Versions: []indexSchema.Version{
						{
							Version:   "1.0.0",
							Resources: []string{"devfile.yaml", "archive.tar"},
						},
					},
				},
				{
					Name: "devfileD",
				},
			},
			V1Index: false,
			Values:  []string{"devfile.yaml", "archive.tar"},
			WantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
					Versions: []indexSchema.Version{
						{
							Version:   "2.0.0",
							Resources: []string{"devfile.yaml", "archive.tar"},
						},
					},
				},
				{
					Name:      "devfileC",
					Resources: []string{"devfile.yaml", "archive.tar"},
					Versions: []indexSchema.Version{
						{
							Version:   "1.0.0",
							Resources: []string{"devfile.yaml", "archive.tar"},
						},
					},
				},
			},
		},
	}
	filterStarterProjectsTestCases = []filterDevfileStrArrayFieldTestCase{
		{
			Name:      "two starter project filters",
			FieldName: ArrayParamStarterProjects,
			Index: []indexSchema.Schema{
				{
					Name:            "devfileA",
					StarterProjects: []string{"starterA", "starterB"},
				},
				{
					Name:            "devfileB",
					StarterProjects: []string{"starterB"},
				},
				{
					Name:            "devfileC",
					StarterProjects: []string{"starterA", "starterC"},
				},
				{
					Name: "devfileD",
				},
			},
			V1Index: true,
			Values:  []string{"starterA", "starterB"},
			WantIndex: []indexSchema.Schema{
				{
					Name:            "devfileA",
					StarterProjects: []string{"starterA", "starterB"},
				},
			},
		},
		{
			Name:      "two starter project filters with v2 index",
			FieldName: ArrayParamStarterProjects,
			Index: []indexSchema.Schema{
				{
					Name: "devfileA",
					Versions: []indexSchema.Version{
						{
							Version:         "1.0.0",
							StarterProjects: []string{"starterA"},
						},
						{
							Version:         "1.1.0",
							StarterProjects: []string{"starterA", "starterB"},
						},
						{
							Version:         "2.0.0",
							StarterProjects: []string{"starterA", "starterB"},
						},
					},
				},
				{
					Name:            "devfileB",
					StarterProjects: []string{"starterB"},
				},
				{
					Name:            "devfileC",
					StarterProjects: []string{"starterA", "starterC"},
					Versions: []indexSchema.Version{
						{
							Version:         "1.0.0",
							StarterProjects: []string{"starterA", "starterC"},
						},
						{
							Version:         "2.0.0",
							StarterProjects: []string{"starterA", "starterB"},
						},
					},
				},
				{
					Name: "devfileD",
				},
			},
			V1Index: false,
			Values:  []string{"starterA", "starterB"},
			WantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
					Versions: []indexSchema.Version{
						{
							Version:         "1.1.0",
							StarterProjects: []string{"starterA", "starterB"},
						},
						{
							Version:         "2.0.0",
							StarterProjects: []string{"starterA", "starterB"},
						},
					},
				},
				{
					Name:            "devfileC",
					StarterProjects: []string{"starterA", "starterC"},
					Versions: []indexSchema.Version{
						{
							Version:         "2.0.0",
							StarterProjects: []string{"starterA", "starterB"},
						},
					},
				},
			},
		},
	}
	filterLinksTestCases = []filterDevfileStrArrayFieldTestCase{
		{
			Name:      "two link filters",
			FieldName: ArrayParamLinks,
			Index: []indexSchema.Schema{
				{
					Name: "devfileA",
					Links: map[string]string{
						"linkA": "git.test.com",
						"linkB": "https://http.test.com",
						"linkC": "https://another.testlink.ca",
					},
				},
				{
					Name: "devfileB",
					Links: map[string]string{
						"linkA": "git.test.com",
						"linkC": "https://another.testlink.ca",
					},
				},
				{
					Name: "devfileC",
					Links: map[string]string{
						"linkA": "git.test.com",
					},
				},
				{
					Name: "devfileD",
				},
			},
			V1Index: true,
			Values:  []string{"linkA", "linkC"},
			WantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
					Links: map[string]string{
						"linkA": "git.test.com",
						"linkB": "https://http.test.com",
						"linkC": "https://another.testlink.ca",
					},
				},
				{
					Name: "devfileB",
					Links: map[string]string{
						"linkA": "git.test.com",
						"linkC": "https://another.testlink.ca",
					},
				},
			},
		},
		{
			Name:      "two link filters with v2 index",
			FieldName: ArrayParamLinks,
			Index: []indexSchema.Schema{
				{
					Name: "devfileA",
					Versions: []indexSchema.Version{
						{
							Version: "1.0.0",
						},
						{
							Version: "1.1.0",
							Links: map[string]string{
								"linkA": "git.test.com",
								"linkB": "https://http.test.com",
								"linkC": "https://another.testlink.ca",
							},
						},
					},
				},
				{
					Name: "devfileB",
					Links: map[string]string{
						"linkA": "git.test.com",
						"linkC": "https://another.testlink.ca",
					},
				},
				{
					Name: "devfileC",
					Links: map[string]string{
						"linkA": "git.test.com",
					},
					Versions: []indexSchema.Version{
						{
							Version: "1.0.0",
							Links: map[string]string{
								"linkA": "git.test.com",
							},
						},
						{
							Version: "1.1.0",
							Links: map[string]string{
								"linkA": "git.test.com",
								"linkC": "https://another.testlink.ca",
							},
						},
					},
				},
				{
					Name: "devfileD",
				},
			},
			V1Index: false,
			Values:  []string{"linkA", "linkC"},
			WantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
					Versions: []indexSchema.Version{
						{
							Version: "1.1.0",
							Links: map[string]string{
								"linkA": "git.test.com",
								"linkB": "https://http.test.com",
								"linkC": "https://another.testlink.ca",
							},
						},
					},
				},
				{
					Name: "devfileB",
					Links: map[string]string{
						"linkA": "git.test.com",
						"linkC": "https://another.testlink.ca",
					},
				},
				{
					Name: "devfileC",
					Links: map[string]string{
						"linkA": "git.test.com",
					},
					Versions: []indexSchema.Version{
						{
							Version: "1.1.0",
							Links: map[string]string{
								"linkA": "git.test.com",
								"linkC": "https://another.testlink.ca",
							},
						},
					},
				},
			},
		},
	}
	filterCommandGroupsTestCases = []filterDevfileStrArrayFieldTestCase{
		{
			Name:      "two command group filters",
			FieldName: ArrayParamCommandGroups,
			Index: []indexSchema.Schema{
				{
					Name: "devfileA",
					CommandGroups: map[indexSchema.CommandGroupKind]bool{
						indexSchema.DebugCommandGroupKind:  false,
						indexSchema.DeployCommandGroupKind: false,
						indexSchema.BuildCommandGroupKind:  true,
						indexSchema.RunCommandGroupKind:    true,
						indexSchema.TestCommandGroupKind:   false,
					},
				},
				{
					Name: "devfileB",
					CommandGroups: map[indexSchema.CommandGroupKind]bool{
						indexSchema.BuildCommandGroupKind: true,
						indexSchema.RunCommandGroupKind:   true,
					},
				},
				{
					Name: "devfileC",
					CommandGroups: map[indexSchema.CommandGroupKind]bool{
						indexSchema.DeployCommandGroupKind: false,
						indexSchema.RunCommandGroupKind:    true,
					},
				},
				{
					Name: "devfileD",
				},
			},
			V1Index: true,
			Values:  []string{string(indexSchema.BuildCommandGroupKind), string(indexSchema.RunCommandGroupKind)},
			WantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
					CommandGroups: map[indexSchema.CommandGroupKind]bool{
						indexSchema.DebugCommandGroupKind:  false,
						indexSchema.DeployCommandGroupKind: false,
						indexSchema.BuildCommandGroupKind:  true,
						indexSchema.RunCommandGroupKind:    true,
						indexSchema.TestCommandGroupKind:   false,
					},
				},
				{
					Name: "devfileB",
					CommandGroups: map[indexSchema.CommandGroupKind]bool{
						indexSchema.BuildCommandGroupKind: true,
						indexSchema.RunCommandGroupKind:   true,
					},
				},
			},
		},
		{
			Name:      "two command group filters with v2 index",
			FieldName: ArrayParamCommandGroups,
			Index: []indexSchema.Schema{
				{
					Name: "devfileA",
					Versions: []indexSchema.Version{
						{
							Version: "1.0.0",
						},
						{
							Version: "1.1.0",
							CommandGroups: map[indexSchema.CommandGroupKind]bool{
								indexSchema.BuildCommandGroupKind: true,
								indexSchema.RunCommandGroupKind:   true,
							},
						},
						{
							Version: "2.0.0",
							CommandGroups: map[indexSchema.CommandGroupKind]bool{
								indexSchema.DebugCommandGroupKind:  false,
								indexSchema.DeployCommandGroupKind: false,
								indexSchema.BuildCommandGroupKind:  true,
								indexSchema.RunCommandGroupKind:    true,
								indexSchema.TestCommandGroupKind:   false,
							},
						},
					},
				},
				{
					Name: "devfileB",
					CommandGroups: map[indexSchema.CommandGroupKind]bool{
						indexSchema.BuildCommandGroupKind: true,
						indexSchema.RunCommandGroupKind:   true,
					},
				},
				{
					Name: "devfileC",
					CommandGroups: map[indexSchema.CommandGroupKind]bool{
						indexSchema.DeployCommandGroupKind: false,
						indexSchema.RunCommandGroupKind:    true,
					},
					Versions: []indexSchema.Version{
						{
							Version: "1.0.0",
							CommandGroups: map[indexSchema.CommandGroupKind]bool{
								indexSchema.DeployCommandGroupKind: false,
								indexSchema.RunCommandGroupKind:    true,
							},
						},
						{
							Version: "2.0.0",
							CommandGroups: map[indexSchema.CommandGroupKind]bool{
								indexSchema.DebugCommandGroupKind:  false,
								indexSchema.DeployCommandGroupKind: false,
								indexSchema.BuildCommandGroupKind:  true,
								indexSchema.RunCommandGroupKind:    true,
								indexSchema.TestCommandGroupKind:   false,
							},
						},
					},
				},
				{
					Name: "devfileD",
				},
			},
			V1Index: false,
			Values:  []string{string(indexSchema.BuildCommandGroupKind), string(indexSchema.RunCommandGroupKind)},
			WantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
					Versions: []indexSchema.Version{
						{
							Version: "1.1.0",
							CommandGroups: map[indexSchema.CommandGroupKind]bool{
								indexSchema.BuildCommandGroupKind: true,
								indexSchema.RunCommandGroupKind:   true,
							},
						},
						{
							Version: "2.0.0",
							CommandGroups: map[indexSchema.CommandGroupKind]bool{
								indexSchema.DebugCommandGroupKind:  false,
								indexSchema.DeployCommandGroupKind: false,
								indexSchema.BuildCommandGroupKind:  true,
								indexSchema.RunCommandGroupKind:    true,
								indexSchema.TestCommandGroupKind:   false,
							},
						},
					},
				},
				{
					Name: "devfileB",
					CommandGroups: map[indexSchema.CommandGroupKind]bool{
						indexSchema.BuildCommandGroupKind: true,
						indexSchema.RunCommandGroupKind:   true,
					},
				},
				{
					Name: "devfileC",
					CommandGroups: map[indexSchema.CommandGroupKind]bool{
						indexSchema.DeployCommandGroupKind: false,
						indexSchema.RunCommandGroupKind:    true,
					},
					Versions: []indexSchema.Version{
						{
							Version: "2.0.0",
							CommandGroups: map[indexSchema.CommandGroupKind]bool{
								indexSchema.DebugCommandGroupKind:  false,
								indexSchema.DeployCommandGroupKind: false,
								indexSchema.BuildCommandGroupKind:  true,
								indexSchema.RunCommandGroupKind:    true,
								indexSchema.TestCommandGroupKind:   false,
							},
						},
					},
				},
			},
		},
	}
	filterDeploymentScopesTestCases = []filterDevfileStrArrayFieldTestCase{
		{
			Name:      "innerloop filters v2 index",
			FieldName: ArrayParamDeploymentScopes,
			Index: []indexSchema.Schema{
				{
					Name: "devfileA",
					DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
						indexSchema.InnerloopKind: true,
						indexSchema.OuterloopKind: true,
					},
				},
				{
					Name: "devfileB",
					Versions: []indexSchema.Version{
						{
							Version: "2.0.0",
							DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
								indexSchema.InnerloopKind: true,
								indexSchema.OuterloopKind: false,
							},
						},
					},
				},
				{
					Name: "devfileC",
					DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
						indexSchema.InnerloopKind: true,
					},
					Versions: []indexSchema.Version{
						{
							Version: "1.0.0",
							DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
								indexSchema.InnerloopKind: true,
								indexSchema.OuterloopKind: true,
							},
						},
					},
				},
				{
					Name: "devfileD",
					DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
						indexSchema.OuterloopKind: true,
					},
				},
				{
					Name: "devfileE",
				},
			},
			V1Index: false,
			Values:  []string{string(indexSchema.InnerloopKind)},
			WantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
					DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
						indexSchema.InnerloopKind: true,
						indexSchema.OuterloopKind: true,
					},
				},
				{
					Name: "devfileB",
					Versions: []indexSchema.Version{
						{
							Version: "2.0.0",
							DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
								indexSchema.InnerloopKind: true,
								indexSchema.OuterloopKind: false,
							},
						},
					},
				},
				{
					Name: "devfileC",
					DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
						indexSchema.InnerloopKind: true,
					},
					Versions: []indexSchema.Version{
						{
							Version: "1.0.0",
							DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
								indexSchema.InnerloopKind: true,
								indexSchema.OuterloopKind: true,
							},
						},
					},
				},
			},
		},
		{
			Name:      "outerloop filters v2 index",
			FieldName: ArrayParamDeploymentScopes,
			Index: []indexSchema.Schema{
				{
					Name: "devfileA",
					DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
						indexSchema.InnerloopKind: true,
						indexSchema.OuterloopKind: true,
					},
				},
				{
					Name: "devfileB",
					Versions: []indexSchema.Version{
						{
							Version: "2.0.0",
							DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
								indexSchema.InnerloopKind: true,
								indexSchema.OuterloopKind: false,
							},
						},
					},
				},
				{
					Name: "devfileC",
					DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
						indexSchema.InnerloopKind: true,
					},
					Versions: []indexSchema.Version{
						{
							Version: "1.0.0",
							DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
								indexSchema.InnerloopKind: true,
								indexSchema.OuterloopKind: true,
							},
						},
					},
				},
				{
					Name: "devfileD",
					DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
						indexSchema.OuterloopKind: true,
					},
				},
				{
					Name: "devfileE",
				},
			},
			V1Index: false,
			Values:  []string{string(indexSchema.OuterloopKind)},
			WantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
					DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
						indexSchema.InnerloopKind: true,
						indexSchema.OuterloopKind: true,
					},
				},
				{
					Name: "devfileC",
					DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
						indexSchema.InnerloopKind: true,
					},
					Versions: []indexSchema.Version{
						{
							Version: "1.0.0",
							DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
								indexSchema.InnerloopKind: true,
								indexSchema.OuterloopKind: true,
							},
						},
					},
				},
				{
					Name: "devfileD",
					DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
						indexSchema.OuterloopKind: true,
					},
				},
			},
		},
		{
			Name:      "all deployment scopes filters v2 index",
			FieldName: ArrayParamDeploymentScopes,
			Index: []indexSchema.Schema{
				{
					Name: "devfileA",
					DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
						indexSchema.InnerloopKind: true,
						indexSchema.OuterloopKind: true,
					},
				},
				{
					Name: "devfileB",
					Versions: []indexSchema.Version{
						{
							Version: "2.0.0",
							DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
								indexSchema.InnerloopKind: true,
								indexSchema.OuterloopKind: false,
							},
						},
					},
				},
				{
					Name: "devfileC",
					DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
						indexSchema.InnerloopKind: true,
					},
					Versions: []indexSchema.Version{
						{
							Version: "1.0.0",
							DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
								indexSchema.InnerloopKind: true,
								indexSchema.OuterloopKind: true,
							},
						},
					},
				},
				{
					Name: "devfileD",
					DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
						indexSchema.OuterloopKind: true,
					},
				},
				{
					Name: "devfileE",
				},
			},
			V1Index: false,
			Values:  []string{string(indexSchema.InnerloopKind), string(indexSchema.OuterloopKind)},
			WantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
					DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
						indexSchema.InnerloopKind: true,
						indexSchema.OuterloopKind: true,
					},
				},
				{
					Name: "devfileC",
					DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
						indexSchema.InnerloopKind: true,
					},
					Versions: []indexSchema.Version{
						{
							Version: "1.0.0",
							DeploymentScopes: map[indexSchema.DeploymentScopeKind]bool{
								indexSchema.InnerloopKind: true,
								indexSchema.OuterloopKind: true,
							},
						},
					},
				},
			},
		},
	}
	filterGitRemoteNamesTestCases = []filterDevfileStrArrayFieldTestCase{
		{
			Name:      "two git remote name filters",
			FieldName: ArrayParamGitRemoteNames,
			Index: []indexSchema.Schema{
				{
					Name: "devfileA",
					Git: &indexSchema.Git{
						Remotes: map[string]string{
							"linkA": "git.test.com",
							"linkB": "https://http.test.com",
							"linkC": "https://another.testlink.ca",
						},
					},
				},
				{
					Name: "devfileB",
					Git: &indexSchema.Git{
						Remotes: map[string]string{
							"linkA": "git.test.com",
							"linkC": "https://another.testlink.ca",
						},
					},
				},
				{
					Name: "devfileC",
					Git: &indexSchema.Git{
						Remotes: map[string]string{
							"linkA": "git.test.com",
						},
					},
				},
				{
					Name: "devfileD",
					Git: &indexSchema.Git{
						RemoteName: "remoteA",
					},
				},
				{
					Name: "devfileE",
				},
			},
			V1Index: true,
			Values:  []string{"linkA", "linkC"},
			WantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
					Git: &indexSchema.Git{
						Remotes: map[string]string{
							"linkA": "git.test.com",
							"linkB": "https://http.test.com",
							"linkC": "https://another.testlink.ca",
						},
					},
				},
				{
					Name: "devfileB",
					Git: &indexSchema.Git{
						Remotes: map[string]string{
							"linkA": "git.test.com",
							"linkC": "https://another.testlink.ca",
						},
					},
				},
			},
		},
		{
			Name:      "two git remote name filters with v2 index",
			FieldName: ArrayParamGitRemoteNames,
			Index: []indexSchema.Schema{
				{
					Name: "devfileA",
					Versions: []indexSchema.Version{
						{
							Version: "1.0.0",
						},
						{
							Version: "1.1.0",
							Git: &indexSchema.Git{
								Remotes: map[string]string{
									"linkA": "git.test.com",
									"linkB": "https://http.test.com",
									"linkC": "https://another.testlink.ca",
								},
							},
						},
					},
				},
				{
					Name: "devfileB",
					Git: &indexSchema.Git{
						Remotes: map[string]string{
							"linkA": "git.test.com",
							"linkC": "https://another.testlink.ca",
						},
					},
				},
				{
					Name: "devfileC",
					Git: &indexSchema.Git{
						Remotes: map[string]string{
							"linkA": "git.test.com",
						},
					},
					Versions: []indexSchema.Version{
						{
							Version: "1.0.0",
							Git: &indexSchema.Git{
								Remotes: map[string]string{
									"linkA": "git.test.com",
								},
							},
						},
						{
							Version: "1.1.0",
							Git: &indexSchema.Git{
								Remotes: map[string]string{
									"linkA": "git.test.com",
									"linkC": "https://another.testlink.ca",
								},
							},
						},
					},
				},
				{
					Name: "devfileD",
					Git: &indexSchema.Git{
						RemoteName: "remoteA",
					},
				},
				{
					Name: "devfileE",
				},
			},
			V1Index: false,
			Values:  []string{"linkA", "linkC"},
			WantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
					Versions: []indexSchema.Version{
						{
							Version: "1.1.0",
							Git: &indexSchema.Git{
								Remotes: map[string]string{
									"linkA": "git.test.com",
									"linkB": "https://http.test.com",
									"linkC": "https://another.testlink.ca",
								},
							},
						},
					},
				},
				{
					Name: "devfileB",
					Git: &indexSchema.Git{
						Remotes: map[string]string{
							"linkA": "git.test.com",
							"linkC": "https://another.testlink.ca",
						},
					},
				},
				{
					Name: "devfileC",
					Git: &indexSchema.Git{
						Remotes: map[string]string{
							"linkA": "git.test.com",
						},
					},
					Versions: []indexSchema.Version{
						{
							Version: "1.1.0",
							Git: &indexSchema.Git{
								Remotes: map[string]string{
									"linkA": "git.test.com",
									"linkC": "https://another.testlink.ca",
								},
							},
						},
					},
				},
			},
		},
	}
	filterGitRemotesTestCases = []filterDevfileStrArrayFieldTestCase{
		{
			Name:      "two git remote filters",
			FieldName: ArrayParamGitRemotes,
			Index: []indexSchema.Schema{
				{
					Name: "devfileA",
					Git: &indexSchema.Git{
						Remotes: map[string]string{
							"linkA": "git.test.com",
							"linkB": "https://http.test.com",
							"linkC": "https://another.testlink.ca",
						},
					},
				},
				{
					Name: "devfileB",
					Git: &indexSchema.Git{
						Remotes: map[string]string{
							"linkA": "git.test.com",
							"linkC": "https://another.testlink.ca",
						},
					},
				},
				{
					Name: "devfileC",
					Git: &indexSchema.Git{
						Remotes: map[string]string{
							"linkA": "git.test.com",
						},
					},
				},
				{
					Name: "devfileD",
					Git: &indexSchema.Git{
						RemoteName: "remoteA",
					},
				},
				{
					Name: "devfileE",
				},
			},
			V1Index: true,
			Values:  []string{"git", ".com"},
			WantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
					Git: &indexSchema.Git{
						Remotes: map[string]string{
							"linkA": "git.test.com",
							"linkB": "https://http.test.com",
							"linkC": "https://another.testlink.ca",
						},
					},
				},
				{
					Name: "devfileB",
					Git: &indexSchema.Git{
						Remotes: map[string]string{
							"linkA": "git.test.com",
							"linkC": "https://another.testlink.ca",
						},
					},
				},
				{
					Name: "devfileC",
					Git: &indexSchema.Git{
						Remotes: map[string]string{
							"linkA": "git.test.com",
						},
					},
				},
			},
		},
		{
			Name:      "two git remote filters with v2 index",
			FieldName: ArrayParamGitRemotes,
			Index: []indexSchema.Schema{
				{
					Name: "devfileA",
					Versions: []indexSchema.Version{
						{
							Version: "1.0.0",
						},
						{
							Version: "1.1.0",
							Git: &indexSchema.Git{
								Remotes: map[string]string{
									"linkA": "git.test.com",
									"linkB": "https://http.test.com",
									"linkC": "https://another.testlink.ca",
								},
							},
						},
					},
				},
				{
					Name: "devfileB",
					Git: &indexSchema.Git{
						Remotes: map[string]string{
							"linkA": "git.test.com",
							"linkC": "https://another.testlink.ca",
						},
					},
				},
				{
					Name: "devfileC",
					Git: &indexSchema.Git{
						Remotes: map[string]string{
							"linkA": "git.test.com",
						},
					},
					Versions: []indexSchema.Version{
						{
							Version: "1.0.0",
							Git: &indexSchema.Git{
								Remotes: map[string]string{
									"linkA": "git.test.com",
								},
							},
						},
						{
							Version: "1.1.0",
							Git: &indexSchema.Git{
								Remotes: map[string]string{
									"linkA": "git.test.com",
									"linkC": "https://another.testlink.ca",
								},
							},
						},
					},
				},
				{
					Name: "devfileD",
					Git: &indexSchema.Git{
						RemoteName: "remoteA",
					},
				},
				{
					Name: "devfileE",
				},
			},
			V1Index: false,
			Values:  []string{"git", ".com"},
			WantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
					Versions: []indexSchema.Version{
						{
							Version: "1.1.0",
							Git: &indexSchema.Git{
								Remotes: map[string]string{
									"linkA": "git.test.com",
									"linkB": "https://http.test.com",
									"linkC": "https://another.testlink.ca",
								},
							},
						},
					},
				},
				{
					Name: "devfileB",
					Git: &indexSchema.Git{
						Remotes: map[string]string{
							"linkA": "git.test.com",
							"linkC": "https://another.testlink.ca",
						},
					},
				},
				{
					Name: "devfileC",
					Git: &indexSchema.Git{
						Remotes: map[string]string{
							"linkA": "git.test.com",
						},
					},
					Versions: []indexSchema.Version{
						{
							Version: "1.0.0",
							Git: &indexSchema.Git{
								Remotes: map[string]string{
									"linkA": "git.test.com",
								},
							},
						},
						{
							Version: "1.1.0",
							Git: &indexSchema.Git{
								Remotes: map[string]string{
									"linkA": "git.test.com",
									"linkC": "https://another.testlink.ca",
								},
							},
						},
					},
				},
			},
		},
	}
	// ======================================
	// Filter Devfile String Field Test Cases
	// ======================================
	filterNameFieldTestCases = []filterDevfileStrFieldTestCase{
		{
			Name:      "name filter",
			FieldName: ParamName,
			Index: []indexSchema.Schema{
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
			V1Index: true,
			Value:   "A",
			WantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
				},
				{
					Name: "devfileAA",
				},
			},
			WantErr: false,
		},
		{
			Name:      "name filter v2",
			FieldName: ParamName,
			Index: []indexSchema.Schema{
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
			Value: "A",
			WantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
				},
				{
					Name: "devfileAA",
				},
			},
			WantErr: false,
		},
	}
	filterDisplayNameFieldTestCases = []filterDevfileStrFieldTestCase{
		{
			Name:      "display name filter",
			FieldName: ParamDisplayName,
			Index: []indexSchema.Schema{
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
			V1Index: true,
			Value:   "Flask",
			WantIndex: []indexSchema.Schema{
				{
					Name:        "devfileC",
					DisplayName: "Flask",
				},
				{
					Name:        "devfileAA",
					DisplayName: "Python - Flask",
				},
			},
			WantErr: false,
		},
		{
			Name:      "display name filter v2",
			FieldName: ParamDisplayName,
			Index: []indexSchema.Schema{
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
			Value: "Flask",
			WantIndex: []indexSchema.Schema{
				{
					Name:        "devfileC",
					DisplayName: "Flask",
				},
				{
					Name:        "devfileAA",
					DisplayName: "Python - Flask",
				},
			},
			WantErr: false,
		},
	}
	filterDescriptionFieldTestCases = []filterDevfileStrFieldTestCase{
		{
			Name:      "description filter",
			FieldName: ParamDescription,
			Index: []indexSchema.Schema{
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
			V1Index: true,
			Value:   "stack",
			WantIndex: []indexSchema.Schema{
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
			WantErr: false,
		},
		{
			Name:      "description filter v2",
			FieldName: ParamDescription,
			Index: []indexSchema.Schema{
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
			Value: "stack",
			WantIndex: []indexSchema.Schema{
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
			WantErr: false,
		},
	}
	filterIconFieldTestCases = []filterDevfileStrFieldTestCase{
		{
			Name:      "icon filter",
			FieldName: ParamIcon,
			Index: []indexSchema.Schema{
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
			V1Index: true,
			Value:   "png",
			WantIndex: []indexSchema.Schema{
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
			WantErr: false,
		},
		{
			Name:      "icon filter v2",
			FieldName: ParamIcon,
			Index: []indexSchema.Schema{
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
			Value: "ico",
			WantIndex: []indexSchema.Schema{
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
			WantErr: false,
		},
	}
	filterProjectTypeFieldTestCases = []filterDevfileStrFieldTestCase{
		{
			Name:      "project type filter",
			FieldName: ParamProjectType,
			Index: []indexSchema.Schema{
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
			V1Index: true,
			Value:   "java",
			WantIndex: []indexSchema.Schema{
				{
					Name:        "devfileD",
					DisplayName: "Java Springboot",
					ProjectType: "java",
				},
			},
			WantErr: false,
		},
		{
			Name:      "project type filter v2",
			FieldName: ParamProjectType,
			Index: []indexSchema.Schema{
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
			Value: "java",
			WantIndex: []indexSchema.Schema{
				{
					Name:        "devfileD",
					DisplayName: "Java Springboot",
					ProjectType: "java",
				},
			},
			WantErr: false,
		},
	}
	filterLanguageFieldTestCases = []filterDevfileStrFieldTestCase{
		{
			Name:      "language filter",
			FieldName: ParamLanguage,
			Index: []indexSchema.Schema{
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
			V1Index: true,
			Value:   "java",
			WantIndex: []indexSchema.Schema{
				{
					Name:        "devfileD",
					DisplayName: "Java Springboot",
					Language:    "Java",
				},
			},
			WantErr: false,
		},
	}
	filterVersionFieldTestCases       = []filterDevfileStrFieldTestCase{}
	filterSchemaVersionFieldTestCases = []filterDevfileStrFieldTestCase{}
	filterDefaultFieldTestCases       = []filterDevfileStrFieldTestCase{}
	filterGitUrlFieldTestCases        = []filterDevfileStrFieldTestCase{}
	filterGitRemoteNameFieldTestCases = []filterDevfileStrFieldTestCase{}
	filterGitSubDirFieldTestCases     = []filterDevfileStrFieldTestCase{}
	filterGitRevisionFieldTestCases   = []filterDevfileStrFieldTestCase{}
	filterProviderFieldTestCases      = []filterDevfileStrFieldTestCase{}
	filterSupportUrlFieldTestCases    = []filterDevfileStrFieldTestCase{}
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
			gotIndex, gotErr := FilterDevfileSchemaVersion(test.index, &test.minSchemaVersion, &test.maxSchemaVersion)
			if gotErr != nil {
				if gotIndex != nil {
					t.Errorf("Unexpected non-nil index on error: %v", gotIndex)
				}
				t.Errorf("Unexpected error: %v", gotErr)
			} else if !reflect.DeepEqual(gotIndex, test.wantIndex) {
				t.Errorf("Got: %v, Expected: %v", gotIndex, test.wantIndex)
			}
		})
	}
}

func TestFilterDevfileVersion(t *testing.T) {

	tests := []struct {
		name       string
		index      []indexSchema.Schema
		minVersion string
		maxVersion string
		wantIndex  []indexSchema.Schema
	}{
		{
			name: "only minVersion",
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
			minVersion: "1.1",
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
			name: "only maxVersion",
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
			maxVersion: "1.1",
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
			minVersion: "1.1",
			maxVersion: "1.2",
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
			gotIndex, gotErr := FilterDevfileVersion(test.index, &test.minVersion, &test.maxVersion)
			if gotErr != nil {
				if gotIndex != nil {
					t.Errorf("Unexpected non-nil index on error: %v", gotIndex)
				}
				t.Errorf("Unexpected error: %v", gotErr)
			} else if !reflect.DeepEqual(gotIndex, test.wantIndex) {
				t.Errorf("Got: %v, Expected: %v", gotIndex, test.wantIndex)
			}
		})
	}
}

func TestFilterDevfileDeprecated(t *testing.T) {
	tests := []struct {
		name       string
		index      []indexSchema.Schema
		deprecated bool
		v1Index    bool
		wantIndex  []indexSchema.Schema
	}{
		{
			name: "Case 1: filter out non-deprecated stacks",
			index: []indexSchema.Schema{
				{
					Name: "devfileA",
					Tags: []string{
						"Go",
						"Gin",
						"Deprecated",
					},
				},
				{
					Name: "devfileB",
					Tags: []string{
						"Deprecated",
						"Python",
					},
				},
				{
					Name: "devfileC",
					Tags: []string{
						"Python",
						"Flask",
					},
				},
				{
					Name: "devfileD",
					Tags: []string{
						"JS",
						"Node.js",
					},
				},
			},
			deprecated: true,
			v1Index:    true,
			wantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
					Tags: []string{
						"Go",
						"Gin",
						"Deprecated",
					},
				},
				{
					Name: "devfileB",
					Tags: []string{
						"Deprecated",
						"Python",
					},
				},
			},
		},
		{
			name: "Case 2: filter out non-deprecated stacks v2",
			index: []indexSchema.Schema{
				{
					Name: "devfileA",
					Tags: []string{
						"Go",
						"Gin",
						"Deprecated",
					},
					Versions: []indexSchema.Version{
						{
							Version: "1.2.0",
							Default: true,
							Tags: []string{
								"Go",
								"Gin",
							},
						},
						{
							Version: "1.0.0",
							Tags: []string{
								"Go",
								"Deprecated",
							},
						},
					},
				},
				{
					Name: "devfileB",
					Tags: []string{
						"Python",
					},
					Versions: []indexSchema.Version{
						{
							Version: "1.2.0",
							Default: true,
							Tags: []string{
								"Deprecated",
								"Python",
							},
						},
					},
				},
				{
					Name: "devfileC",
					Tags: []string{
						"Python",
						"Flask",
					},
					Versions: []indexSchema.Version{
						{
							Version: "1.2.0",
							Default: true,
							Tags: []string{
								"Python",
								"Flask",
							},
						},
						{
							Version: "1.0.0",
							Tags: []string{
								"Deprecated",
								"Python",
							},
						},
					},
				},
				{
					Name: "devfileD",
					Tags: []string{
						"JS",
						"Node.js",
					},
				},
			},
			deprecated: true,
			wantIndex: []indexSchema.Schema{
				{
					Name: "devfileA",
					Tags: []string{
						"Go",
						"Gin",
						"Deprecated",
					},
					Versions: []indexSchema.Version{
						{
							Version: "1.2.0",
							Default: true,
							Tags: []string{
								"Go",
								"Gin",
							},
						},
						{
							Version: "1.0.0",
							Tags: []string{
								"Go",
								"Deprecated",
							},
						},
					},
				},
				{
					Name: "devfileB",
					Tags: []string{
						"Python",
					},
					Versions: []indexSchema.Version{
						{
							Version: "1.2.0",
							Default: true,
							Tags: []string{
								"Deprecated",
								"Python",
							},
						},
					},
				},
			},
		},
		{
			name: "Case 3: filter out non-deprecated stacks with no deprecated stacks",
			index: []indexSchema.Schema{
				{
					Name: "devfileC",
					Tags: []string{
						"Python",
						"Flask",
					},
					Versions: []indexSchema.Version{
						{
							Version: "1.2.0",
							Default: true,
							Tags: []string{
								"Python",
								"Flask",
							},
						},
						{
							Version: "1.0.0",
							Tags: []string{
								"Deprecated",
								"Python",
							},
						},
					},
				},
				{
					Name: "devfileD",
					Tags: []string{
						"JS",
						"Node.js",
					},
				},
			},
			deprecated: true,
			wantIndex:  make([]indexSchema.Schema, 0, 2),
		},
		{
			name:       "Case 4: filter out non-deprecated stacks with empty index schema",
			deprecated: true,
		},
		{
			name: "Case 5: filter out deprecated stacks",
			index: []indexSchema.Schema{
				{
					Name: "devfileA",
					Tags: []string{
						"Go",
						"Gin",
						"Deprecated",
					},
				},
				{
					Name: "devfileB",
					Tags: []string{
						"Deprecated",
						"Python",
					},
				},
				{
					Name: "devfileC",
					Tags: []string{
						"Python",
						"Flask",
					},
				},
				{
					Name: "devfileD",
					Tags: []string{
						"JS",
						"Node.js",
					},
				},
			},
			v1Index: true,
			wantIndex: []indexSchema.Schema{
				{
					Name: "devfileC",
					Tags: []string{
						"Python",
						"Flask",
					},
				},
				{
					Name: "devfileD",
					Tags: []string{
						"JS",
						"Node.js",
					},
				},
			},
		},
		{
			name: "Case 6: filter out deprecated stacks v2",
			index: []indexSchema.Schema{
				{
					Name: "devfileA",
					Tags: []string{
						"Go",
						"Gin",
						"Deprecated",
					},
					Versions: []indexSchema.Version{
						{
							Version: "1.2.0",
							Default: true,
							Tags: []string{
								"Go",
								"Gin",
							},
						},
						{
							Version: "1.0.0",
							Tags: []string{
								"Go",
								"Deprecated",
							},
						},
					},
				},
				{
					Name: "devfileB",
					Tags: []string{
						"Python",
					},
					Versions: []indexSchema.Version{
						{
							Version: "1.2.0",
							Default: true,
							Tags: []string{
								"Deprecated",
								"Python",
							},
						},
					},
				},
				{
					Name: "devfileC",
					Tags: []string{
						"Python",
						"Flask",
					},
					Versions: []indexSchema.Version{
						{
							Version: "1.2.0",
							Default: true,
							Tags: []string{
								"Python",
								"Flask",
							},
						},
						{
							Version: "1.0.0",
							Tags: []string{
								"Deprecated",
								"Python",
							},
						},
					},
				},
				{
					Name: "devfileD",
					Tags: []string{
						"JS",
						"Node.js",
					},
				},
			},
			wantIndex: []indexSchema.Schema{
				{
					Name: "devfileC",
					Tags: []string{
						"Python",
						"Flask",
					},
					Versions: []indexSchema.Version{
						{
							Version: "1.2.0",
							Default: true,
							Tags: []string{
								"Python",
								"Flask",
							},
						},
						{
							Version: "1.0.0",
							Tags: []string{
								"Deprecated",
								"Python",
							},
						},
					},
				},
				{
					Name: "devfileD",
					Tags: []string{
						"JS",
						"Node.js",
					},
				},
			},
		},
		{
			name: "Case 7: filter out deprecated stacks with no deprecated stacks",
			index: []indexSchema.Schema{
				{
					Name: "devfileC",
					Tags: []string{
						"Python",
						"Flask",
					},
					Versions: []indexSchema.Version{
						{
							Version: "1.2.0",
							Default: true,
							Tags: []string{
								"Python",
								"Flask",
							},
						},
						{
							Version: "1.0.0",
							Tags: []string{
								"Deprecated",
								"Python",
							},
						},
					},
				},
				{
					Name: "devfileD",
					Tags: []string{
						"JS",
						"Node.js",
					},
				},
			},
			wantIndex: []indexSchema.Schema{
				{
					Name: "devfileC",
					Tags: []string{
						"Python",
						"Flask",
					},
					Versions: []indexSchema.Version{
						{
							Version: "1.2.0",
							Default: true,
							Tags: []string{
								"Python",
								"Flask",
							},
						},
						{
							Version: "1.0.0",
							Tags: []string{
								"Deprecated",
								"Python",
							},
						},
					},
				},
				{
					Name: "devfileD",
					Tags: []string{
						"JS",
						"Node.js",
					},
				},
			},
		},
		{
			name: "Case 8: filter out deprecated stacks with empty index schema",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filteredIndex := deepcopy.Copy(test.index).([]indexSchema.Schema)
			FilterDevfileDeprecated(&filteredIndex, test.deprecated, test.v1Index)

			if !reflect.DeepEqual(filteredIndex, test.wantIndex) {
				t.Errorf("\nExpected: %v\nGot: %v", test.wantIndex, filteredIndex)
			}
		})
	}
}

func TestFilterDevfileStrArrayField(t *testing.T) {
	tests := []filterDevfileStrArrayFieldTestCase{}
	tests = append(tests, filterAttributeNamesTestCases...)
	tests = append(tests, filterArchitecturesTestCases...)
	tests = append(tests, filterTagsTestCases...)
	tests = append(tests, filterResourcesTestCases...)
	tests = append(tests, filterStarterProjectsTestCases...)
	tests = append(tests, filterLinksTestCases...)
	tests = append(tests, filterCommandGroupsTestCases...)
	tests = append(tests, filterDeploymentScopesTestCases...)
	tests = append(tests, filterGitRemoteNamesTestCases...)
	tests = append(tests, filterGitRemotesTestCases...)

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			gotResult := FilterDevfileStrArrayField(test.Index, test.FieldName, test.Values, test.V1Index)
			if gotResult.Error != nil {
				t.Errorf("Unexpected error: %v", gotResult.Error)
			} else if !reflect.DeepEqual(gotResult.Index, test.WantIndex) {
				t.Errorf("Got: %v, Expected: %v", gotResult.Index, test.WantIndex)
			}
		})
	}
}

func TestFilterDevfileStrField(t *testing.T) {
	tests := []filterDevfileStrFieldTestCase{}
	tests = append(tests, filterNameFieldTestCases...)
	tests = append(tests, filterDisplayNameFieldTestCases...)
	tests = append(tests, filterDescriptionFieldTestCases...)
	tests = append(tests, filterIconFieldTestCases...)
	tests = append(tests, filterProjectTypeFieldTestCases...)
	tests = append(tests, filterLanguageFieldTestCases...)
	tests = append(tests, filterVersionFieldTestCases...)
	tests = append(tests, filterSchemaVersionFieldTestCases...)
	tests = append(tests, filterDefaultFieldTestCases...)
	tests = append(tests, filterGitUrlFieldTestCases...)
	tests = append(tests, filterGitRemoteNameFieldTestCases...)
	tests = append(tests, filterGitSubDirFieldTestCases...)
	tests = append(tests, filterGitRevisionFieldTestCases...)
	tests = append(tests, filterProviderFieldTestCases...)
	tests = append(tests, filterSupportUrlFieldTestCases...)

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			gotResult := FilterDevfileStrField(test.Index, test.FieldName, test.Value, test.V1Index)
			if !test.WantErr && gotResult.Error != nil {
				t.Errorf("Unexpected error: %v", gotResult.Error)
			} else if !test.WantErr && !reflect.DeepEqual(gotResult.Index, test.WantIndex) {
				t.Errorf("Got: %v, Expected: %v", gotResult.Index, test.WantIndex)
			} else if test.WantErr && !strings.HasPrefix(gotResult.Error.Error(), test.WantErrStr) {
				t.Errorf("Got: %v, Expected: %v", gotResult.Error.Error(), test.WantErrStr)
			}
		})
	}
}

func TestAndFilter(t *testing.T) {
	tests := []struct {
		name        string
		filters     []*FilterResult
		wantIndex   []indexSchema.Schema
		wantNotEval bool
		wantErr     bool
		wantErrStr  string
	}{
		{
			name: "Test Valid And with same results",
			filters: []*FilterResult{
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
			filters: []*FilterResult{
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
			filters: []*FilterResult{
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
				},
				{
					Index: []indexSchema.Schema{
						{
							Name:        "go",
							DisplayName: "Go",
							Description: "Go Gin Project",
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
		},
		{
			name: "Test Valid And with overlapping results v2",
			filters: []*FilterResult{
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
			filters: []*FilterResult{
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
			filters:   []*FilterResult{},
			wantIndex: []indexSchema.Schema{},
		},
		{
			name: "Test Invalid And with single error",
			filters: []*FilterResult{
				{
					Error: fmt.Errorf("A test error"),
				},
			},
			wantIndex:  []indexSchema.Schema{},
			wantErr:    true,
			wantErrStr: "A test error",
		},
		{
			name: "Test Invalid And with multiple errors",
			filters: []*FilterResult{
				{
					Error: fmt.Errorf("First test error"),
				},
				{
					Error: fmt.Errorf("Second test error"),
				},
				{
					Error: fmt.Errorf("Third test error"),
				},
			},
			wantIndex:  []indexSchema.Schema{},
			wantErr:    true,
			wantErrStr: "First test error",
		},
		{
			name: "Test Invalid And with valid filters and errors",
			filters: []*FilterResult{
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
				},
				{
					Error: fmt.Errorf("First test error"),
				},
				{
					Index: []indexSchema.Schema{
						{
							Name:        "go",
							DisplayName: "Go",
							Description: "Go Gin Project",
						},
					},
				},
				{
					Error: fmt.Errorf("Second test error"),
				},
				{
					Error: fmt.Errorf("Third test error"),
				},
			},
			wantIndex:  []indexSchema.Schema{},
			wantErr:    true,
			wantErrStr: "First test error",
		},
		{
			name: "Test Invalid And with valid filters and errors v2",
			filters: []*FilterResult{
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
				},
				{
					Error: fmt.Errorf("First test error"),
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
				},
				{
					Error: fmt.Errorf("Second test error"),
				},
				{
					Error: fmt.Errorf("Third test error"),
				},
			},
			wantIndex:  []indexSchema.Schema{},
			wantErr:    true,
			wantErrStr: "First test error",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotResult := AndFilter(test.filters...)

			// sort result and expected as sorting of result is not consistent
			sort.SliceStable(gotResult.Index, func(i, j int) bool {
				return gotResult.Index[i].Name < gotResult.Index[j].Name
			})
			sort.SliceStable(test.wantIndex, func(i, j int) bool {
				return test.wantIndex[i].Name < test.wantIndex[j].Name
			})

			if test.wantErr && !strings.Contains(gotResult.Error.Error(), test.wantErrStr) {
				t.Errorf("Got: %v, Expected: %v", gotResult.Error.Error(), test.wantErrStr)
			} else if !test.wantErr && gotResult.Error != nil {
				t.Errorf("Unexpected error: %v", gotResult.Error)
			} else if !test.wantErr && !reflect.DeepEqual(gotResult.Index, test.wantIndex) {
				t.Errorf("Got: %v, Expected: %v", gotResult.Index, test.wantIndex)
			}
		})
	}
}
