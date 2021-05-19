package library

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/devfile/registry-support/index/generator/schema"
)

func TestValidateIndexComponent(t *testing.T) {
	tests := []struct {
		name           string
		indexComponent schema.Schema
		componentType  schema.DevfileType
		wantErr        bool
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
			true,
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
			true,
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
			true,
		},
		{
			"Case 4: test index component git is empty for sample component",
			schema.Schema{
				Name: "nodejs",
			},
			schema.SampleDevfileType,
			true,
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
			},
			schema.StackDevfileType,
			false,
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
			},
			schema.SampleDevfileType,
			false,
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
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := false
			err := validateIndexComponent(tt.indexComponent, tt.componentType)
			if err != nil {
				gotErr = true
			}
			if gotErr != tt.wantErr {
				t.Errorf("Got error: %t, want error: %t, function return error: %v", gotErr, tt.wantErr, err)
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
