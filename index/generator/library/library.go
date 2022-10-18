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

package library

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	devfileParser "github.com/devfile/library/v2/pkg/devfile"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	"github.com/devfile/registry-support/index/generator/schema"
	"gopkg.in/yaml.v2"
)

const (
	devfile             = "devfile.yaml"
	devfileHidden       = ".devfile.yaml"
	extraDevfileEntries = "extraDevfileEntries.yaml"
	stackYaml           = "stack.yaml"
)

// MissingArchError is an error if the architecture list is empty
type MissingArchError struct {
	devfile string
}

func (e *MissingArchError) Error() string {
	return fmt.Sprintf("the %s devfile has no architecture(s) mentioned\n", e.devfile)
}

// MissingProviderError is an error if the provider field is missing
type MissingProviderError struct {
	devfile string
}

func (e *MissingProviderError) Error() string {
	return fmt.Sprintf("the %s devfile has no provider mentioned\n", e.devfile)
}

// MissingSupportUrlError is an error if the supportUrl field is missing
type MissingSupportUrlError struct {
	devfile string
}

func (e *MissingSupportUrlError) Error() string {
	return fmt.Sprintf("the %s devfile has no supportUrl mentioned\n", e.devfile)
}

// GenerateIndexStruct parses registry then generates index struct according to the schema
func GenerateIndexStruct(registryDirPath string, force bool) ([]schema.Schema, error) {
	// Parse devfile registry then populate index struct
	index, err := parseDevfileRegistry(registryDirPath, force)
	if err != nil {
		return index, err
	}

	// Parse extraDevfileEntries.yaml then populate the index struct (optional)
	extraDevfileEntriesPath := path.Join(registryDirPath, extraDevfileEntries)
	if fileExists(extraDevfileEntriesPath) {
		indexFromExtraDevfileEntries, err := parseExtraDevfileEntries(registryDirPath, force)
		if err != nil {
			return index, err
		}
		index = append(index, indexFromExtraDevfileEntries...)
	}

	return index, nil
}

// CreateIndexFile creates index file in disk
func CreateIndexFile(index []schema.Schema, indexFilePath string) error {
	bytes, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal %s data: %v", indexFilePath, err)
	}

	/* #nosec G306 -- index file does not contain any sensitive data*/
	err = ioutil.WriteFile(indexFilePath, bytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to write %s: %v", indexFilePath, err)
	}

	return nil
}

func validateIndexComponent(indexComponent schema.Schema, componentType schema.DevfileType) error {
	if componentType == schema.StackDevfileType {
		if indexComponent.Name == "" {
			return fmt.Errorf("index component name is not initialized")
		}
		if indexComponent.Versions == nil || len(indexComponent.Versions) == 0 {
			return fmt.Errorf("index component versions list is empty")
		} else {
			defaultFound := false
			for _, version := range indexComponent.Versions {
				if version.Version == "" {
					return fmt.Errorf("index component versions list contains an entry with no version specified")
				}
				if version.SchemaVersion == "" {
					return fmt.Errorf("index component version %s: schema version is empty", version.Version)
				}
				if version.Links == nil || len(version.Links) == 0 {
					return fmt.Errorf("index component version %s: links are empty", version.Version)
				}
				if version.Resources == nil || len(version.Resources) == 0 {
					return fmt.Errorf("index component version %s: resources are empty", version.Version)
				}
				if version.Default {
					if !defaultFound {
						defaultFound = true
					} else {
						return fmt.Errorf("index component has multiple default versions")
					}
				}
			}
			if !defaultFound {
				return fmt.Errorf("index component has no default version defined")
			}
		}
	} else if componentType == schema.SampleDevfileType {
		if indexComponent.Versions != nil && len(indexComponent.Versions) > 0 {
			defaultFound := false
			for _, version := range indexComponent.Versions {
				if version.Version == "" {
					return fmt.Errorf("index component versions list contains an entry with no version specified")
				}
				if version.SchemaVersion == "" {
					return fmt.Errorf("index component version %s: schema version is empty", version.Version)
				}
				if version.Git == nil {
					return fmt.Errorf("index component version %s: git is empty", version.Version)
				}
				if version.Default {
					if !defaultFound {
						defaultFound = true
					} else {
						return fmt.Errorf("index component has multiple default versions")
					}
				}
			}
			if !defaultFound {
				return fmt.Errorf("index component has no default version defined")
			}
		} else {
			if indexComponent.Git == nil {
				return fmt.Errorf("index component git is empty")
			}
			if len(indexComponent.Git.Remotes) > 1 {
				return fmt.Errorf("index component has multiple remotes")
			}
		}
	}

	// Fields to be validated for both stacks and samples
	if indexComponent.Provider == "" {
		return &MissingProviderError{devfile: indexComponent.Name}
	}
	if indexComponent.SupportUrl == "" {
		return &MissingSupportUrlError{devfile: indexComponent.Name}
	}
	if len(indexComponent.Architectures) == 0 {
		return &MissingArchError{devfile: indexComponent.Name}
	}

	return nil
}

func fileExists(filepath string) bool {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return false
	}

	return true
}

func dirExists(dirpath string) error {
	dir, err := os.Stat(dirpath)
	if os.IsNotExist(err) {
		return fmt.Errorf("path: %s does not exist: %w", dirpath, err)
	}
	if !dir.IsDir() {
		return fmt.Errorf("%s is not a directory", dirpath)
	}
	return nil
}

func parseDevfileRegistry(registryDirPath string, force bool) ([]schema.Schema, error) {

	var index []schema.Schema
	stackDirPath := path.Join(registryDirPath, "stacks")
	stackDir, err := ioutil.ReadDir(stackDirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read stack directory %s: %v", stackDirPath, err)
	}
	for _, stackFolderDir := range stackDir {
		if !stackFolderDir.IsDir() {
			continue
		}
		stackFolderPath := filepath.Join(stackDirPath, stackFolderDir.Name())
		stackYamlPath := filepath.Join(stackFolderPath, stackYaml)
		// if stack.yaml exist,  parse stack.yaml
		var indexComponent schema.Schema
		if fileExists(stackYamlPath) {
			indexComponent, err = parseStackInfo(stackYamlPath)
			if err != nil {
				return nil, err
			}
			if !force {
				stackYamlErrors := validateStackInfo(indexComponent, stackFolderPath)
				if stackYamlErrors != nil {
					return nil, fmt.Errorf("%s stack.yaml is not valid: %v", stackFolderDir.Name(), stackYamlErrors)
				}
			}

			indexComponent.Versions = SortVersionByDescendingOrder(indexComponent.Versions)

			i := 0
			for i < len(indexComponent.Versions) {
				versionComponent := indexComponent.Versions[i]
				if versionComponent.Git != nil {
					// Todo: implement Git reference support, get stack content from remote repository and store in OCI registry
					fmt.Printf("stack: %v, version:%v, Git reference is currently not supported", stackFolderDir.Name(), versionComponent.Version)
					indexComponent.Versions = append(indexComponent.Versions[:i], indexComponent.Versions[i+1:]...)
					continue
				}
				stackVersonDirPath := filepath.Join(stackFolderPath, versionComponent.Version)

				err := parseStackDevfile(stackVersonDirPath, stackFolderDir.Name(), force, &versionComponent, &indexComponent)
				if err != nil {
					return nil, err
				}
				indexComponent.Versions[i] = versionComponent
				i++
			}

			for _, version := range indexComponent.Versions {
				// if a particular version supports all architectures, the top architecture List should be empty (support all) as well
				if version.Architectures == nil || len(version.Architectures) == 0 {
					indexComponent.Architectures = nil
					break
				}
			}
		} else { // if stack.yaml not exist, old stack repo struct, directly lookfor & parse devfile.yaml
			versionComponent := schema.Version{}
			err := parseStackDevfile(stackFolderPath, stackFolderDir.Name(), force, &versionComponent, &indexComponent)
			if err != nil {
				return nil, err
			}
			versionComponent.Default = true
			indexComponent.Versions = append(indexComponent.Versions, versionComponent)
		}
		indexComponent.Type = schema.StackDevfileType

		if !force {
			// Index component validation
			err := validateIndexComponent(indexComponent, schema.StackDevfileType)
			switch err.(type) {
			case *MissingProviderError, *MissingSupportUrlError, *MissingArchError:
				// log to the console as FYI if the devfile has no architectures/provider/supportUrl
				fmt.Printf("%s", err.Error())
			default:
				// only return error if we dont want to print
				if err != nil {
					return nil, fmt.Errorf("%s index component is not valid: %v", stackFolderDir.Name(), err)
				}
			}
		}

		index = append(index, indexComponent)
	}

	return index, nil
}

func parseStackDevfile(devfileDirPath string, stackName string, force bool, versionComponent *schema.Version, indexComponent *schema.Schema) error {
	// Allow devfile.yaml or .devfile.yaml
	devfilePath := filepath.Join(devfileDirPath, devfile)
	devfileHiddenPath := filepath.Join(devfileDirPath, devfileHidden)
	if fileExists(devfilePath) && fileExists(devfileHiddenPath) {
		return fmt.Errorf("both %s and %s exist", devfilePath, devfileHiddenPath)
	}
	if fileExists(devfileHiddenPath) {
		devfilePath = devfileHiddenPath
	}
	convertUri := false
	if !force {
		// Devfile validation
		devfileObj, _, err := devfileParser.ParseDevfileAndValidate(parser.ParserArgs{
			ConvertKubernetesContentInUri: &convertUri,
			Path:                          devfilePath})
		if err != nil {
			return fmt.Errorf("%s devfile is not valid: %v", devfileDirPath, err)
		}

		metadataErrors := checkForRequiredMetadata(devfileObj)
		if metadataErrors != nil {
			return fmt.Errorf("%s devfile is not valid: %v", devfileDirPath, metadataErrors)
		}
	}

	/* #nosec G304 -- devfilePath is produced using filepath.Join which cleans the input path */
	bytes, err := ioutil.ReadFile(devfilePath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %v", devfilePath, err)
	}

	var devfile schema.Devfile
	err = yaml.Unmarshal(bytes, &devfile)
	if err != nil {
		return fmt.Errorf("failed to unmarshal %s data: %v", devfilePath, err)
	}
	metaBytes, err := yaml.Marshal(devfile.Meta)
	if err != nil {
		return fmt.Errorf("failed to unmarshal %s data: %v", devfilePath, err)
	}
	var versionProp schema.Version
	err = yaml.Unmarshal(metaBytes, &versionProp)
	if err != nil {
		return fmt.Errorf("failed to unmarshal %s data: %v", devfilePath, err)
	}

	// set common properties if not set
	if indexComponent.ProjectType == "" {
		indexComponent.ProjectType = devfile.Meta.ProjectType
	}
	if indexComponent.Language == "" {
		indexComponent.Language = devfile.Meta.Language
	}
	if indexComponent.Provider == "" {
		indexComponent.Provider = devfile.Meta.Provider
	}
	if indexComponent.SupportUrl == "" {
		indexComponent.SupportUrl = devfile.Meta.SupportUrl
	}

	// for single version stack with only devfile.yaml, without stack.yaml
	// set the top-level properties for this stack
	if indexComponent.Name == "" {
		indexComponent.Name = devfile.Meta.Name
	}
	if indexComponent.DisplayName == "" {
		indexComponent.DisplayName = devfile.Meta.DisplayName
	}
	if indexComponent.Description == "" {
		indexComponent.Description = devfile.Meta.Description
	}
	if indexComponent.Icon == "" {
		indexComponent.Icon = devfile.Meta.Icon
	}

	versionProp.Default = versionComponent.Default
	*versionComponent = versionProp
	if versionComponent.Links == nil {
		versionComponent.Links = make(map[string]string)
	}
	versionComponent.Links["self"] = fmt.Sprintf("%s/%s:%s", "devfile-catalog", stackName, versionComponent.Version)
	versionComponent.SchemaVersion = devfile.SchemaVersion

	for _, starterProject := range devfile.StarterProjects {
		versionComponent.StarterProjects = append(versionComponent.StarterProjects, starterProject.Name)
	}

	for _, tag := range versionComponent.Tags {
		if !inArray(indexComponent.Tags, tag) {
			indexComponent.Tags = append(indexComponent.Tags, tag)
		}
	}

	for _, arch := range versionComponent.Architectures {
		if !inArray(indexComponent.Architectures, arch) {
			indexComponent.Architectures = append(indexComponent.Architectures, arch)
		}
	}

	// Get the files in the stack folder
	stackFiles, err := ioutil.ReadDir(devfileDirPath)
	if err != nil {
		return err
	}
	for _, stackFile := range stackFiles {
		// The registry build should have already packaged any folders and miscellaneous files into an archive.tar file
		// But, add this check as a safeguard, as OCI doesn't support unarchived folders being pushed up.
		if !stackFile.IsDir() {
			versionComponent.Resources = append(versionComponent.Resources, stackFile.Name())
		}
	}
	return nil
}

func parseExtraDevfileEntries(registryDirPath string, force bool) ([]schema.Schema, error) {
	var index []schema.Schema
	extraDevfileEntriesPath := path.Join(registryDirPath, extraDevfileEntries)
	/* #nosec G304 -- extraDevfileEntriesPath is produced using path.Join which cleans the input path */
	bytes, err := ioutil.ReadFile(extraDevfileEntriesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %v", extraDevfileEntriesPath, err)
	}

	// Only validate samples if they have been cached
	samplesDir := filepath.Join(registryDirPath, "samples")
	validateSamples := false
	if _, err := os.Stat(samplesDir); !os.IsNotExist(err) {
		validateSamples = true
	}

	var devfileEntries schema.ExtraDevfileEntries
	err = yaml.Unmarshal(bytes, &devfileEntries)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s data: %v", extraDevfileEntriesPath, err)
	}
	devfileTypes := []schema.DevfileType{schema.SampleDevfileType, schema.StackDevfileType}
	for _, devfileType := range devfileTypes {
		var devfileEntriesWithType []schema.Schema
		if devfileType == schema.SampleDevfileType {
			devfileEntriesWithType = devfileEntries.Samples
		} else if devfileType == schema.StackDevfileType {
			devfileEntriesWithType = devfileEntries.Stacks
		}
		for _, devfileEntry := range devfileEntriesWithType {
			indexComponent := devfileEntry
			indexComponent.Type = devfileType
			if !force {
				// If sample, validate devfile associated with sample as well
				// Can't handle during registry build since we don't have access to devfile library/parser
				if indexComponent.Type == schema.SampleDevfileType && validateSamples {
					if indexComponent.Versions != nil && len(indexComponent.Versions) > 0 {
						for _, version := range indexComponent.Versions {
							sampleVersonDirPath := filepath.Join(samplesDir, devfileEntry.Name, version.Version)
							devfilePath := filepath.Join(sampleVersonDirPath, "devfile.yaml")
							_, err := os.Stat(filepath.Join(devfilePath))
							if err != nil {
								// This error shouldn't occur since we check for the devfile's existence during registry build, but check for it regardless
								return nil, fmt.Errorf("%s devfile sample does not have a devfile.yaml: %v", indexComponent.Name, err)
							}
							convertUri := false
							// Validate the sample devfile
							_, _, err = devfileParser.ParseDevfileAndValidate(parser.ParserArgs{
								ConvertKubernetesContentInUri: &convertUri,
								Path:                          devfilePath})
							if err != nil {
								return nil, fmt.Errorf("%s sample devfile is not valid: %v", devfileEntry.Name, err)
							}
						}
					} else {
						devfilePath := filepath.Join(samplesDir, devfileEntry.Name, "devfile.yaml")
						_, err := os.Stat(filepath.Join(devfilePath))
						if err != nil {
							// This error shouldn't occur since we check for the devfile's existence during registry build, but check for it regardless
							return nil, fmt.Errorf("%s devfile sample does not have a devfile.yaml: %v", indexComponent.Name, err)
						}
						convertUri := false
						// Validate the sample devfile
						_, _, err = devfileParser.ParseDevfileAndValidate(parser.ParserArgs{Path: devfilePath,
							ConvertKubernetesContentInUri: &convertUri})
						if err != nil {
							return nil, fmt.Errorf("%s sample devfile is not valid: %v", devfileEntry.Name, err)
						}
					}
				}

				// Index component validation
				err := validateIndexComponent(indexComponent, devfileType)
				switch err.(type) {
				case *MissingProviderError, *MissingSupportUrlError, *MissingArchError:
					// log to the console as FYI if the devfile has no architectures/provider/supportUrl
					fmt.Printf("%s", err.Error())
				default:
					// only return error if we dont want to print
					if err != nil {
						return nil, fmt.Errorf("%s index component is not valid: %v", indexComponent.Name, err)
					}
				}
			}
			index = append(index, indexComponent)
		}
	}

	return index, nil
}

/* #nosec G304 -- stackYamlPath is produced from file.Join which cleans the input path */
func parseStackInfo(stackYamlPath string) (schema.Schema, error) {
	var index schema.Schema
	bytes, err := ioutil.ReadFile(stackYamlPath)
	if err != nil {
		return schema.Schema{}, fmt.Errorf("failed to read %s: %v", stackYamlPath, err)
	}
	err = yaml.Unmarshal(bytes, &index)
	if err != nil {
		return schema.Schema{}, fmt.Errorf("failed to unmarshal %s data: %v", stackYamlPath, err)
	}
	return index, nil
}

// checkForRequiredMetadata validates that a given devfile has the necessary metadata fields
func checkForRequiredMetadata(devfileObj parser.DevfileObj) []error {
	devfileMetadata := devfileObj.Data.GetMetadata()
	var metadataErrors []error

	if devfileMetadata.Name == "" {
		metadataErrors = append(metadataErrors, fmt.Errorf("metadata.name is not set"))
	}
	if devfileMetadata.DisplayName == "" {
		metadataErrors = append(metadataErrors, fmt.Errorf("metadata.displayName is not set"))
	}
	if devfileMetadata.Language == "" {
		metadataErrors = append(metadataErrors, fmt.Errorf("metadata.language is not set"))
	}
	if devfileMetadata.ProjectType == "" {
		metadataErrors = append(metadataErrors, fmt.Errorf("metadata.projectType is not set"))
	}

	return metadataErrors
}

func validateStackInfo(stackInfo schema.Schema, stackfolderDir string) []error {
	var errors []error

	if stackInfo.Name == "" {
		errors = append(errors, fmt.Errorf("name is not set in stack.yaml"))
	}
	if stackInfo.DisplayName == "" {
		errors = append(errors, fmt.Errorf("displayName is not set stack.yaml"))
	}
	if stackInfo.Icon == "" {
		errors = append(errors, fmt.Errorf("icon is not set stack.yaml"))
	}
	if stackInfo.Versions == nil || len(stackInfo.Versions) == 0 {
		errors = append(errors, fmt.Errorf("versions list is not set stack.yaml, or is empty"))
	}
	hasDefault := false
	for _, version := range stackInfo.Versions {
		if version.Default {
			if !hasDefault {
				hasDefault = true
			} else {
				errors = append(errors, fmt.Errorf("stack.yaml has multiple default versions"))
			}
		}

		if version.Git == nil {
			versionFolder := path.Join(stackfolderDir, version.Version)
			err := dirExists(versionFolder)
			if err != nil {
				errors = append(errors, fmt.Errorf("cannot find resorce folder for version %s defined in stack.yaml: %v", version.Version, err))
			}
		}
	}
	if !hasDefault {
		errors = append(errors, fmt.Errorf("stack.yaml does not contain a default version"))
	}

	return errors
}

// In checks if the value is in the array
func inArray(arr []string, value string) bool {
	for _, item := range arr {
		if item == value {
			return true
		}
	}
	return false
}
