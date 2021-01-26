package library

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"text/tabwriter"
	"time"

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

	scheme                = "http"
	httpRequestTimeout    = 30 * time.Second // httpRequestTimeout configures timeout of all HTTP requests
	responseHeaderTimeout = 30 * time.Second // responseHeaderTimeout is the timeout to retrieve the server's response headers
)

var (
	stacksPath = os.Getenv("DEVFILE_STACKS")
	indexPath  = os.Getenv("DEVFILE_INDEX")
)

// GetRegistryIndex gets the index of the specified registry
func GetRegistryIndex(registry string) ([]indexSchema.Schema, error) {
	// Call index server REST API to get the index
	url := scheme + "://" + path.Join(registry, "index")
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
	var registryIndex []indexSchema.Schema
	err = json.Unmarshal(bytes, &registryIndex)
	if err != nil {
		return nil, err
	}
	return registryIndex, nil
}

// ListRegistryStacks lists the stacks of the registry
func ListRegistryStacks(registry string) error {
	// Get the registry index
	registryIndex, err := GetRegistryIndex(registry)
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

// PullStackFromRegistry pulls stack from registry with allowed media types
func PullStackFromRegistry(registry string, stack string, allowedMediaTypes []string) error {
	// Get the registry index
	registryIndex, err := GetRegistryIndex(registry)
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
	ctx := context.Background()
	resolver := docker.NewResolver(docker.ResolverOptions{PlainHTTP: true})
	ref := path.Join(registry, stackIndex.Links["self"])
	fileStore := content.NewFileStore("")
	defer fileStore.Close()

	// Pull stack from registry and save it to disk
	log.Printf("Pulling stack %s from %s with allowed media types %v...\n", stack, ref, allowedMediaTypes)
	desc, _, err := oras.Pull(ctx, resolver, ref, fileStore, oras.WithAllowedMediaTypes(allowedMediaTypes))
	if err != nil {
		return fmt.Errorf("Failed to pull stack %s from %s: %v", stack, ref, err)
	}
	log.Printf("Pulled stack %s from %s with digest %s\n", stack, ref, desc.Digest)

	// Decompress archive.tar
	if _, err := os.Stat("archive.tar"); err == nil {
		err := decompress(".", "archive.tar")
		if err != nil {
			return err
		}

		err = os.RemoveAll("archive.tar")
		if err != nil {
			return err
		}
	}

	return nil
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
