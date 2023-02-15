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
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/devfile/library/v2/pkg/testingutil/filesystem"
	dfutil "github.com/devfile/library/v2/pkg/util"
	"github.com/devfile/registry-support/index/generator/schema"
	gitpkg "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

var semverRe = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)

// CloneRemoteStack downloads the stack version from a git repo outside of the registry by
// cloning then removing the local .git folder. When git.SubDir is set, fetches specified
// subdirectory only.
func CloneRemoteStack(git *schema.Git, path string, verbose bool) (err error) {

	// convert revision to referenceName type, ref name could be a branch or tag
	// if revision is not specified it would be the default branch of the project
	revision := git.Revision
	refName := plumbing.ReferenceName(git.Revision)

	if plumbing.IsHash(revision) {
		// Specifying commit in the reference name is not supported by the go-git library
		// while doing git.PlainClone()
		return fmt.Errorf("specifying commit in 'revision' is not yet supported")
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
		if err = os.RemoveAll(filepath.Join(path, ".git")); err != nil {
			return err
		}

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

// DownloadStackFromGit downloads the stack from a git repo then adds folder contents into a zip archive,
// returns byte array of zip file and error if occurs otherwise is nil. If git.SubDir is set, then
// zip file will contain contents of the specified subdirectory instead of the whole downloaded git repo.
func DownloadStackFromGit(git *schema.Git, path string, verbose bool) ([]byte, error) {
	cleanPath := filepath.Clean(path)
	zipPath := fmt.Sprintf("%s.zip", cleanPath)

	// Download from given git url. Downloaded result contains subDir
	// when specified, if error return empty bytes.
	if err := CloneRemoteStack(git, cleanPath, verbose); err != nil {
		return []byte{}, err
	}

	// Throw error if path was not created
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		return []byte{}, err
	}

	// Zip directory containing downloaded git repo
	if err := ZipDir(cleanPath, zipPath); err != nil {
		return []byte{}, err
	}

	// Read bytes from response and return, error will be nil if successful
	/* #nosec G304 -- zipPath is constructed from a clean path */
	return ioutil.ReadFile(zipPath)
}

// DownloadStackFromZipUrl downloads the zip file containing the stack at a given url, uses default filesystem
func DownloadStackFromZipUrl(zipUrl string, subDir string, path string) ([]byte, error) {
	return downloadStackFromZipUrl(zipUrl, subDir, path, filesystem.DefaultFs{})
}

// downloadStackFromZipUrl downloads the zip file containing the stack at a given url
func downloadStackFromZipUrl(zipUrl string, subDir string, path string, fs filesystem.Filesystem) ([]byte, error) {
	zipDst := fmt.Sprintf("%s.zip", filepath.Clean(path))

	// Create path if does not exist
	if err := fs.MkdirAll(path, os.ModePerm); err != nil {
		return []byte{}, err
	}

	// If subDir is specified extraction and rezipping after download is required,
	// else just download and return original zip file
	if subDir != "" {
		unzipDst := filepath.Join(path, subDir)

		// Create unzip destination if does not exist
		if err := fs.MkdirAll(unzipDst, os.ModePerm); err != nil {
			return []byte{}, err
		}

		// Download from given url and unzip subDir to given path, if error
		// return empty bytes.
		if err := dfutil.GetAndExtractZip(zipUrl, unzipDst, subDir); err != nil {
			return []byte{}, err
		}

		// Zip directory containing unzipped content
		if err := ZipDir(unzipDst, zipDst); err != nil {
			return []byte{}, err
		}
	} else {
		params := dfutil.DownloadParams{
			Request: dfutil.HTTPRequestParams{
				URL: zipUrl,
			},
			Filepath: zipDst,
		}
		if err := dfutil.DownloadFile(params); err != nil {
			return []byte{}, err
		}
	}

	// Read bytes from response and return, error will be nil if successful
	/* #nosec G304 -- zipDest is produced using a cleaned path */
	return ioutil.ReadFile(zipDst)
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

		// Create destinationPath if does not exist
		if _, err = fs.Stat(destinationPath); os.IsNotExist(err) {
			var srcinfo os.FileInfo

			if srcinfo, err = fs.Stat(srcPath); err != nil {
				return err
			}

			if err = fs.MkdirAll(destinationPath, srcinfo.Mode()); err != nil {
				return err
			}
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
func copyFileWithFs(src, unZipDst string, fs filesystem.Filesystem) error {
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

	dstfd, err := fs.Create(unZipDst)
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
	return fs.Chmod(unZipDst, srcinfo.Mode())
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
		srcfp := filepath.Join(src, fd.Name())
		dstfp := filepath.Join(dst, fd.Name())

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

// ZipDir creates a zip file from a given directory specified by the src argument into a zip archive
// specified by the *dst* argument, uses default filesystem
func ZipDir(src string, dst string) error {
	return zipDir(src, dst, filesystem.DefaultFs{})
}

// zipDir creates a zip file from a given directory specified by the src argument into a zip archive
// specified by the *dst* argument, takes a filesystem to use
func zipDir(src string, dst string, fs filesystem.Filesystem) error {
	zipFile, err := fs.Create(dst)
	if err != nil {
		return err
	}

	writer := zip.NewWriter(zipFile)
	zipper := createZipper(writer, src, fs)
	defer writer.Close()

	return fs.Walk(src, zipper)
}

// createZipper creates walk function to populate a zip file with the given writer argument
func createZipper(writer *zip.Writer, root string, fs filesystem.Filesystem) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		} else if !info.IsDir() {
			srcFile, err := fs.Open(path)
			if err != nil {
				return err
			}
			defer srcFile.Close()

			dstFile, err := writer.Create(filepath.Join(".", strings.Split(path, root)[1]))
			if err != nil {
				return err
			}

			if _, err := io.Copy(dstFile, srcFile); err != nil {
				return err
			}
		}

		return nil
	}
}

type Semver struct {
	major int
	minor int
	patch int
}

func SortVersionByDescendingOrder(versions []schema.Version) []schema.Version {
	semvers := make([]struct {
		index  int
		semver Semver
	}, len(versions))

	// convert to semver
	for i, version := range versions {
		matches := semverRe.FindStringSubmatch(version.Version)
		if len(matches) != 4 {
			fmt.Printf("error occurred while parsing semver %s", version.Version)
		}

		major, err := strconv.Atoi(matches[1])
		if err != nil {
			fmt.Printf("error %v occurred while parsing major version", err)
		}

		minor, err := strconv.Atoi(matches[2])
		if err != nil {
			fmt.Printf("error %v occurred while parsing minor version", err)
		}

		patch, err := strconv.Atoi(matches[3])
		if err != nil {
			fmt.Printf("error %v occurred while parsing patch version", err)
		}

		semvers[i] = struct {
			index  int
			semver Semver
		}{
			index: i,
			semver: Semver{
				major: major,
				minor: minor,
				patch: patch,
			},
		}
	}

	// sort semver
	sort.SliceStable(semvers, func(i, j int) bool {
		if semvers[i].semver.major > semvers[j].semver.major {
			return true
		} else if semvers[i].semver.major == semvers[j].semver.major {
			if semvers[i].semver.minor > semvers[j].semver.minor {
				return true
			} else if semvers[i].semver.minor == semvers[j].semver.minor {
				return semvers[i].semver.patch > semvers[j].semver.patch
			}
		}

		return false
	})

	// convert back to version
	sortedVersions := make([]schema.Version, len(versions))
	for i, semver := range semvers {
		sortedVersions[i] = versions[semver.index]
	}

	return sortedVersions
}
