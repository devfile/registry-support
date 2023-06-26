//
// Copyright 2022-2023 Red Hat, Inc.
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
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	indexLibrary "github.com/devfile/registry-support/index/generator/library"
	indexSchema "github.com/devfile/registry-support/index/generator/schema"

	oapiMiddleware "github.com/deepmap/oapi-codegen/pkg/gin-middleware"
	_ "github.com/devfile/registry-support/index/server/docs"
	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/util/wait"
)

var eventTrackMap = map[string]string{
	"list":       "list devfile",
	"view":       "view devfile",
	"download":   "download devfile",
	"spdownload": "Starter Project Downloaded",
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

	// Get OpenAPI spec
	swagger, err := GetSwagger()
	if err != nil {
		log.Fatalf("Error loading OpenAPI spec: %v", err)
	}

	swagger.Servers = nil

	// Create server context
	server := &Server{}

	// Start the server and serve requests and index.json
	router := gin.Default()

	// Register Devfile Registry REST APIs and use OpenAPI validator middleware
	router = RegisterHandlersWithOptions(router, server, GinServerOptions{
		Middlewares: []MiddlewareFunc{
			func(c *gin.Context) {
				oapiMiddleware.OapiRequestValidator(swagger)(c)
			},
		},
	})

	// Set up a simple proxy for /v2 endpoints
	// Only allow HEAD and GET requests
	router.HEAD("/v2/*proxyPath", ServeOciProxy)
	router.GET("/v2/*proxyPath", ServeOciProxy)

	// Set up routes for the registry viewer
	router.GET("/viewer", ServeUI)
	router.GET("/viewer/*proxyPath", ServeUI)

	// Serve static content for stacks
	router.Static("/stacks", stacksPath)

	router.Run(":8080")
}
