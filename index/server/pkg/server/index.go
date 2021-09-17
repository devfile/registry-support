package server

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	indexLibrary "github.com/devfile/registry-support/index/generator/library"
	indexSchema "github.com/devfile/registry-support/index/generator/schema"

	_ "github.com/devfile/registry-support/index/server/docs"
	"github.com/gin-gonic/gin"
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

func ServeRegistry() {
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
