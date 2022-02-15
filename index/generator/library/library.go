package library

import (
	"encoding/json"
	"fmt"
	"github.com/devfile/library/pkg/testingutil/filesystem"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"

	devfileParser "github.com/devfile/library/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser"
	gitpkg "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/devfile/registry-support/index/generator/schema"
	"gopkg.in/yaml.v2"
)

const (
	devfile             = "devfile.yaml"
	devfileHidden       = ".devfile.yaml"
	extraDevfileEntries = "extraDevfileEntries.yaml"
	stackYaml			= "stack.yaml"
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

func dirExists(dirpath string) error {
	dir, err := os.Stat(dirpath)
	if os.IsNotExist(err){
		return fmt.Errorf("path: %s does not exist: %w",dirpath, err)
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

			for i, versionComponent:= range indexComponent.Versions {
				if versionComponent.Git == nil {
					stackVersonDirPath := filepath.Join(stackFolderPath, versionComponent.Version)

					err := parseStackDevfile(stackVersonDirPath, stackFolderDir.Name(), force, &versionComponent, &indexComponent)
					if err != nil {
						return nil, err
					}
					indexComponent.Versions[i] = versionComponent
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

		//// Allow devfile.yaml or .devfile.yaml
		//devfilePath := filepath.Join(stackDirPath, stackFolderDir.Name(), devfile)
		//devfileHiddenPath := filepath.Join(stackDirPath, stackFolderDir.Name(), devfileHidden)
		//if fileExists(devfilePath) && fileExists(devfileHiddenPath) {
		//	return nil, fmt.Errorf("both %s and %s exist", devfilePath, devfileHiddenPath)
		//}
		//if fileExists(devfileHiddenPath) {
		//	devfilePath = devfileHiddenPath
		//}
		//
		//if !force {
		//	// Devfile validation
		//	devfileObj,_, err := devfileParser.ParseDevfileAndValidate(parser.ParserArgs{Path: devfilePath})
		//	if err != nil {
		//		return nil, fmt.Errorf("%s devfile is not valid: %v", stackFolderDir.Name(), err)
		//	}
		//
		//	metadataErrors := checkForRequiredMetadata(devfileObj)
		//	if metadataErrors != nil {
		//		return nil, fmt.Errorf("%s devfile is not valid: %v", stackFolderDir.Name(), metadataErrors)
		//	}
		//}
		//
		//bytes, err := ioutil.ReadFile(devfilePath)
		//if err != nil {
		//	return nil, fmt.Errorf("failed to read %s: %v", devfilePath, err)
		//}
		//var devfile schema.Devfile
		//err = yaml.Unmarshal(bytes, &devfile)
		//if err != nil {
		//	return nil, fmt.Errorf("failed to unmarshal %s data: %v", devfilePath, err)
		//}
		//indexComponent := devfile.Meta
		//if indexComponent.Links == nil {
		//	indexComponent.Links = make(map[string]string)
		//}
		//indexComponent.Links["self"] = fmt.Sprintf("%s/%s:%s", "devfile-catalog", indexComponent.Name, "latest")
		//indexComponent.Type = schema.StackDevfileType
		//
		//for _, starterProject := range devfile.StarterProjects {
		//	indexComponent.StarterProjects = append(indexComponent.StarterProjects, starterProject.Name)
		//}
		//
		//// Get the files in the stack folder
		//stackFolder := filepath.Join(stackDirPath, stackFolderDir.Name())
		//stackFiles, err := ioutil.ReadDir(stackFolder)
		//if err != nil {
		//	return index, err
		//}
		//for _, stackFile := range stackFiles {
		//	// The registry build should have already packaged any folders and miscellaneous files into an archive.tar file
		//	// But, add this check as a safeguard, as OCI doesn't support unarchived folders being pushed up.
		//	if !stackFile.IsDir() {
		//		indexComponent.Resources = append(indexComponent.Resources, stackFile.Name())
		//	}
		//}
		//
		//if !force {
		//	// Index component validation
		//	err := validateIndexComponent(indexComponent, schema.StackDevfileType)
		//	switch err.(type) {
		//	case *MissingProviderError, *MissingSupportUrlError, *MissingArchError:
		//		// log to the console as FYI if the devfile has no architectures/provider/supportUrl
		//		fmt.Printf("%s", err.Error())
		//	default:
		//		// only return error if we dont want to print
		//		if err != nil {
		//			return nil, fmt.Errorf("%s index component is not valid: %v", stackFolderDir.Name(), err)
		//		}
		//	}
		//}

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

	if !force {
		// Devfile validation
		devfileObj,_, err := devfileParser.ParseDevfileAndValidate(parser.ParserArgs{Path: devfilePath})
		if err != nil {
			return fmt.Errorf("%s devfile is not valid: %v", devfileDirPath, err)
		}

		metadataErrors := checkForRequiredMetadata(devfileObj)
		if metadataErrors != nil {
			return fmt.Errorf("%s devfile is not valid: %v", devfileDirPath, metadataErrors)
		}
	}

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

	//if !force {
	//	// Index component validation
	//	err := validateIndexComponent(versionComponent, schema.StackDevfileType)
	//	switch err.(type) {
	//	case *MissingProviderError, *MissingSupportUrlError, *MissingArchError:
	//		// log to the console as FYI if the devfile has no architectures/provider/supportUrl
	//		fmt.Printf("%s", err.Error())
	//	default:
	//		// only return error if we dont want to print
	//		if err != nil {
	//			return schema.Version{}, fmt.Errorf("%s index component is not valid: %v", stackFolder, err)
	//		}
	//	}
	//}
	return nil
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

func validateStackInfo (stackInfo schema.Schema, stackfolderDir string) []error {
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

// downloadRemoteStack downloads the stack version outside of the registry repo
func downloadRemoteStack(git *schema.Git, path string, verbose bool) (err error) {

	// convert revision to referenceName type, ref name could be a branch or tag
	// if revision is not specified it would be the default branch of the project
	revision := git.Revision
	refName := plumbing.ReferenceName(git.Revision)

	if plumbing.IsHash(revision) {
		// Specifying commit in the reference name is not supported by the go-git library
		// while doing git.PlainClone()
		fmt.Printf("Specifying commit in 'revision' is not yet supported.")
		// overriding revision to empty as we do not support this
		revision = ""
	}

	if revision != "" {
		// lets consider revision to be a branch name first
		refName = plumbing.NewBranchReferenceName(revision)
	}


	cloneOptions := &gitpkg.CloneOptions{
		URL:           git.Url,
		RemoteName:    git.RemoteName,
		ReferenceName: refName,
		SingleBranch:  true,
		// we don't need history for starter projects
		Depth: 1,
	}

	originalPath := ""
	if git.SubDir != "" {
		originalPath = path
		path, err = ioutil.TempDir("", "")
		if err != nil {
			return err
		}
	}

	_, err = gitpkg.PlainClone(path, false, cloneOptions)

	if err != nil {

		// it returns the following error if no matching ref found
		// if we get this error, we are trying again considering revision as tag, only if revision is specified.
		if _, ok := err.(gitpkg.NoMatchingRefSpecError); !ok || revision == "" {
			return err
		}

		// try again to consider revision as tag name
		cloneOptions.ReferenceName = plumbing.NewTagReferenceName(revision)
		// remove if any .git folder downloaded in above try
		_ = os.RemoveAll(filepath.Join(path, ".git"))
		_, err = gitpkg.PlainClone(path, false, cloneOptions)
		if err != nil {
			return err
		}
	}

	// we don't want to download project be a git repo
	err = os.RemoveAll(filepath.Join(path, ".git"))
	if err != nil {
		// we don't need to return (fail) if this happens
		fmt.Printf("Unable to delete .git from cloned devfile repository")
	}

	if git.SubDir != "" {
		err = GitSubDir(path, originalPath,
			git.SubDir)
		if err != nil {
			return err
		}
	}

	return nil

}

// GitSubDir handles subDir for git components using the default filesystem
func GitSubDir(srcPath, destinationPath, subDir string) error {
	return gitSubDir(srcPath, destinationPath, subDir, filesystem.DefaultFs{})
}

// gitSubDir handles subDir for git components
func gitSubDir(srcPath, destinationPath, subDir string, fs filesystem.Filesystem) error {
	go StartSignalWatcher([]os.Signal{syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt}, func(_ os.Signal) {
		err := cleanDir(destinationPath, map[string]bool{
			"devfile.yaml": true,
		}, fs)
		if err != nil {
			fmt.Printf("error %v occurred while calling handleInterruptedSubDir", err)
		}
		err = fs.RemoveAll(srcPath)
		if err != nil {
			fmt.Printf("error %v occurred during temp folder clean up", err)
		}
	})

	err := func() error {
		// Open the directory.
		outputDirRead, err := fs.Open(filepath.Join(srcPath, subDir))
		if err != nil {
			return err
		}
		defer func() {
			if err1 := outputDirRead.Close(); err1 != nil {
				fmt.Printf("err occurred while closing temp dir: %v", err1)

			}
		}()
		// Call Readdir to get all files.
		outputDirFiles, err := outputDirRead.Readdir(0)
		if err != nil {
			return err
		}

		// Loop over files.
		for outputIndex := range outputDirFiles {
			outputFileHere := outputDirFiles[outputIndex]

			// Get name of file.
			fileName := outputFileHere.Name()

			oldPath := filepath.Join(srcPath, subDir, fileName)

			if outputFileHere.IsDir() {
				err = copyDirWithFS(oldPath, filepath.Join(destinationPath, fileName), fs)
			} else {
				err = copyFileWithFs(oldPath, filepath.Join(destinationPath, fileName), fs)
			}

			if err != nil {
				return err
			}
		}
		return nil
	}()
	if err != nil {
		return err
	}
	return fs.RemoveAll(srcPath)
}

// copyFileWithFs copies a single file from src to dst
func copyFileWithFs(src, dst string, fs filesystem.Filesystem) error {
	var err error
	var srcinfo os.FileInfo

	srcfd, err := fs.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if e := srcfd.Close(); e != nil {
			fmt.Printf("err occurred while closing file: %v", e)
		}
	}()

	dstfd, err := fs.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if e := dstfd.Close(); e != nil {
			fmt.Printf("err occurred while closing file: %v", e)
		}
	}()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = fs.Stat(src); err != nil {
		return err
	}
	return fs.Chmod(dst, srcinfo.Mode())
}

// copyDirWithFS copies a whole directory recursively
func copyDirWithFS(src string, dst string, fs filesystem.Filesystem) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = fs.Stat(src); err != nil {
		return err
	}

	if err = fs.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = fs.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = copyDirWithFS(srcfp, dstfp, fs); err != nil {
				return err
			}
		} else {
			if err = copyFileWithFs(srcfp, dstfp, fs); err != nil {
				return err
			}
		}
	}
	return nil
}

// StartSignalWatcher watches for signals and handles the situation before exiting the program
func StartSignalWatcher(watchSignals []os.Signal, handle func(receivedSignal os.Signal)) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, watchSignals...)
	defer signal.Stop(signals)

	receivedSignal := <-signals
	handle(receivedSignal)
	// exit here to stop spinners from rotating
	os.Exit(1)
}

// cleanDir cleans the original folder during events like interrupted copy etc
// it leaves the given files behind for later use
func cleanDir(originalPath string, leaveBehindFiles map[string]bool, fs filesystem.Filesystem) error {
	// Open the directory.
	outputDirRead, err := fs.Open(originalPath)
	if err != nil {
		return err
	}

	// Call Readdir to get all files.
	outputDirFiles, err := outputDirRead.Readdir(0)
	if err != nil {
		return err
	}

	// Loop over files.
	for _, file := range outputDirFiles {
		if value, ok := leaveBehindFiles[file.Name()]; ok && value {
			continue
		}
		err = fs.RemoveAll(filepath.Join(originalPath, file.Name()))
		if err != nil {
			return err
		}
	}
	return err
}