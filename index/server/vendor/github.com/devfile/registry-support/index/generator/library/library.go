package library

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	devfileParser "github.com/devfile/library/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/registry-support/index/generator/schema"
	"gopkg.in/yaml.v2"
)

const (
	devfile             = "devfile.yaml"
	devfileHidden       = ".devfile.yaml"
	extraDevfileEntries = "extraDevfileEntries.yaml"
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
		if indexComponent.Links == nil {
			return fmt.Errorf("index component links are empty")
		}
		if indexComponent.Resources == nil {
			return fmt.Errorf("index component resources are empty")
		}
	} else if componentType == schema.SampleDevfileType {
		if indexComponent.Git == nil {
			return fmt.Errorf("index component git is empty")
		}
		if len(indexComponent.Git.Remotes) > 1 {
			return fmt.Errorf("index component has multiple remotes")
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

func parseDevfileRegistry(registryDirPath string, force bool) ([]schema.Schema, error) {
	var index []schema.Schema
	stackDirPath := path.Join(registryDirPath, "stacks")
	stackDir, err := ioutil.ReadDir(stackDirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read stack directory %s: %v", stackDirPath, err)
	}
	for _, devfileDir := range stackDir {
		if !devfileDir.IsDir() {
			continue
		}

		// Allow devfile.yaml or .devfile.yaml
		devfilePath := filepath.Join(stackDirPath, devfileDir.Name(), devfile)
		devfileHiddenPath := filepath.Join(stackDirPath, devfileDir.Name(), devfileHidden)
		if fileExists(devfilePath) && fileExists(devfileHiddenPath) {
			return nil, fmt.Errorf("both %s and %s exist", devfilePath, devfileHiddenPath)
		}
		if fileExists(devfileHiddenPath) {
			devfilePath = devfileHiddenPath
		}

		if !force {
			// Devfile validation
			devfileObj, err := devfileParser.ParseAndValidate(devfilePath)
			if err != nil {
				return nil, fmt.Errorf("%s devfile is not valid: %v", devfileDir.Name(), err)
			}

			metadataErrors := checkForRequiredMetadata(devfileObj)
			if metadataErrors != nil {
				return nil, fmt.Errorf("%s devfile is not valid: %v", devfileDir.Name(), metadataErrors)
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
		indexComponent.Type = schema.StackDevfileType

		for _, starterProject := range devfile.StarterProjects {
			indexComponent.StarterProjects = append(indexComponent.StarterProjects, starterProject.Name)
		}

		// Get the files in the stack folder
		stackFolder := filepath.Join(stackDirPath, devfileDir.Name())
		stackFiles, err := ioutil.ReadDir(stackFolder)
		if err != nil {
			return index, err
		}
		for _, stackFile := range stackFiles {
			// The registry build should have already packaged any folders and miscellaneous files into an archive.tar file
			// But, add this check as a safeguard, as OCI doesn't support unarchived folders being pushed up.
			if !stackFile.IsDir() {
				indexComponent.Resources = append(indexComponent.Resources, stackFile.Name())
			}
		}

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
					return nil, fmt.Errorf("%s index component is not valid: %v", devfileDir.Name(), err)
				}
			}
		}

		index = append(index, indexComponent)
	}

	return index, nil
}

func parseExtraDevfileEntries(registryDirPath string, force bool) ([]schema.Schema, error) {
	var index []schema.Schema
	extraDevfileEntriesPath := path.Join(registryDirPath, extraDevfileEntries)
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
					devfilePath := filepath.Join(samplesDir, devfileEntry.Name, "devfile.yaml")
					_, err := os.Stat(filepath.Join(devfilePath))
					if err != nil {
						// This error shouldn't occur since we check for the devfile's existence during registry build, but check for it regardless
						return nil, fmt.Errorf("%s devfile sample does not have a devfile.yaml: %v", indexComponent.Name, err)
					}

					// Validate the sample devfile
					_, err = devfileParser.ParseAndValidate(devfilePath)
					if err != nil {
						return nil, fmt.Errorf("%s sample devfile is not valid: %v", devfileEntry.Name, err)
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
