package library

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/devfile/registry-support/index/generator/schema"
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
				Resources: []string{
					"devfile.yaml",
				},
			},
			schema.StackDevfileType,
			&linkEmptyErr,
		},
		{
			"Case 3: test index component resources are empty for stack component",
			schema.Schema{
				Name: "nodejs",
				Links: map[string]string{
					"self": "devfile-catalog/java-maven:latest",
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
				Links: map[string]string{
					"self": "devfile-catalog/java-maven:latest",
				},
				Resources: []string{
					"devfile.yaml",
				},
				Architectures: []string{
					"amd64",
				},
				Provider: "Red Hat",
				SupportUrl: "http://testurl/support.md",
			},
			schema.StackDevfileType,
			nil,
		},
		{
			"Case 6: test happy path for for sample component",
			schema.Schema{
				Name: "nodejs",
				Git: &schema.Git{
					Remotes: map[string]string{
						"origin": "https://github.com/redhat-developer/devfile-sample",
					},
				},
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
				Git: &schema.Git{
					Remotes: map[string]string{
						"origin": "https://github.com/redhat-developer/devfile-sample",
					},
				},
			},
			schema.SampleDevfileType,
			&noArchErr,
		},
		{
			"Case 9: check for missing provider",
			schema.Schema{
				Name: "nodejs",
				Links: map[string]string{
					"self": "devfile-catalog/java-maven:latest",
				},
				Resources: []string{
					"devfile.yaml",
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
				Links: map[string]string{
					"self": "devfile-catalog/java-maven:latest",
				},
				Resources: []string{
					"devfile.yaml",
				},
				Architectures: []string{
					"amd64",
				},
				Provider: "Red Hat",
			},
			schema.StackDevfileType,
			&noSupportUrlErr,
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
	bytes, err := ioutil.ReadFile(wantIndexFilePath)
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
			t.Errorf("Want index %v, got index %v", wantIndex, gotIndex)
		}
	})
}

func TestParseExtraDevfileEntries(t *testing.T) {
	registryDirPath := "../tests/registry"
	wantIndexFilePath := "../tests/registry/index_extra.json"
	bytes, err := ioutil.ReadFile(wantIndexFilePath)
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
			t.Errorf("Want index %v, got index %v", wantIndex, gotIndex)
		}
	})
}

func TestGenerateIndexStruct(t *testing.T) {
	registryDirPath := "../tests/registry"
	wantIndexFilePath := "../tests/registry/index_main.json"
	bytes, err := ioutil.ReadFile(wantIndexFilePath)
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
			t.Errorf("Want index %v, got index %v", wantIndex, gotIndex)
		}
	})
}
