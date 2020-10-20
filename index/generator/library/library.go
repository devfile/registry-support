package library

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/devfile/registry-support/index/generator/schema"
	"gopkg.in/yaml.v2"
)

const (
	meta    = "meta.yaml"
	devfile = "devfile.yaml"
)

// GenerateIndexStruct parses registry then generates index struct according to the schema
func GenerateIndexStruct(registryDirPath string) ([]schema.Schema, error) {
	registryDir, err := ioutil.ReadDir(registryDirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read registry directory %s: %v", registryDirPath, err)
	}

	var index []schema.Schema
	for _, devfileDir := range registryDir {
		if !devfileDir.IsDir() {
			return nil, fmt.Errorf("%s is not a directory", filepath.Join(registryDirPath, devfileDir.Name()))
		}

		metaFilePath := filepath.Join(registryDirPath, devfileDir.Name(), meta)
		bytes, err := ioutil.ReadFile(metaFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %v", metaFilePath, err)
		}
		var indexComponent schema.Schema
		err = yaml.Unmarshal(bytes, &indexComponent)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s data: %v", metaFilePath, err)
		}

		if indexComponent.Links == nil {
			indexComponent.Links = make(map[string]string)
		}
		indexComponent.Links["self"] = fmt.Sprintf("%s/%s:%s", "devfile-catalog", indexComponent.Name, "latest")

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
