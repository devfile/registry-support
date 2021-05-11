package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/deislabs/oras/pkg/content"
	"github.com/deislabs/oras/pkg/oras"
	indexLibrary "github.com/devfile/registry-support/index/generator/library"
	indexSchema "github.com/devfile/registry-support/index/generator/schema"

	"github.com/containerd/containerd/remotes/docker"
	_ "github.com/devfile/registry-support/index/server/docs"
	"github.com/gin-gonic/gin"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	// Constants for resource names and media types
	archiveMediaType       = "application/x-tar"
	archiveName            = "archive.tar"
	devfileName            = "devfile.yaml"
	devfileNameHidden      = ".devfile.yaml"
	devfileConfigMediaType = "application/vnd.devfileio.devfile.config.v2+json"
	devfileMediaType       = "application/vnd.devfileio.devfile.layer.v1"
	pngLogoMediaType       = "image/png"
	pngLogoName            = "logo.png"
	svgLogoMediaType       = "image/svg+xml"
	svgLogoName            = "logo.svg"
	vsxMediaType           = "application/vnd.devfileio.vsx.layer.v1.tar"
	vsxName                = "vsx"

	scheme          = "http"
	registryService = "localhost:5000"
)

var mediaTypeMapping = map[string]string{
	devfileName:       devfileMediaType,
	devfileNameHidden: devfileMediaType,
	vsxName:           vsxMediaType,
	svgLogoName:       svgLogoMediaType,
	pngLogoName:       pngLogoMediaType,
	archiveName:       archiveMediaType,
}

var (
	stacksPath      = os.Getenv("DEVFILE_STACKS")
	indexPath       = os.Getenv("DEVFILE_INDEX")
	sampleIndexPath = os.Getenv("DEVFILE_SAMPLE_INDEX")
	stackIndexPath  = os.Getenv("DEVFILE_STACK_INDEX")
	getIndexLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "index_http_request_duration_seconds",
			Help:    "Latency of index request in seconds.",
			Buckets: prometheus.LinearBuckets(0.5, 0.5, 10),
		},
		[]string{"status"},
	)
)

func main() {
	// Enable metrics
	// Run on a separate port and router from the index server so that it's not exposed publicly
	http.Handle("/metrics", promhttp.Handler())
	prometheus.MustRegister(getIndexLatency)
	go http.ListenAndServe(":7071", nil)

	// Wait until registry is up and running
	err := wait.PollImmediate(time.Millisecond, time.Second*30, func() (bool, error) {
		resp, err := http.Get(scheme + "://" + registryService)
		if err != nil {
			log.Println(err.Error())
			return false, nil
		}

		if resp.StatusCode == http.StatusOK {
			log.Println("Registry is up and running")
			return true, nil
		}

		log.Println("Waiting for registry to start...")
		return false, nil
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	// Load index file
	bytes, err := ioutil.ReadFile(indexPath)
	if err != nil {
		log.Fatalf("failed to read index file: %v", err)
	}

	// TODO: add code block to parse index.json by using common library
	// Issue: https://github.com/devfile/api/issues/223
	var index []indexSchema.Schema
	err = json.Unmarshal(bytes, &index)
	if err != nil {
		log.Fatalf("failed to unmarshal index file: %v", err)
	}

	// Before starting the server, push the devfile artifacts to the registry
	// Build sample_index.json and stack_index.json given index.json
	var sampleIndex []indexSchema.Schema
	var stackIndex []indexSchema.Schema
	for _, devfileIndex := range index {
		if devfileIndex.Type == indexSchema.SampleDevfileType {
			sampleIndex = append(sampleIndex, devfileIndex)
		} else if devfileIndex.Type == indexSchema.StackDevfileType {
			stackIndex = append(stackIndex, devfileIndex)
		}

		if len(devfileIndex.Resources) != 0 {
			err := pushStackToRegistry(devfileIndex)
			if err != nil {
				log.Fatal(err.Error())
			}
		}
	}
	err = indexLibrary.CreateIndexFile(sampleIndex, sampleIndexPath)
	if err != nil {
		log.Fatalf("failed to generate %s: %v", sampleIndexPath, err)
	}
	indexLibrary.CreateIndexFile(stackIndex, stackIndexPath)
	if err != nil {
		log.Fatalf("failed to generate %s: %v", stackIndexPath, err)
	}

	// Start the server and serve requests and index.json
	router := gin.Default()

	router.GET("/", serveDevfileIndex)
	router.GET("/index", serveDevfileIndex)
	router.GET("/index.json", serveDevfileIndex)

	router.GET("/index/:type", func(c *gin.Context) {
		indexType := c.Param("type")

		// Serve the index with type
		if indexType == string(indexSchema.SampleDevfileType) {
			c.File(sampleIndexPath)
		} else if indexType == "all" {
			c.File(indexPath)
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"status": fmt.Sprintf("the devfile with %s type didn't exist", indexType),
			})
		}
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "the server is up and running",
		})
	})

	router.GET("/devfiles/:name", func(c *gin.Context) {
		name := c.Param("name")
		for _, devfileIndex := range index {
			if devfileIndex.Name == name {
				bytes, err := pullStackFromRegistry(devfileIndex)
				if err != nil {
					log.Fatal(err.Error())
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":  err.Error(),
						"status": fmt.Sprintf("failed to pull the devfile of %s", name),
					})
					return
				}

				c.Data(http.StatusOK, http.DetectContentType(bytes), bytes)
				return
			}
		}

		c.JSON(http.StatusNotFound, gin.H{
			"status": fmt.Sprintf("the devfile of %s didn't exist", name),
		})
	})

	router.Static("/stacks", stacksPath)

	router.Run(":7070")
}

// pushStackToRegistry pushes the given devfile stack to the OCI registry
func pushStackToRegistry(devfileIndex indexSchema.Schema) error {
	// Load the devfile into memory and set up the pushing resource (file name, file content, media type, ref)
	memoryStore := content.NewMemoryStore()
	pushContents := []ocispec.Descriptor{}
	for _, resource := range devfileIndex.Resources {
		if resource == "meta.yaml" {
			// Some registries may still have the meta.yaml in it, but we don't need it, so skip pushing it up
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

		// Load the resource into memory and add to the push contents
		resourceContent, err := ioutil.ReadFile(filepath.Join(stacksPath, devfileIndex.Name, resource))
		if err != nil {
			return err
		}

		desc := memoryStore.Add(resource, mediaType, resourceContent)
		pushContents = append(pushContents, desc)

	}

	ref := path.Join(registryService, "/", devfileIndex.Links["self"])

	ctx := context.Background()
	resolver := docker.NewResolver(docker.ResolverOptions{PlainHTTP: true})
	log.Printf("Pushing %s to %s...\n", devfileIndex.Name, ref)
	desc, err := oras.Push(ctx, resolver, ref, memoryStore, pushContents, oras.WithConfigMediaType(devfileConfigMediaType))
	if err != nil {
		return fmt.Errorf("failed to push %s to %s: %v", devfileIndex.Name, ref, err)
	}
	log.Printf("Pushed to %s with digest %s\n", ref, desc.Digest)
	return nil
}

// pullStackFromRegistry pulls the given devfile stack from the OCI registry
func pullStackFromRegistry(devfileIndex indexSchema.Schema) ([]byte, error) {
	// Pull the devfile from registry and save to disk
	ref := path.Join(registryService, "/", devfileIndex.Links["self"])

	ctx := context.Background()
	resolver := docker.NewResolver(docker.ResolverOptions{PlainHTTP: true})

	// Initialize memory store
	memoryStore := content.NewMemoryStore()
	allowedMediaTypes := []string{devfileMediaType}

	var devfile string
	for _, resource := range devfileIndex.Resources {
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

// serveDevfileIndex serves the index.json file located in the container at `serveDevfileIndex`
func serveDevfileIndex(c *gin.Context) {
	// Start the counter for the request
	var status string
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		getIndexLatency.WithLabelValues(status).Observe(v)
	}))
	defer func() {
		timer.ObserveDuration()
	}()

	// Serve the index.json file
	c.File(stackIndexPath)
}
