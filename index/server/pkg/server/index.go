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
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/devfile/registry-support/index/server/pkg/util"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	indexLibrary "github.com/devfile/registry-support/index/generator/library"
	indexSchema "github.com/devfile/registry-support/index/generator/schema"

	_ "github.com/devfile/registry-support/index/server/docs"
	"github.com/gin-gonic/gin"
	"gopkg.in/segmentio/analytics-go.v3"
	"k8s.io/apimachinery/pkg/util/wait"
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

var getIndexLatency = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "index_http_request_duration_seconds",
		Help:    "Latency of index request in seconds.",
		Buckets: prometheus.LinearBuckets(0.5, 0.5, 10),
	},
	[]string{"status"},
)

func ServeRegistry() {
	// Enable metrics
	// Run on a separate port and router from the index server so that it's not exposed publicly

	handler := http.NewServeMux()
	handler.Handle("/metrics", promhttp.Handler())
	prometheus.MustRegister(getIndexLatency)
	indexServer := &http.Server{
		Addr:         ":7071",
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go indexServer.ListenAndServe()

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

		if devfileIndex.Versions != nil && len(devfileIndex.Versions) != 0 {
			for _, versionComponent := range devfileIndex.Versions {
				if len(versionComponent.Resources) != 0 {
					err := pushStackToRegistry(versionComponent, devfileIndex.Name)
					if err != nil {
						log.Fatal(err.Error())
					}
				}
			}
		}
	}
	err = indexLibrary.CreateIndexFile(sampleIndex, sampleIndexPath)
	if err != nil {
		log.Fatalf("failed to generate %s: %v", sampleIndexPath, err)
	}
	err = indexLibrary.CreateIndexFile(stackIndex, stackIndexPath)
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
	router.GET("/index", serveDevfileIndexV1)
	router.GET("/index/:type", serveDevfileIndexV1WithType)
	router.GET("/health", serveHealthCheck)
	router.GET("/devfiles/:name", serveDevfile)
	router.GET("/devfiles/:name/:version", serveDevfileWithVersion)
	router.GET("/devfiles/:name/starter-projects/:starterProjectName", serveDevfileStarterProject)
	router.GET("/devfiles/:name/:version/starter-projects/:starterProjectName", serveDevfileStarterProjectWithVersion)

	// Registry REST APIs for index v2
	router.GET("/v2index", serveDevfileIndexV2)
	router.GET("/v2index/:type", serveDevfileIndexV2WithType)

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
	// Track event for telemetry for GET requests only
	if enableTelemetry && c.Request.Method == http.MethodGet {
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

			//Ignore events from the registry-viewer and DevConsole since those are tracked on the client side
			if resource == "blobs" && !util.IsWebClient(c) {
				user := util.GetUser(c)
				client := util.GetClient(c)

				err := util.TrackEvent(analytics.Track{
					Event:   eventTrackMap["download"],
					UserId:  user,
					Context: util.SetContext(c),
					Properties: analytics.NewProperties().
						Set("name", name).
						Set("registry", registry).
						Set("client", client),
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
