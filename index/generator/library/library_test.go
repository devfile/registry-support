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

package library

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfilepkg "github.com/devfile/api/v2/pkg/devfile"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	v2 "github.com/devfile/library/v2/pkg/devfile/parser/data/v2"
	"github.com/devfile/registry-support/index/generator/schema"
	"github.com/nsf/jsondiff"
	"github.com/stretchr/testify/assert"
)

func TestValidateIndexComponent(t *testing.T) {

	nameNotInitErr := ".*name is not initialized.*"
	linkEmptyErr := ".*links are empty.*"
	resourcesEmptyErr := ".*resources are empty.*"
	gitEmptyErr := ".*git is empty.*"
	multipleRemotesErr := ".*has multiple remotes.*"
	noArchErr := ".*has no architecture.*"
	noProviderErr := ".*has no provider.*"
	noSupportUrlErr := ".*has no supportUrl.*"
	versionsEmptyErr := ".*versions list is empty.*"
	noDefaultVersionErr := ".*has no default version.*"
	noVersionErr := ".*no version specified.*"
	schemaVersionEmptyErr := ".*schema version is empty.*"
	multipleVersionErr := ".*has multiple default versions.*"
	iconUrlBrokenErr := ".*has broken or not existing icon.*"
	invalidDeploymentScopeErr := ".*Deployment scope \\S+ is incorrect for devfile \\S+.*"
	tooManyDeploymentScopesErr := ".*has too many deployment scopes.*"

	tests := []struct {
		name           string
		indexComponent schema.Schema
		componentType  schema.DevfileType
		wantErr        *string
	}{
		{
			"Case 1: test index component is not initialized for stack component",
			schema.Schema{
				Links: map[string]string{
					"self": "devfile-catalog/java-maven:latest",
				},
				Resources: []string{
					"devfile.yaml",
				},
			},
			schema.StackDevfileType,
			&nameNotInitErr,
		},
		{
			"Case 2: test index component links are empty for stack component",
			schema.Schema{
				Name: "nodejs",
				Versions: []schema.Version{
					{
						Version:       "1.0.0",
						SchemaVersion: "2.0.0",
						Resources: []string{
							"devfile.yaml",
						},
					},
				},
			},
			schema.StackDevfileType,
			&linkEmptyErr,
		},
		{
			"Case 3: test index component resources are empty for stack component",
			schema.Schema{
				Name: "nodejs",
				Versions: []schema.Version{
					{
						Version:       "1.0.0",
						SchemaVersion: "2.0.0",
						Links: map[string]string{
							"self": "devfile-catalog/java-maven:latest",
						},
					},
				},
			},
			schema.StackDevfileType,
			&resourcesEmptyErr,
		},
		{
			"Case 4: test index component git is empty for sample component",
			schema.Schema{
				Name: "nodejs",
			},
			schema.SampleDevfileType,
			&gitEmptyErr,
		},
		{
			"Case 5: test happy path for for stack component",
			schema.Schema{
				Name: "nodejs",
				Icon: "https://raw.githubusercontent.com/devfile-samples/devfile-stack-icons/main/java-maven.jpg",
				Architectures: []string{
					"amd64",
				},
				Provider:   "Red Hat",
				SupportUrl: "http://testurl/support.md",
				Versions: []schema.Version{
					{
						Version:       "1.0.0",
						SchemaVersion: "2.0.0",
						Default:       true,
						Links: map[string]string{
							"self": "devfile-catalog/java-maven:1.0.0",
						},
						Resources: []string{
							"devfile.yaml",
						},
					},
					{
						Version:       "1.1.0",
						SchemaVersion: "2.1.0",
						Links: map[string]string{
							"self": "devfile-catalog/java-maven:2.1.0",
						},
						Resources: []string{
							"devfile.yaml",
						},
					},
				},
			},
			schema.StackDevfileType,
			nil,
		},
		{
			"Case 6: test happy path for for sample component with old struct",
			schema.Schema{
				Name: "nodejs",
				Icon: "https://raw.githubusercontent.com/devfile-samples/devfile-stack-icons/main/java-maven.jpg",
				Git: &schema.Git{
					Remotes: map[string]string{
						"origin": "https://github.com/redhat-developer/devfile-sample",
					},
				},
				SupportUrl: "http://testurl/support.md",
				Provider:   "Red Hat",
				Architectures: []string{
					"amd64",
				},
			},
			schema.SampleDevfileType,
			nil,
		},
		{
			"Case 7: test index component git has multiple remotes",
			schema.Schema{
				Name: "nodejs",
				Git: &schema.Git{
					Remotes: map[string]string{
						"origin": "https://github.com/redhat-developer/devfile-sample",
						"test":   "https://github.com/redhat-developer/test",
					},
				},
			},
			schema.SampleDevfileType,
			&multipleRemotesErr,
		},

		{
			"Case 8: check for missing arch",
			schema.Schema{
				Name: "nodejs",
				Icon: "https://raw.githubusercontent.com/devfile-samples/devfile-stack-icons/main/node-js.svg",
				Git: &schema.Git{
					Remotes: map[string]string{
						"origin": "https://github.com/redhat-developer/devfile-sample",
					},
				},
				SupportUrl: "http://testurl/support.md",
				Provider:   "Red Hat",
			},
			schema.SampleDevfileType,
			&noArchErr,
		},
		{
			"Case 9: check for missing provider",
			schema.Schema{
				Name: "nodejs",
				Icon: "https://raw.githubusercontent.com/devfile-samples/devfile-stack-icons/main/node-js.svg",
				Versions: []schema.Version{
					{
						Version:       "1.0.0",
						SchemaVersion: "2.0.0",
						Default:       true,
						Links: map[string]string{
							"self": "devfile-catalog/java-maven:latest",
						},
						Resources: []string{
							"devfile.yaml",
						},
					},
				},
				Architectures: []string{
					"amd64",
				},
				SupportUrl: "http://testurl/support.md",
			},
			schema.StackDevfileType,
			&noProviderErr,
		},
		{
			"Case 10: check for missing supportUrl",
			schema.Schema{
				Name: "nodejs",
				Icon: "https://raw.githubusercontent.com/devfile-samples/devfile-stack-icons/main/node-js.svg",
				Versions: []schema.Version{
					{
						Version:       "1.0.0",
						SchemaVersion: "2.0.0",
						Default:       true,
						Links: map[string]string{
							"self": "devfile-catalog/java-maven:latest",
						},
						Resources: []string{
							"devfile.yaml",
						},
					},
				},
				Architectures: []string{
					"amd64",
				},
				Provider: "Red Hat",
			},
			schema.StackDevfileType,
			&noSupportUrlErr,
		},
		{
			"Case 11: empty version list",
			schema.Schema{
				Name:     "nodejs",
				Versions: []schema.Version{},
			},
			schema.StackDevfileType,
			&versionsEmptyErr,
		},
		{
			"Case 12: test stack component missing default version",
			schema.Schema{
				Name: "nodejs",
				Architectures: []string{
					"amd64",
				},
				Provider:   "Red Hat",
				SupportUrl: "http://testurl/support.md",
				Versions: []schema.Version{
					{
						Version:       "1.0.0",
						SchemaVersion: "2.0.0",
						Links: map[string]string{
							"self": "devfile-catalog/java-maven:latest",
						},
						Resources: []string{
							"devfile.yaml",
						},
					},
				},
			},
			schema.StackDevfileType,
			&noDefaultVersionErr,
		},
		{
			"Case 13: test stack component missing version",
			schema.Schema{
				Name: "nodejs",
				Versions: []schema.Version{
					{
						SchemaVersion: "2.0.0",
						Links: map[string]string{
							"self": "devfile-catalog/java-maven:latest",
						},
						Resources: []string{
							"devfile.yaml",
						},
					},
				},
			},
			schema.StackDevfileType,
			&noVersionErr,
		},
		{
			"Case 14: test stack component missing schema version",
			schema.Schema{
				Name: "nodejs",
				Versions: []schema.Version{
					{
						Version: "1.0.0",
						Links: map[string]string{
							"self": "devfile-catalog/java-maven:latest",
						},
						Resources: []string{
							"devfile.yaml",
						},
					},
				},
			},
			schema.StackDevfileType,
			&schemaVersionEmptyErr,
		},
		{
			"Case 15: test stack component multiple default version",
			schema.Schema{
				Name: "nodejs",
				Architectures: []string{
					"amd64",
				},
				Provider:   "Red Hat",
				SupportUrl: "http://testurl/support.md",
				Versions: []schema.Version{
					{
						Version:       "1.0.0",
						SchemaVersion: "2.0.0",
						Default:       true,
						Links: map[string]string{
							"self": "devfile-catalog/java-maven:1.0.0",
						},
						Resources: []string{
							"devfile.yaml",
						},
					},
					{
						Version:       "1.1.0",
						SchemaVersion: "2.1.0",
						Default:       true,
						Links: map[string]string{
							"self": "devfile-catalog/java-maven:1.1.0",
						},
						Resources: []string{
							"devfile.yaml",
						},
					},
				},
			},
			schema.StackDevfileType,
			&multipleVersionErr,
		},
		{
			"Case 16: test happy path for for sample component with new struct",
			schema.Schema{
				Name: "nodejs",
				Icon: "https://raw.githubusercontent.com/devfile-samples/devfile-stack-icons/main/node-js.svg",
				Versions: []schema.Version{
					{
						Version:       "1.0.0",
						SchemaVersion: "2.0.0",
						Default:       true,
						Git: &schema.Git{
							Remotes: map[string]string{
								"origin": "https://github.com/redhat-developer/devfile-sample/nodejs",
							},
						},
					},
					{
						Version:       "1.1.0",
						SchemaVersion: "2.1.0",
						Git: &schema.Git{
							Remotes: map[string]string{
								"origin": "https://github.com/redhat-developer/devfile-sample/nodejs-2.1.0",
							},
						},
					},
				},
				SupportUrl: "http://testurl/support.md",
				Provider:   "Red Hat",
				Architectures: []string{
					"amd64",
				},
			},
			schema.SampleDevfileType,
			nil,
		},
		{
			"Case 17: test sample component missing default version",
			schema.Schema{
				Name: "nodejs",
				Architectures: []string{
					"amd64",
				},
				Provider:   "Red Hat",
				SupportUrl: "http://testurl/support.md",
				Versions: []schema.Version{
					{
						Version:       "1.0.0",
						SchemaVersion: "2.0.0",
						Git: &schema.Git{
							Remotes: map[string]string{
								"origin": "https://github.com/redhat-developer/devfile-sample/nodejs",
							},
						},
					},
				},
			},
			schema.SampleDevfileType,
			&noDefaultVersionErr,
		},
		{
			"Case 18: test sample component missing version",
			schema.Schema{
				Name: "nodejs",
				Versions: []schema.Version{
					{
						SchemaVersion: "2.0.0",
						Git: &schema.Git{
							Remotes: map[string]string{
								"origin": "https://github.com/redhat-developer/devfile-sample/nodejs",
							},
						},
					},
				},
			},
			schema.SampleDevfileType,
			&noVersionErr,
		},
		{
			"Case 19: test sample component missing schema version",
			schema.Schema{
				Name: "nodejs",
				Versions: []schema.Version{
					{
						Version: "1.0.0",
						Git: &schema.Git{
							Remotes: map[string]string{
								"origin": "https://github.com/redhat-developer/devfile-sample/nodejs",
							},
						},
					},
				},
			},
			schema.SampleDevfileType,
			&schemaVersionEmptyErr,
		},
		{
			"Case 20: test sample component multiple default version",
			schema.Schema{
				Name: "nodejs",
				Architectures: []string{
					"amd64",
				},
				Provider:   "Red Hat",
				SupportUrl: "http://testurl/support.md",
				Versions: []schema.Version{
					{
						Version:       "1.0.0",
						SchemaVersion: "2.0.0",
						Default:       true,
						Git: &schema.Git{
							Remotes: map[string]string{
								"origin": "https://github.com/redhat-developer/devfile-sample/nodejs",
							},
						},
					},
					{
						Version:       "1.1.0",
						SchemaVersion: "2.1.0",
						Default:       true,
						Git: &schema.Git{
							Remotes: map[string]string{
								"origin": "https://github.com/redhat-developer/devfile-sample/nodejs-2.1.0",
							},
						},
					},
				},
			},
			schema.SampleDevfileType,
			&multipleVersionErr,
		},
		{
			"Case 21: test stack component missing iconUrl",
			schema.Schema{
				Name: "nodejs",
				Versions: []schema.Version{
					{
						Version:       "1.0.0",
						SchemaVersion: "2.0.0",
						Default:       true,
						Links: map[string]string{
							"self": "devfile-catalog/java-maven:latest",
						},
						Resources: []string{
							"devfile.yaml",
						},
					},
				},
			},
			schema.StackDevfileType,
			&iconUrlBrokenErr,
		},
		{
			"Case 22: test stack component has wrong iconUrl",
			schema.Schema{
				Name: "nodejs",
				Icon: "https://github.com/",
				Versions: []schema.Version{
					{
						Version:       "1.0.0",
						SchemaVersion: "2.0.0",
						Default:       true,
						Links: map[string]string{
							"self": "devfile-catalog/java-maven:latest",
						},
						Resources: []string{
							"devfile.yaml",
						},
					},
				},
			},
			schema.StackDevfileType,
			&iconUrlBrokenErr,
		},
		{
			name: "Case 23: deployment scope has invalid kind",
			indexComponent: schema.Schema{
				Name:       "java-maven",
				Icon:       "https://raw.githubusercontent.com/devfile-samples/devfile-stack-icons/main/java-maven.jpg",
				Provider:   "Red Hat",
				SupportUrl: "https://devfile.io",
				Architectures: []string{
					"amd64",
				},
				Versions: []schema.Version{
					{
						Version:       "1.0.0",
						SchemaVersion: "2.0.0",
						Default:       true,
						Links: map[string]string{
							"self": "devfile-catalog/java-maven:latest",
						},
						Resources: []string{
							"devfile.yaml",
						},
					},
				},
				DeploymentScopes: map[schema.DeploymentScopeKind]bool{
					"foo": true,
				},
			},
			componentType: schema.StackDevfileType,
			wantErr:       &invalidDeploymentScopeErr,
		},
		{
			name: "Case 24: deployment scope has invalid kind",
			indexComponent: schema.Schema{
				Name:       "java-maven",
				Icon:       "https://raw.githubusercontent.com/devfile-samples/devfile-stack-icons/main/java-maven.jpg",
				Provider:   "Red Hat",
				SupportUrl: "https://devfile.io",
				Architectures: []string{
					"amd64",
				},
				Versions: []schema.Version{
					{
						Version:       "1.0.0",
						SchemaVersion: "2.0.0",
						Default:       true,
						Links: map[string]string{
							"self": "devfile-catalog/java-maven:latest",
						},
						Resources: []string{
							"devfile.yaml",
						},
					},
				},
				DeploymentScopes: map[schema.DeploymentScopeKind]bool{
					schema.InnerloopKind: true,
					schema.OuterloopKind: false,
					"foo":                false,
				},
			},
			componentType: schema.StackDevfileType,
			wantErr:       &tooManyDeploymentScopesErr,
		},
		{
			name: "Case 25: version deployment scope has invalid kind",
			indexComponent: schema.Schema{
				Name:       "java-maven",
				Icon:       "https://raw.githubusercontent.com/devfile-samples/devfile-stack-icons/main/java-maven.jpg",
				Provider:   "Red Hat",
				SupportUrl: "https://devfile.io",
				Architectures: []string{
					"amd64",
				},
				Versions: []schema.Version{
					{
						Version:       "1.0.0",
						SchemaVersion: "2.0.0",
						Default:       true,
						Links: map[string]string{
							"self": "devfile-catalog/java-maven:latest",
						},
						Resources: []string{
							"devfile.yaml",
						},
						DeploymentScopes: map[schema.DeploymentScopeKind]bool{
							"foo": true,
						},
					},
				},
			},
			componentType: schema.StackDevfileType,
			wantErr:       &invalidDeploymentScopeErr,
		},
		{
			name: "Case 26: version deployment scope has invalid kind",
			indexComponent: schema.Schema{
				Name:       "java-maven",
				Icon:       "https://raw.githubusercontent.com/devfile-samples/devfile-stack-icons/main/java-maven.jpg",
				Provider:   "Red Hat",
				SupportUrl: "https://devfile.io",
				Architectures: []string{
					"amd64",
				},
				Versions: []schema.Version{
					{
						Version:       "1.0.0",
						SchemaVersion: "2.0.0",
						Default:       true,
						Links: map[string]string{
							"self": "devfile-catalog/java-maven:latest",
						},
						Resources: []string{
							"devfile.yaml",
						},
						DeploymentScopes: map[schema.DeploymentScopeKind]bool{
							schema.InnerloopKind: true,
							schema.OuterloopKind: false,
							"foo":                false,
						},
					},
				},
			},
			componentType: schema.StackDevfileType,
			wantErr:       &tooManyDeploymentScopesErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIndexComponent(tt.indexComponent, tt.componentType)
			if tt.wantErr != nil && assert.Error(t, err) {
				assert.Regexp(t, *tt.wantErr, err.Error(), "Error message should match")
			} else {
				assert.NoError(t, err, "Expected error to be nil")
			}
		})
	}
}

func TestParseDevfileRegistry(t *testing.T) {
	registryDirPath := "../tests/registry"
	wantIndexFilePath := "../tests/registry/index_registry.json"
	bytes, err := os.ReadFile(wantIndexFilePath)
	if err != nil {
		t.Errorf("Failed to read index_registry.json: %v", err)
	}
	var wantIndex []schema.Schema
	err = json.Unmarshal(bytes, &wantIndex)
	if err != nil {
		t.Errorf("Failed to unmarshal index_registry.json")
	}

	t.Run("Test parse devfile registry", func(t *testing.T) {
		gotIndex, err := parseDevfileRegistry(registryDirPath, false)
		if err != nil {
			t.Errorf("Failed to call function parseDevfileRegistry: %v", err)
		}
		if !reflect.DeepEqual(wantIndex, gotIndex) {
			bWantIndex, _ := json.Marshal(wantIndex)
			bGotIndex, _ := json.Marshal(gotIndex)

			options := jsondiff.DefaultConsoleOptions()
			options.SkipMatches = true

			diff, str := jsondiff.Compare(bWantIndex, bGotIndex, &options)

			t.Errorf("Difference type %v, diff %v", diff, str)
		}
	})
}

func TestParseExtraDevfileEntries(t *testing.T) {
	registryDirPath := "../tests/registry"
	wantIndexFilePath := "../tests/registry/index_extra.json"
	bytes, err := os.ReadFile(wantIndexFilePath)
	if err != nil {
		t.Errorf("Failed to read index_extra.json: %v", err)
	}
	var wantIndex []schema.Schema
	err = json.Unmarshal(bytes, &wantIndex)
	if err != nil {
		t.Errorf("Failed to unmarshal index_extra.json")
	}

	t.Run("Test parse extra devfile entries", func(t *testing.T) {
		gotIndex, err := parseExtraDevfileEntries(registryDirPath, false)
		if err != nil {
			t.Errorf("Failed to call function parseExtraDevfileEntries: %v", err)
		}
		if !reflect.DeepEqual(wantIndex, gotIndex) {
			bWantIndex, _ := json.Marshal(wantIndex)
			bGotIndex, _ := json.Marshal(gotIndex)

			options := jsondiff.DefaultConsoleOptions()
			options.SkipMatches = true

			diff, str := jsondiff.Compare(bWantIndex, bGotIndex, &options)

			t.Errorf("Difference type %v, diff %v", diff, str)
		}
	})
}

func TestGenerateIndexStruct(t *testing.T) {
	registryDirPath := "../tests/registry"
	wantIndexFilePath := "../tests/registry/index_main.json"
	bytes, err := os.ReadFile(wantIndexFilePath)
	if err != nil {
		t.Errorf("Failed to read index_main.json: %v", err)
	}
	var wantIndex []schema.Schema
	err = json.Unmarshal(bytes, &wantIndex)
	if err != nil {
		t.Errorf("Failed to unmarshal index_main.json")
	}

	t.Run("Test generate index", func(t *testing.T) {
		gotIndex, err := GenerateIndexStruct(registryDirPath, false)
		if err != nil {
			t.Errorf("Failed to call function GenerateIndexStruct: %v", err)
		}
		if !reflect.DeepEqual(wantIndex, gotIndex) {
			bWantIndex, _ := json.Marshal(wantIndex)
			bGotIndex, _ := json.Marshal(gotIndex)

			options := jsondiff.DefaultConsoleOptions()
			options.SkipMatches = true

			diff, str := jsondiff.Compare(bWantIndex, bGotIndex, &options)

			t.Errorf("Difference type %v, diff %v", diff, str)
		}
	})
}

func TestCheckForRequiredMetadata(t *testing.T) {
	noNameError := fmt.Errorf("metadata.name is not set")
	noDisplayNameError := fmt.Errorf("metadata.displayName is not set")
	noLanguageError := fmt.Errorf("metadata.language is not set")
	noProjectTypeError := fmt.Errorf("metadata.projectType is not set")

	tests := []struct {
		name       string
		devfileObj parser.DevfileObj
		wantErr    []error
	}{
		{
			name: "No missing metadata",
			devfileObj: parser.DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1alpha2.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							Metadata: devfilepkg.DevfileMetadata{
								Name:        "java-maven",
								DisplayName: "Java Maven Stack",
								Language:    "Java",
								ProjectType: "Maven",
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "Simple devfile, missing name",
			devfileObj: parser.DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1alpha2.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							Metadata: devfilepkg.DevfileMetadata{
								DisplayName: "Java Maven Stack",
								Language:    "Java",
								ProjectType: "Maven",
							},
						},
					},
				},
			},
			wantErr: []error{noNameError},
		},
		{
			name: "Simple devfile, missing language and project",
			devfileObj: parser.DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1alpha2.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							Metadata: devfilepkg.DevfileMetadata{
								Name:        "java-maven",
								DisplayName: "Java Maven Stack",
							},
						},
					},
				},
			},
			wantErr: []error{noLanguageError, noProjectTypeError},
		},
		{
			name: "Devfile, no metadata set",
			devfileObj: parser.DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1alpha2.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							Metadata: devfilepkg.DevfileMetadata{},
						},
					},
				},
			},
			wantErr: []error{noNameError, noDisplayNameError, noLanguageError, noProjectTypeError},
		},
	}

	for _, tt := range tests {
		metadataValidateErr := checkForRequiredMetadata(tt.devfileObj)
		if !reflect.DeepEqual(tt.wantErr, metadataValidateErr) {
			t.Errorf("TestCheckForRequiredMetadata Error: Want %v, got %v", tt.wantErr, metadataValidateErr)
		}
	}
}

func TestSetLastModifiedValue(t *testing.T) {
	testRegistryDirPath := "../tests/registry"
	index := []schema.Schema{
		schema.Schema{
			Name:        "go",
			DisplayName: "Go Runtime",
			Description: "Stack with the latest Go version",
			Type:        "stack",
			Tags:        []string{"Go"},
			Icon:        "https://raw.githubusercontent.com/devfile-samples/devfile-stack-icons/main/golang.svg",
			ProjectType: "go",
			Language:    "go",
			Provider:    "Red Hat",
			Versions: []schema.Version{
				schema.Version{
					Version:         "1.2.0",
					SchemaVersion:   "2.1.0",
					Description:     "Stack with the latest Go version with devfile v2.1.0 schema version",
					Icon:            "https://raw.githubusercontent.com/devfile-samples/devfile-stack-icons/main/golang.svg",
					Tags:            []string{"testtag"},
					Links:           map[string]string{"self": "devfile-catalog/go:1.2.0"},
					Resources:       []string{"devfile.yaml"},
					StarterProjects: []string{"go-starter"},
					CommandGroups:   map[schema.CommandGroupKind]bool{"build": true, "debug": false, "deploy": false, "run": true, "test": false},
				},
				schema.Version{
					Version:         "1.1.0",
					SchemaVersion:   "2.0.0",
					Description:     "Stack with the latest Go version with devfile v2.0.0 schema version",
					Icon:            "https://raw.githubusercontent.com/devfile-samples/devfile-stack-icons/main/golang.svg",
					Tags:            []string{"Go"},
					Links:           map[string]string{"self": "devfile-catalog/go:1.1.0"},
					Resources:       []string{"devfile.yaml"},
					StarterProjects: []string{"go-starter"},
					CommandGroups:   map[schema.CommandGroupKind]bool{"build": true, "debug": false, "deploy": false, "run": true, "test": false},
				},
			},
		},
		schema.Schema{
			Name:        "code-with-quarkus",
			DisplayName: "Basic Quarkus",
			Description: "A simple Hello World Java application using Quarkus",
			Type:        "sample",
			Tags:        []string{"Java", "Quarkus"},
			Icon:        "https://raw.githubusercontent.com/elsony/devfile-sample-code-with-quarkus/main/.devfile/icon/quarkus.png",
			ProjectType: "quarkus",
			Language:    "java",
			Provider:    "Red Hat",
			Versions: []schema.Version{
				schema.Version{
					Version:       "1.1.0",
					SchemaVersion: "2.0.0",
					Description:   "java quarkus with devfile v2.0.0",
					Default:       true,
					Git: &schema.Git{
						Remotes: map[string]string{"origin": "https://github.com/elsony/devfile-sample-code-with-quarkus.git"},
					},
				},
			},
		},
	}

	wantIndex := []schema.Schema{
		schema.Schema{
			Name:         "go",
			DisplayName:  "Go Runtime",
			Description:  "Stack with the latest Go version",
			Type:         "stack",
			Tags:         []string{"Go"},
			Icon:         "https://raw.githubusercontent.com/devfile-samples/devfile-stack-icons/main/golang.svg",
			ProjectType:  "go",
			Language:     "go",
			Provider:     "Red Hat",
			LastModified: "2023-11-08T12:54:08Z",
			Versions: []schema.Version{
				schema.Version{
					Version:         "1.2.0",
					SchemaVersion:   "2.1.0",
					Description:     "Stack with the latest Go version with devfile v2.1.0 schema version",
					Icon:            "https://raw.githubusercontent.com/devfile-samples/devfile-stack-icons/main/golang.svg",
					Tags:            []string{"testtag"},
					Links:           map[string]string{"self": "devfile-catalog/go:1.2.0"},
					Resources:       []string{"devfile.yaml"},
					StarterProjects: []string{"go-starter"},
					CommandGroups:   map[schema.CommandGroupKind]bool{"build": true, "debug": false, "deploy": false, "run": true, "test": false},
					LastModified:    "2023-04-08T11:51:08Z",
				},
				schema.Version{
					Version:         "1.1.0",
					SchemaVersion:   "2.0.0",
					Description:     "Stack with the latest Go version with devfile v2.0.0 schema version",
					Icon:            "https://raw.githubusercontent.com/devfile-samples/devfile-stack-icons/main/golang.svg",
					Tags:            []string{"Go"},
					Links:           map[string]string{"self": "devfile-catalog/go:1.1.0"},
					Resources:       []string{"devfile.yaml"},
					StarterProjects: []string{"go-starter"},
					CommandGroups:   map[schema.CommandGroupKind]bool{"build": true, "debug": false, "deploy": false, "run": true, "test": false},
					LastModified:    "2023-11-08T12:54:08Z",
				},
			},
		},
		schema.Schema{
			Name:         "code-with-quarkus",
			DisplayName:  "Basic Quarkus",
			Description:  "A simple Hello World Java application using Quarkus",
			Type:         "sample",
			Tags:         []string{"Java", "Quarkus"},
			Icon:         "https://raw.githubusercontent.com/elsony/devfile-sample-code-with-quarkus/main/.devfile/icon/quarkus.png",
			ProjectType:  "quarkus",
			Language:     "java",
			Provider:     "Red Hat",
			LastModified: "2024-04-19T11:45:48+01:00",
			Versions: []schema.Version{
				schema.Version{
					Version:       "1.1.0",
					SchemaVersion: "2.0.0",
					Description:   "java quarkus with devfile v2.0.0",
					Default:       true,
					Git: &schema.Git{
						Remotes: map[string]string{"origin": "https://github.com/elsony/devfile-sample-code-with-quarkus.git"},
					},
					LastModified: "2024-04-19T11:45:48+01:00",
				},
			},
		},
	}

	tests := []struct {
		name      string
		index     []schema.Schema
		wantIndex []schema.Schema
	}{
		{
			name:      "Test SetLastModifiedValue",
			index:     index,
			wantIndex: wantIndex,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIndex, err := SetLastModifiedValue(index, testRegistryDirPath)
			if err != nil {
				t.Errorf("Failed to set last modified value: %v", err)
			}
			if !reflect.DeepEqual(tt.wantIndex, gotIndex) {
				t.Errorf("Generated index does not match")
			}
		})
	}
}
