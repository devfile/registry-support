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

package ocitest

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

// ResponseError repersents an error returned in an errors response by an OCI server,
// see https://github.com/opencontainers/distribution-spec/blob/main/spec.md#error-codes
type ResponseError struct {
	Code    string                 `json:"code"`    // Error code
	Message string                 `json:"message"` // Error Message
	Detail  map[string]interface{} `json:"detail"`  // Additional detail on the error (optional)
}

// MockOCIServer is an entity for mocking an OCI server
// for testing. At the moment, this is only needed for
// the devfile registry index server endpoint testing,
// however, this entity could be used in a testing scenario
// where an OCI server is needed.
//
// More on the OCI server specification, see https://github.com/opencontainers/distribution-spec/blob/main/spec.md
type MockOCIServer struct {
	httpserver    *httptest.Server     // Test server entity
	router        *gin.Engine          // Router engine for route management
	ServeManifest func(c *gin.Context) // Handler for serving a manifest for a blob
	ServeBlob     func(c *gin.Context) // Handler for serving a blob from the OCI server
}

// servePing is a custom handler to test if
// MockOCIServer is listening for requests
func servePing(c *gin.Context) {
	data, err := json.Marshal(gin.H{
		"message": "ok",
	})
	if err != nil {
		log.Fatal(err)
	}

	c.JSON(http.StatusOK, data)
}

// WriteErrors writes error response object for OCI server
// errors
func WriteErrors(errors []ResponseError) map[string]interface{} {
	return gin.H{
		"errors": errors,
	}
}

// NewMockOCIServer creates a MockOCIServer entity
func NewMockOCIServer() *MockOCIServer {
	gin.SetMode(gin.TestMode)

	mockOCIServer := &MockOCIServer{
		// Create router engine of mock OCI server
		router: gin.Default(),
	}

	// Create mock OCI server using the router engine
	mockOCIServer.httpserver = httptest.NewUnstartedServer(mockOCIServer.router)

	return mockOCIServer
}

// Start listening on listenAddr for requests to the MockOCIServer
func (server *MockOCIServer) Start(listenAddr string) error {
	// Testing Route for checking mock OCI server
	server.router.GET("/v2/ping", servePing)

	// Pull Routes, see https://github.com/opencontainers/distribution-spec/blob/main/spec.md#pull
	// Fetch manifest routes
	if server.ServeManifest != nil {
		server.router.GET("/v2/devfile-catalog/:name/manifests/:ref", server.ServeManifest)
		server.router.HEAD("/v2/devfile-catalog/:name/manifests/:ref", server.ServeManifest)
	}

	// Fetch blob routes
	if server.ServeBlob != nil {
		server.router.GET("/v2/devfile-catalog/:name/blobs/:digest", server.ServeBlob)
		server.router.HEAD("/v2/devfile-catalog/:name/blobs/:digest", server.ServeBlob)
	}

	l, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("unexpected error while creating listener: %v", err)
	}

	server.httpserver.Listener.Close()
	server.httpserver.Listener = l

	server.httpserver.Start()

	return nil
}

// Close the MockOCIServer connection
func (server *MockOCIServer) Close() {
	server.httpserver.Close()
}

// ProxyRecorder is an extension of the ResponseRecorder
// struct within httptest with an additional receiver CloseNotifier
// which is needed for testing the proxy route to the OCI server
type ProxyRecorder struct {
	*httptest.ResponseRecorder
	http.CloseNotifier
}

// NewProxyRecorder creates a new ProxyRecorder entity
func NewProxyRecorder() *ProxyRecorder {
	return &ProxyRecorder{
		ResponseRecorder: httptest.NewRecorder(),
	}
}

// CloseNotify creates a bool channel for notifying a
// closure of a request
func (rec *ProxyRecorder) CloseNotify() <-chan bool {
	return make(<-chan bool)
}
