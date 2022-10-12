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

package server

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	indexSchema "github.com/devfile/registry-support/index/generator/schema"

	"github.com/containerd/containerd/remotes/docker"
	"github.com/deislabs/oras/pkg/content"
	"github.com/deislabs/oras/pkg/oras"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// pushStackToRegistry pushes the given devfile stack to the OCI registry
func pushStackToRegistry(versionComponent indexSchema.Version, stackName string) error {
	// Load the devfile into memory and set up the pushing resource (file name, file content, media type, ref)
	memoryStore := content.NewMemoryStore()
	pushContents := []ocispec.Descriptor{}
	for _, resource := range versionComponent.Resources {
		if resource == "meta.yaml" || strings.HasSuffix(resource, "-offline.zip") {
			// Some registries may still have the meta.yaml (we don't need it) or offline resources in it, so skip pushing these up
			continue
		}

		// Get the media type that corresponds to the resource
		// Some resources have media types that depends on the entire filename (e.g. devfile.yaml, archive.tar),
		// others just depend on the file extension (e.g. vsx files)
		var mediaType string
		var found bool
		switch resource {
		case devfileName, devfileNameHidden, svgLogoName, pngLogoName, archiveName:
			// Get the media type associated with the file
			if mediaType, found = mediaTypeMapping[resource]; !found {
				return errors.New("media type not found for file " + resource)
			}
		default:
			// Probably vsx file, but get the extension of the file just in case
			fileExtension := filepath.Ext(resource)
			if mediaType, found = mediaTypeMapping[fileExtension]; !found {
				return errors.New("media type not found for file extension" + fileExtension)
			}
		}

		resourcePath := filepath.Join(stacksPath, stackName, versionComponent.Version, resource)
		// Load the resource into memory and add to the push contents
		if _, err := os.Stat(resourcePath); os.IsNotExist(err) {
			resourcePath = filepath.Join(stacksPath, stackName, resource)
		}
		/* #nosec G304 -- resourcePath is constructed from filepath.Join which cleans the input paths */
		resourceContent, err := ioutil.ReadFile(resourcePath)
		if err != nil {
			return err
		}

		desc := memoryStore.Add(resource, mediaType, resourceContent)
		pushContents = append(pushContents, desc)

	}

	ref := path.Join(registryService, "/", versionComponent.Links["self"])

	ctx := context.Background()
	resolver := docker.NewResolver(docker.ResolverOptions{PlainHTTP: true})
	log.Printf("Pushing %s version %s to %s...\n", stackName, versionComponent.Version, ref)
	desc, err := oras.Push(ctx, resolver, ref, memoryStore, pushContents, oras.WithConfigMediaType(devfileConfigMediaType))
	if err != nil {
		return fmt.Errorf("failed to push %s version %s to %s: %v", stackName, versionComponent.Version, ref, err)
	}
	log.Printf("Pushed to %s with digest %s\n", ref, desc.Digest)
	return nil
}

// pullStackFromRegistry pulls the given devfile stack from the OCI registry
func pullStackFromRegistry(versionComponent indexSchema.Version) ([]byte, error) {
	// Pull the devfile from registry and save to disk
	ref := path.Join(registryService, "/", versionComponent.Links["self"])

	ctx := context.Background()
	resolver := docker.NewResolver(docker.ResolverOptions{PlainHTTP: true})

	// Initialize memory store
	memoryStore := content.NewMemoryStore()
	allowedMediaTypes := []string{devfileMediaType}

	var devfile string
	for _, resource := range versionComponent.Resources {
		if resource == devfileName {
			devfile = devfileName
			break
		}
		if resource == devfileNameHidden {
			devfile = devfileNameHidden
			break
		}
	}
	log.Printf("Pulling %s from %s...\n", devfile, ref)
	desc, _, err := oras.Pull(ctx, resolver, ref, memoryStore, oras.WithAllowedMediaTypes(allowedMediaTypes))
	if err != nil {
		return nil, fmt.Errorf("failed to pull %s from %s: %v", devfile, ref, err)
	}
	_, bytes, ok := memoryStore.GetByName(devfile)
	if !ok {
		return nil, fmt.Errorf("failed to load %s to memory", devfile)
	}

	log.Printf("Pulled from %s with digest %s\n", ref, desc.Digest)
	return bytes, nil
}
