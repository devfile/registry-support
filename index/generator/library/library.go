package library

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	devfileParser "github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/registry-support/index/generator/schema"
	"gopkg.in/yaml.v2"
)

const (
	devfile       = "devfile.yaml"
	devfileHidden = ".devfile.yaml"
)

// GenerateIndexStruct parses registry then generates index struct according to the schema
func GenerateIndexStruct(registryDirPath string, force bool) ([]schema.Schema, error) {
	registryDir, err := ioutil.ReadDir(registryDirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read registry directory %s: %v", registryDirPath, err)
	}

	var index []schema.Schema
	for _, devfileDir := range registryDir {
		if !devfileDir.IsDir() {
			continue
		}

		// Allow devfile.yaml or .devfile.yaml
		devfilePath := filepath.Join(registryDirPath, devfileDir.Name(), devfile)
		devfileHiddenPath := filepath.Join(registryDirPath, devfileDir.Name(), devfileHidden)
		if fileExists(devfilePath) && fileExists(devfileHiddenPath) {
			return nil, fmt.Errorf("both %s and %s exist", devfilePath, devfileHiddenPath)
		}
		if fileExists(devfileHiddenPath) {
			devfilePath = devfileHiddenPath
		}

		if !force {
			// Devfile validation
			_, err := devfileParser.Parse(devfilePath)
			if err != nil {
				return nil, fmt.Errorf("%s devfile is not valid: %v", devfileDir.Name(), err)
			}
		}

		bytes, err := ioutil.ReadFile(devfilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %v", devfilePath, err)
		}
		var devfile schema.Devfile
		err = yaml.Unmarshal(bytes, &devfile)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s data: %v", devfilePath, err)
		}
		indexComponent := devfile.Meta
		if indexComponent.Links == nil {
			indexComponent.Links = make(map[string]string)
		}
		indexComponent.Links["self"] = fmt.Sprintf("%s/%s:%s", "devfile-catalog", indexComponent.Name, "latest")

		for _, starterProject := range devfile.StarterProjects {
			indexComponent.StarterProjects = append(indexComponent.StarterProjects, starterProject.Name)
		}

		// Get the files in the stack folder
		stackFolder := filepath.Join(registryDirPath, devfileDir.Name())
		stackFiles, err := ioutil.ReadDir(stackFolder)
		for _, stackFile := range stackFiles {
			// The registry build should have already packaged any folders and miscellaneous files into an archive.tar file
			// But, add this check as a safeguard, as OCI doesn't support unarchived folders being pushed up.
			if !stackFile.IsDir() {
				indexComponent.Resources = append(indexComponent.Resources, stackFile.Name())
			}
		}

		if !force {
			// Index component validation
			err := validateIndexComponent(indexComponent)
			if err != nil {
				return nil, fmt.Errorf("%s index component is not valid: %v", devfileDir.Name(), err)
			}
		}

		index = append(index, indexComponent)
	}

	return index, nil
}

// CreateIndexFile creates index file in disk
func CreateIndexFile(index []schema.Schema, indexFilePath string) error {
	bytes, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal %s data: %v", indexFilePath, err)
	}

	err = ioutil.WriteFile(indexFilePath, bytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to write %s: %v", indexFilePath, err)
	}

	return nil
}

func validateIndexComponent(indexComponent schema.Schema) error {
	if indexComponent.Name == "" {
		return fmt.Errorf("index component name is not initialized")
	}
	if indexComponent.Links == nil {
		return fmt.Errorf("index component links are empty")
	}
	if indexComponent.Resources == nil {
		return fmt.Errorf("index component resources are empty")
	}

	return nil
}

func fileExists(filepath string) bool {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return false
	}

	return true
}
