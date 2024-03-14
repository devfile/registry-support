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
			FieldName: ARRAY_PARAM_ATTRIBUTE_NAMES,
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
			FieldName: ARRAY_PARAM_ATTRIBUTE_NAMES,
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
			FieldName: ARRAY_PARAM_ARCHITECTURES,
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
			FieldName: ARRAY_PARAM_ARCHITECTURES,
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
			FieldName: ARRAY_PARAM_TAGS,
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
			FieldName: ARRAY_PARAM_TAGS,
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
			FieldName: ARRAY_PARAM_RESOURCES,
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
			FieldName: ARRAY_PARAM_RESOURCES,
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
			FieldName: ARRAY_PARAM_STARTER_PROJECTS,
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
			FieldName: ARRAY_PARAM_STARTER_PROJECTS,
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
			FieldName: ARRAY_PARAM_LINKS,
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
			FieldName: ARRAY_PARAM_LINKS,
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
			FieldName: ARRAY_PARAM_COMMAND_GROUPS,
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
			FieldName: ARRAY_PARAM_COMMAND_GROUPS,
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
	filterGitRemotesTestCases = []filterDevfileStrArrayFieldTestCase{}
	// ======================================
	// Filter Devfile String Field Test Cases
	// ======================================
	filterNameFieldTestCases = []filterDevfileStrFieldTestCase{
		{
			Name:      "name filter",
			FieldName: PARAM_NAME,
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
			FieldName: PARAM_NAME,
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
			FieldName: PARAM_DISPLAY_NAME,
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
			FieldName: PARAM_DISPLAY_NAME,
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
			FieldName: PARAM_DESCRIPTION,
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
			FieldName: PARAM_DESCRIPTION,
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
			FieldName: PARAM_ICON,
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
			FieldName: PARAM_ICON,
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
			FieldName: PARAM_PROJECT_TYPE,
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
			FieldName: PARAM_PROJECT_TYPE,
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
			FieldName: PARAM_LANGUAGE,
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
	tests := []filterDevfileStrArrayFieldTestCase{}
	tests = append(tests, filterAttributeNamesTestCases...)
	tests = append(tests, filterArchitecturesTestCases...)
	tests = append(tests, filterTagsTestCases...)
	tests = append(tests, filterResourcesTestCases...)
	tests = append(tests, filterStarterProjectsTestCases...)
	tests = append(tests, filterLinksTestCases...)
	tests = append(tests, filterCommandGroupsTestCases...)
	tests = append(tests, filterGitRemotesTestCases...)

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			gotResult := FilterDevfileStrArrayField(test.Index, test.FieldName, test.Values, test.V1Index)
			gotResult.Eval()
			if !gotResult.IsEval {
				t.Errorf("Got unexpected unevaluated result: %v", gotResult)
			} else if gotResult.Error != nil {
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

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			gotResult := FilterDevfileStrField(test.Index, test.FieldName, test.Value, test.V1Index)
			gotResult.Eval()
			if !gotResult.IsEval {
				t.Errorf("Got unexpected unevaluated result: %v", gotResult)
			} else if !test.WantErr && gotResult.Error != nil {
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
