//
// Copyright (c) 2020 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation

package library

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"text/tabwriter"
	"time"

	orasctx "github.com/deislabs/oras/pkg/context"

	"github.com/containerd/containerd/remotes/docker"
	"github.com/deislabs/oras/pkg/content"
	"github.com/deislabs/oras/pkg/oras"
	indexSchema "github.com/devfile/registry-support/index/generator/schema"
)

const (
	// Devfile media types
	DevfileConfigMediaType  = "application/vnd.devfileio.devfile.config.v2+json"
	DevfileMediaType        = "application/vnd.devfileio.devfile.layer.v1"
	DevfileVSXMediaType     = "application/vnd.devfileio.vsx.layer.v1.tar"
	DevfileSVGLogoMediaType = "image/svg+xml"
	DevfilePNGLogoMediaType = "image/png"
	DevfileArchiveMediaType = "application/x-tar"

	httpRequestTimeout    = 30 * time.Second // httpRequestTimeout configures timeout of all HTTP requests
	responseHeaderTimeout = 30 * time.Second // responseHeaderTimeout is the timeout to retrieve the server's response headers
)

var (
	stacksPath               = os.Getenv("DEVFILE_STACKS")
	indexPath                = os.Getenv("DEVFILE_INDEX")
	DevfileMediaTypeList     = []string{DevfileMediaType}
	DevfileAllMediaTypesList = []string{DevfileMediaType, DevfilePNGLogoMediaType, DevfileSVGLogoMediaType, DevfileVSXMediaType, DevfileArchiveMediaType}
)

// GetRegistryIndex returns the list of stacks and/or samples, more specifically
// it gets the stacks and/or samples content of the index of the specified registry
// for listing the stacks and/or samples
func GetRegistryIndex(registry string, devfileTypes ...indexSchema.DevfileType) ([]indexSchema.Schema, error) {
	var registryIndex []indexSchema.Schema

	// Call index server REST API to get the index
	urlObj, err := url.Parse(registry)
	if err != nil {
		return nil, err
	}
	getStack := false
	getSample := false
	for _, devfileType := range devfileTypes {
		if devfileType == indexSchema.StackDevfileType {
			getStack = true
		} else if devfileType == indexSchema.SampleDevfileType {
			getSample = true
		}
	}
	if getStack && getSample {
		urlObj.Path = path.Join(urlObj.Path, "index", "all")
	} else if getStack && !getSample {
		urlObj.Path = path.Join(urlObj.Path, "index")
	} else if getSample && !getStack {
		urlObj.Path = path.Join(urlObj.Path, "index", "sample")
	} else {
		return registryIndex, nil
	}
	url := urlObj.String()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	httpClient := &http.Client{
		Transport: &http.Transport{
			ResponseHeaderTimeout: responseHeaderTimeout,
		},
		Timeout: httpRequestTimeout,
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &registryIndex)
	if err != nil {
		return nil, err
	}
	return registryIndex, nil
}

// PrintRegistry prints the registry with devfile type
func PrintRegistry(registry string, devfileType string) error {
	// Get the registry index
	var registryIndex []indexSchema.Schema
	var err error
	if devfileType == string(indexSchema.StackDevfileType) {
		registryIndex, err = GetRegistryIndex(registry, indexSchema.StackDevfileType)
	} else if devfileType == string(indexSchema.SampleDevfileType) {
		registryIndex, err = GetRegistryIndex(registry, indexSchema.SampleDevfileType)
	} else if devfileType == "all" {
		registryIndex, err = GetRegistryIndex(registry, indexSchema.StackDevfileType, indexSchema.SampleDevfileType)
	}
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "Name", "\t", "Description", "\t")
	for _, stackIndex := range registryIndex {
		fmt.Fprintln(w, stackIndex.Name, "\t", stackIndex.Description, "\t")
	}
	w.Flush()
	return nil
}

// PullStackByMediaTypesFromRegistry pulls stack from registry with allowed media types to the destination directory
func PullStackByMediaTypesFromRegistry(registry string, stack string, allowedMediaTypes []string, destDir string) error {
	// Get the registry index
	registryIndex, err := GetRegistryIndex(registry, indexSchema.StackDevfileType)
	if err != nil {
		return err
	}

	// Parse the index to get the specified stack's metadata in the index
	var stackIndex indexSchema.Schema
	exist := false
	for _, item := range registryIndex {
		if item.Name == stack {
			stackIndex = item
			exist = true
			break
		}
	}
	if !exist {
		return fmt.Errorf("Stack %s does not exist in the registry %s", stack, registry)
	}

	// Pull stack initialization
	ctx := orasctx.Background()
	urlObj, err := url.Parse(registry)
	if err != nil {
		return err
	}
	plainHTTP := true
	if urlObj.Scheme == "https" {
		plainHTTP = false
	}
	resolver := docker.NewResolver(docker.ResolverOptions{PlainHTTP: plainHTTP})
	ref := path.Join(urlObj.Host, stackIndex.Links["self"])
	fileStore := content.NewFileStore(destDir)
	defer fileStore.Close()

	// Pull stack from registry and save it to disk
	_, _, err = oras.Pull(ctx, resolver, ref, fileStore, oras.WithAllowedMediaTypes(allowedMediaTypes))
	if err != nil {
		return fmt.Errorf("Failed to pull stack %s from %s with allowed media types %v: %v", stack, ref, allowedMediaTypes, err)
	}

	// Decompress archive.tar
	archivePath := filepath.Join(destDir, "archive.tar")
	if _, err := os.Stat(archivePath); err == nil {
		err := decompress(destDir, archivePath)
		if err != nil {
			return err
		}

		err = os.RemoveAll(archivePath)
		if err != nil {
			return err
		}
	}

	return nil
}

// PullStackFromRegistry pulls stack from registry with all stack resources (all media types) to the destination directory
func PullStackFromRegistry(registry string, stack string, destDir string) error {
	return PullStackByMediaTypesFromRegistry(registry, stack, DevfileAllMediaTypesList, destDir)
}

// decompress extracts the archive file
func decompress(targetDir string, tarFile string) error {
	reader, err := os.Open(tarFile)
	if err != nil {
		return err
	}
	defer reader.Close()

	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		target := path.Join(targetDir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(target, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
		case tar.TypeReg:
			w, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			_, err = io.Copy(w, tarReader)
			if err != nil {
				return err
			}
			w.Close()
		default:
			log.Printf("Unsupported type: %v", header.Typeflag)
			break
		}
	}

	return nil
}
