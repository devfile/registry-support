package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
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
	"gopkg.in/segmentio/analytics-go.v3"
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
	viewerService   = "localhost:3000"
	encodeFormat    = "base64"
	telemetryKey    = "6HBMiy5UxBtsbxXx7O4n0t0u4dt8IAR3"
	defaultUser     = "anonymous"
)

var eventTrackMap = map[string]string{
	"list":     "list devfile",
	"view":     "view devfile",
	"download": "download devfile",
}

var mediaTypeMapping = map[string]string{
	devfileName:       devfileMediaType,
	devfileNameHidden: devfileMediaType,
	vsxName:           vsxMediaType,
	svgLogoName:       svgLogoMediaType,
	pngLogoName:       pngLogoMediaType,
	archiveName:       archiveMediaType,
}

var (
	stacksPath            = os.Getenv("DEVFILE_STACKS")
	samplesPath           = os.Getenv("DEVFILE_SAMPLES")
	indexPath             = os.Getenv("DEVFILE_INDEX")
	base64IndexPath       = os.Getenv("DEVFILE_BASE64_INDEX")
	sampleIndexPath       = os.Getenv("DEVFILE_SAMPLE_INDEX")
	sampleBase64IndexPath = os.Getenv("DEVFILE_SAMPLE_BASE64_INDEX")
	stackIndexPath        = os.Getenv("DEVFILE_STACK_INDEX")
	stackBase64IndexPath  = os.Getenv("DEVFILE_STACK_BASE64_INDEX")
	enableTelemetry       = getOptionalEnv("ENABLE_TELEMETRY", false).(bool)
	registry              = getOptionalEnv("REGISTRY_NAME", "anonymous")
	getIndexLatency       = prometheus.NewHistogramVec(
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
		log.Fatalf("failed to read index file: %s", err.Error())
	}
	var index []indexSchema.Schema
	err = json.Unmarshal(bytes, &index)
	if err != nil {
		log.Fatalf("failed to unmarshal index file: %s", err.Error())
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

	// Logs for telemetry configuration
	if enableTelemetry {
		log.Println("Telemetry is enabled")
		log.Printf("The registry name for telemetry is %s\n", registry)
	} else {
		log.Println("Telemetry is not enabled")
	}

	// Start the server and serve requests and index.json
	router := gin.Default()

	// Registry REST APIs
	router.GET("/", serveRootEndpoint)
	router.GET("/index", serveDevfileIndex)
	router.GET("/index/:type", serveDevfileIndexWithType)
	router.GET("/health", serveHealthCheck)
	router.GET("/devfiles/:name", serveDevfile)

	// Set up a simple proxy for /v2 endpoints
	// Only allow HEAD and GET requests
	router.HEAD("/v2/*proxyPath", ociServerProxy)
	router.GET("/v2/*proxyPath", ociServerProxy)

	// Set up routes for the registry viewer
	router.GET("/viewer", serveUI)
	router.GET("/viewer/*proxyPath", serveUI)
	// Static content not available under /viewer that the registry viewer needs
	router.Static("/images", "/app/public/images")
	router.StaticFile("/manifest.json/", "/app/public/manifest.json")

	// Serve static content for stacks
	router.Static("/stacks", stacksPath)

	router.Run(":8080")
}

// serveRootEndpoint sets up the handler for the root (/) endpoint on the server
// If html is requested (i.e. from a web browser), the viewer is displayed, otherwise the devfile index is served.
func serveRootEndpoint(c *gin.Context) {
	// Determine if text/html was requested by the client
	acceptHeader := c.Request.Header.Values("Accept")
	if isHtmlRequested(acceptHeader) {
		c.Redirect(http.StatusFound, "/viewer")
	} else {
		serveDevfileIndex(c)
	}
}

func isHtmlRequested(acceptHeader []string) bool {
	for _, header := range acceptHeader {
		if strings.Contains(header, "text/html") {
			return true
		}
	}
	return false
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

	indexType := "stack"
	iconType := c.Query("icon")

	// Serve the index.json file
	buildIndexAPIResponse(c, indexType, iconType)
}

// serveDevfileIndexWithType returns the index file content with specific devfile type
func serveDevfileIndexWithType(c *gin.Context) {
	indexType := c.Param("type")
	iconType := c.Query("icon")

	// Serve the index with type
	buildIndexAPIResponse(c, indexType, iconType)
}

// serveHealthCheck serves endpoint `/health` for registry health check
func serveHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "the server is up and running",
	})
}

// serveDevfile returns the devfile content
func serveDevfile(c *gin.Context) {
	name := c.Param("name")

	var index []indexSchema.Schema
	bytes, err := ioutil.ReadFile(indexPath)
	if err != nil {
		log.Print(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  err.Error(),
			"status": fmt.Sprintf("failed to pull the devfile of %s", name),
		})
		return
	}
	err = json.Unmarshal(bytes, &index)
	if err != nil {
		log.Print(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  err.Error(),
			"status": fmt.Sprintf("failed to pull the devfile of %s", name),
		})
		return
	}
	for _, devfileIndex := range index {
		if devfileIndex.Name == name {
			var bytes []byte
			if devfileIndex.Type == indexSchema.StackDevfileType {
				bytes, err = pullStackFromRegistry(devfileIndex)
			} else {
				// Retrieve the sample devfile stored under /registry/samples/<devfile>
				sampleDevfilePath := path.Join(samplesPath, devfileIndex.Name, devfileName)
				if _, err = os.Stat(sampleDevfilePath); err == nil {
					bytes, err = ioutil.ReadFile(sampleDevfilePath)
				}
			}
			if err != nil {
				log.Print(err.Error())
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":  err.Error(),
					"status": fmt.Sprintf("failed to pull the devfile of %s", name),
				})
				return
			}

			// Track event for telemetry
			if enableTelemetry {
				user := getUser(c)

				err := trackEvent(analytics.Track{
					Event:  eventTrackMap["view"],
					UserId: user,
					Properties: analytics.NewProperties().
						Set("name", name).
						Set("type", string(devfileIndex.Type)).
						Set("registry", registry),
				})
				if err != nil {
					log.Println(err)
				}
			}
			c.Data(http.StatusOK, http.DetectContentType(bytes), bytes)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"status": fmt.Sprintf("the devfile of %s didn't exist", name),
	})
}

// ociServerProxy forwards all GET requests on /v2 to the OCI registry server
func ociServerProxy(c *gin.Context) {
	remote, err := url.Parse(scheme + "://" + registryService + "/v2")
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)

	// Set up the request to the proxy
	// Track event for telemetry
	if enableTelemetry {
		proxyPath := c.Param("proxyPath")
		if proxyPath != "" {
			var name string
			var resource string
			parts := strings.Split(proxyPath, "/")
			// Check proxyPath (e.g. /devfile-catalog/java-quarkus/blobs/sha256:d913cab108c3bc1bd06ce61f1e0cdb6eea2222a7884378f7e656fa26249990b9)
			if len(parts) == 5 {
				name = parts[2]
				resource = parts[3]
			}

			if resource == "blobs" {
				user := getUser(c)

				err := trackEvent(analytics.Track{
					Event:  eventTrackMap["download"],
					UserId: user,
					Properties: analytics.NewProperties().
						Set("name", name).
						Set("registry", registry),
				})
				if err != nil {
					log.Println(err.Error())
				}
			}
		}
	}

	proxy.Director = func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", remote.Host)
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

func serveUI(c *gin.Context) {
	remote, err := url.Parse(scheme + "://" + viewerService + "/viewer/")
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)

	// Set up the request to the proxy
	// This is a good place to set up telemetry for requests to the OCI server (e.g. by parsing the path)
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", remote.Host)
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

// encodeIndexIconToBase64 encodes all index icons to base64 format given the index file path
func encodeIndexIconToBase64(indexPath string, base64IndexPath string) ([]byte, error) {
	// load index
	bytes, err := ioutil.ReadFile(indexPath)
	if err != nil {
		return nil, err
	}
	var index []indexSchema.Schema
	err = json.Unmarshal(bytes, &index)
	if err != nil {
		return nil, err
	}

	// encode all index icons to base64 format
	for i, indexEntry := range index {
		if indexEntry.Icon != "" {
			base64Icon, err := encodeToBase64(indexEntry.Icon)
			index[i].Icon = base64Icon
			if err != nil {
				return nil, err
			}
		}
	}
	err = indexLibrary.CreateIndexFile(index, base64IndexPath)
	if err != nil {
		return nil, err
	}
	bytes, err = json.MarshalIndent(&index, "", "  ")
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// encodeToBase64 encodes the content from the given uri to base64 format
func encodeToBase64(uri string) (string, error) {
	url, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	// load the content from the given uri
	var bytes []byte
	if url.Scheme == "http" || url.Scheme == "https" {
		resp, err := http.Get(uri)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		bytes, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
	} else {
		bytes, err = ioutil.ReadFile(uri)
		if err != nil {
			return "", err
		}
	}

	// encode the content to base64 format
	var base64Encoding string
	mimeType := http.DetectContentType(bytes)
	switch mimeType {
	case "image/jpeg":
		base64Encoding += "data:image/jpeg;base64,"
	case "image/png":
		base64Encoding += "data:image/png;base64,"
	default:
		base64Encoding += "data:image/svg+xml;base64,"
	}
	base64Encoding += base64.StdEncoding.EncodeToString(bytes)
	return base64Encoding, nil
}

// buildIndexAPIResponse builds the response of the REST API of getting the devfile index
func buildIndexAPIResponse(c *gin.Context, indexType string, iconType string) {
	var responseIndexPath string
	var responseBase64IndexPath string
	switch indexType {
	case string(indexSchema.StackDevfileType):
		responseIndexPath = stackIndexPath
		responseBase64IndexPath = stackBase64IndexPath
	case string(indexSchema.SampleDevfileType):
		responseIndexPath = sampleIndexPath
		responseBase64IndexPath = sampleBase64IndexPath
	case "all":
		responseIndexPath = indexPath
		responseBase64IndexPath = base64IndexPath
	default:
		c.JSON(http.StatusNotFound, gin.H{
			"status": fmt.Sprintf("the devfile with %s type didn't exist", indexType),
		})
		return
	}
	if iconType != "" {
		if iconType == encodeFormat {
			if _, err := os.Stat(responseBase64IndexPath); os.IsNotExist(err) {
				bytes, err := encodeIndexIconToBase64(responseIndexPath, responseBase64IndexPath)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"status": fmt.Sprintf("failed to encode %s icons to base64 format: %v", indexType, err),
					})
					return
				}
				c.Data(http.StatusOK, http.DetectContentType(bytes), bytes)
			} else {
				c.File(responseBase64IndexPath)
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": fmt.Sprintf("the icon type %s is not supported", iconType),
			})
			return
		}
	} else {
		c.File(responseIndexPath)
	}

	// Track event for telemetry
	if enableTelemetry {
		user := getUser(c)

		err := trackEvent(analytics.Track{
			Event:  eventTrackMap["list"],
			UserId: user,
			Properties: analytics.NewProperties().
				Set("type", indexType).
				Set("registry", registry),
		})
		if err != nil {
			log.Println(err)
		}
	}
}

// trackEvent tracks event for telemetry
func trackEvent(event analytics.Message) error {
	// Initialize client for telemetry
	client := analytics.New(telemetryKey)
	defer client.Close()

	err := client.Enqueue(event)
	if err != nil {
		return err
	}
	return nil
}

// getUser gets the user
func getUser(c *gin.Context) string {
	user := defaultUser
	if len(c.Request.Header["User"]) != 0 {
		user = c.Request.Header["User"][0]
	}
	return user
}

// getOptionalEnv gets the optional environment variable
func getOptionalEnv(key string, defaultValue interface{}) interface{} {
	if value, present := os.LookupEnv(key); present {
		switch defaultValue.(type) {
		case bool:
			boolValue, err := strconv.ParseBool(value)
			if err != nil {
				log.Print(err)
			}
			return boolValue

		case int:
			intValue, err := strconv.Atoi(value)
			if err != nil {
				log.Print(err)
			}
			return intValue

		default:
			return value
		}
	}
	return defaultValue
}
