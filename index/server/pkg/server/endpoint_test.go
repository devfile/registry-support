//
// Copyright Red Hat
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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/devfile/library/v2/pkg/devfile/parser"
	"github.com/devfile/registry-support/index/server/pkg/ocitest"
	"github.com/gin-gonic/gin"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	ociServerIP = "127.0.0.1:5000" // Mock OCI server listen address and port number, format: '<address>:<port>'
)

var (
	// mock manifest repository for OCI server
	manifests = map[string]map[string]ocispec.Manifest{
		"java-maven": {
			"1.1.0": {
				Versioned: specs.Versioned{SchemaVersion: 2},
				Config: ocispec.Descriptor{
					MediaType: devfileConfigMediaType,
				},
				Layers: []ocispec.Descriptor{
					{
						MediaType: devfileMediaType,
						Digest:    "sha256:b81a4a857ebbd6b7093c38703e3b7c6d7a2652abfd55898f82fdea45634fd549",
						Size:      1251,
						Annotations: map[string]string{
							"org.opencontainers.image.title": devfileName,
						},
					},
				},
			},
		},
		"go": {
			"1.1.0": {
				Versioned: specs.Versioned{SchemaVersion: 2},
				Config: ocispec.Descriptor{
					MediaType: devfileConfigMediaType,
				},
				Layers: []ocispec.Descriptor{
					{
						MediaType: devfileMediaType,
						Digest:    "sha256:4cad7c1629ba848245205a08107b296adba307f77ca3635b16920473589cb12e",
						Size:      1086,
						Annotations: map[string]string{
							"org.opencontainers.image.title": devfileName,
						},
					},
				},
			},
			"1.2.0": {
				Versioned: specs.Versioned{SchemaVersion: 2},
				Config: ocispec.Descriptor{
					MediaType: devfileConfigMediaType,
				},
				Layers: []ocispec.Descriptor{
					{
						MediaType: devfileMediaType,
						Digest:    "sha256:bb4c6b96292bbcd48f445436f7945399a4d314b111ee976d6235199e854bfb68",
						Size:      1091,
						Annotations: map[string]string{
							"org.opencontainers.image.title": devfileName,
						},
					},
				},
			},
		},
		"java-quarkus": {
			"1.1.0": {
				Versioned: specs.Versioned{SchemaVersion: 2},
				Config: ocispec.Descriptor{
					MediaType: devfileConfigMediaType,
				},
				Layers: []ocispec.Descriptor{
					{
						MediaType: devfileMediaType,
						Digest:    "sha256:6143ffeab064d0b9a6991a6e1025a1b63d01be87e197fabb8a34d5a9fc3fcba0",
						Size:      1935,
						Annotations: map[string]string{
							"org.opencontainers.image.title": devfileName,
						},
					},
				},
			},
		},
		"java-wildfly": {
			"1.0.0": {
				Versioned: specs.Versioned{SchemaVersion: 2},
				Config: ocispec.Descriptor{
					MediaType: devfileConfigMediaType,
				},
				Layers: []ocispec.Descriptor{
					{
						MediaType: devfileMediaType,
						Digest:    "sha256:a6363457f7603259b7c55c5d5752adcf1cfb146227e90890a3ca8aa6b71879dd",
						Size:      7151,
						Annotations: map[string]string{
							"org.opencontainers.image.title": devfileName,
						},
					},
				},
			},
		},
	}
)

// serveManifest custom handler for fetching a manifest from the
// mock OCI server
func serveManifest(c *gin.Context) {
	name, ref := c.Param("name"), c.Param("ref")
	var (
		stackManifest ocispec.Manifest
		found         bool
		bytes         []byte
		err           error
	)

	if strings.HasPrefix(ref, "sha256:") {
		stackManifests, found := manifests[name]
		if !found {
			notFoundManifest(c, ref)
			return
		}

		found = false
		for _, manifest := range stackManifests {
			dgst, err := digestEntity(manifest)
			if err != nil {
				log.Fatal(err)
			} else if reflect.DeepEqual(ref, dgst) {
				stackManifest = manifest
				found = true
				break
			}
		}

		if !found {
			notFoundManifest(c, ref)
			return
		}
	} else {
		stackManifest, found = manifests[name][ref]

		if !found {
			notFoundManifest(c, ref)
			return
		}
	}

	if c.Request.Method == http.MethodGet {
		bytes, err = json.Marshal(stackManifest)
		if err != nil {
			log.Fatal(err)
		}
	}

	c.Data(http.StatusOK, ocispec.MediaTypeImageManifest, bytes)
}

// notFoundManifest custom handler for manifest not found status of the mock OCI server
func notFoundManifest(c *gin.Context, tag string) {
	var data gin.H = nil

	if c.Request.Method == http.MethodGet {
		data = ocitest.WriteErrors([]ocitest.ResponseError{
			{
				Code:    "MANIFEST_UNKNOWN",
				Message: "manifest unknown",
				Detail: map[string]interface{}{
					"Tag": tag,
				},
			},
		})
	}

	c.JSON(http.StatusNotFound, data)
}

// digestEntity generates sha256 digest of any entity type
func digestEntity(e interface{}) (string, error) {
	bytes, err := json.Marshal(e)
	if err != nil {
		return "", err
	}

	return digest.FromBytes(bytes).String(), nil
}

// digestFile generates sha256 digest from file contents
func digestFile(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	dgst, err := digest.FromReader(file)
	if err != nil {
		return "", err
	}

	return dgst.String(), nil
}

// serveBlob custom handler for fetching a blob from the
// mock OCI server
func serveBlob(c *gin.Context) {
	name, dgst := c.Param("name"), c.Param("digest")
	stackRoot := filepath.Join(stacksPath, name)
	var (
		blobPath string
		found    bool
		err      error
	)

	found = false
	err = filepath.WalkDir(stackRoot, func(path string, d fs.DirEntry, err error) error {
		var fdgst string

		if err != nil {
			return err
		}

		if found || d.IsDir() {
			return nil
		}

		fdgst, err = digestFile(path)
		if err != nil {
			return err
		} else if reflect.DeepEqual(dgst, fdgst) {
			blobPath = path
			found = true
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	} else if !found {
		notFoundBlob(c)
		return
	}

	file, err := os.Open(blobPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	c.Data(http.StatusOK, http.DetectContentType(bytes), bytes)
}

// notFoundBlob custom handler for blob not found status of the mock OCI server
func notFoundBlob(c *gin.Context) {
	c.Data(http.StatusNotFound, "plain/text", []byte{})
}

// setupMockOCIServer sets up mock OCI server for endpoint testing environment
func setupMockOCIServer() (func(), error) {
	mockOCIServer := ocitest.NewMockOCIServer()

	// Pull Routes
	mockOCIServer.ServeManifest = serveManifest
	mockOCIServer.ServeBlob = serveBlob

	if err := mockOCIServer.Start(ociServerIP); err != nil {
		return nil, err
	}

	return mockOCIServer.Close, nil
}

// setupVars sets up registry index server global variables for endpoint testing environment
func setupVars() {
	var registryPath string

	if _, found := os.LookupEnv("DEVFILE_REGISTRY"); found {
		registryPath = os.Getenv("DEVFILE_REGISTRY")
	} else {
		registryPath = "../../tests/registry"
	}

	if stacksPath == "" {
		stacksPath = filepath.Join(registryPath, "stacks")
	}
	if samplesPath == "" {
		samplesPath = filepath.Join(registryPath, "samples")
	}
	if indexPath == "" {
		indexPath = filepath.Join(registryPath, "index_main.json")
	}
	if sampleIndexPath == "" {
		sampleIndexPath = filepath.Join(registryPath, "index_extra.json")
	}
	if stackIndexPath == "" {
		stackIndexPath = filepath.Join(registryPath, "index_registry.json")
	}
}

// TestMockOCIServer tests if MockOCIServer is listening for
// requests using the custom '/v2/ping' route
func TestMockOCIServer(t *testing.T) {
	mockOCIServer := ocitest.NewMockOCIServer()
	if err := mockOCIServer.Start(ociServerIP); err != nil {
		t.Errorf("Failed to setup mock OCI server: %v", err)
		return
	}
	defer mockOCIServer.Close()
	setupVars()

	resp, err := http.Get(fmt.Sprintf("http://%s", filepath.Join(ociServerIP, "/v2/ping")))
	if err != nil {
		t.Errorf("Error in request: %v", err)
		return
	}

	if !reflect.DeepEqual(resp.StatusCode, http.StatusOK) {
		t.Errorf("Did not get expected status code, Got: %v, Expected: %v", resp.StatusCode, http.StatusOK)
	}
}

// TestServeHealthCheck tests health check endpoint '/health'
func TestServeHealthCheck(t *testing.T) {
	var got gin.H

	gin.SetMode(gin.TestMode)

	setupVars()

	server := &Server{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	server.ServeHealthCheck(c)

	wantStatusCode := http.StatusOK
	if gotStatusCode := w.Code; !reflect.DeepEqual(gotStatusCode, wantStatusCode) {
		t.Errorf("Did not get expected status code, Got: %v, Expected: %v", gotStatusCode, wantStatusCode)
		return
	}

	wantContentType := "application/json"
	header := w.Header()
	if gotContentType := strings.Split(header.Get("Content-Type"), ";")[0]; !reflect.DeepEqual(gotContentType, wantContentType) {
		t.Errorf("Did not get expected content type, Got: %v, Expected: %v", gotContentType, wantContentType)
		return
	}

	bytes, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatalf("Did not expect error: %v", err)
		return
	}

	if err = json.Unmarshal(bytes, &got); err != nil {
		t.Fatalf("Did not expect error: %v", err)
		return
	}

	wantMessage := "the server is up and running"
	gotMessage, found := got["message"]
	if !found {
		t.Error("Did not get any body or message.")
		return
	} else if !reflect.DeepEqual(gotMessage, wantMessage) {
		t.Errorf("Did not get expected body or message, Got: %v, Expected: %v", gotMessage, wantMessage)
		return
	}
}

// TestServeDevfileIndexV1 tests '/index' endpoint
func TestServeDevfileIndexV1(t *testing.T) {
	const wantStatusCode = http.StatusOK

	setupVars()

	gin.SetMode(gin.TestMode)

	server := &Server{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	server.ServeDevfileIndexV1(c, ServeDevfileIndexV1Params{})

	if gotStatusCode := w.Code; !reflect.DeepEqual(gotStatusCode, wantStatusCode) {
		t.Errorf("Did not get expected status code, Got: %v, Expected: %v", gotStatusCode, wantStatusCode)
		return
	}
}

// TestServeDevfileIndexV1WithType tests '/index/:type' endpoint
func TestServeDevfileIndexV1WithType(t *testing.T) {
	setupVars()
	tests := []struct {
		name     string
		params   gin.Params
		wantCode int
	}{
		{
			name: "GET /index/stack - Successful Response Test",
			params: gin.Params{
				gin.Param{Key: "type", Value: "stack"},
			},
			wantCode: http.StatusOK,
		},
		{
			name: "GET /index/sample - Successful Response Test",
			params: gin.Params{
				gin.Param{Key: "type", Value: "sample"},
			},
			wantCode: http.StatusOK,
		},
		{
			name: "GET /index/all - Successful Response Test",
			params: gin.Params{
				gin.Param{Key: "type", Value: "all"},
			},
			wantCode: http.StatusOK,
		},
		{
			name: "GET /index/notatype - Type Not Found Response Test",
			params: gin.Params{
				gin.Param{Key: "type", Value: "notatype"},
			},
			wantCode: http.StatusNotFound,
		},
	}
	server := &Server{}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			gin.SetMode(gin.TestMode)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Params = append(c.Params, test.params...)

			indexType, _ := c.Params.Get("type")

			server.ServeDevfileIndexV1WithType(c, indexType, ServeDevfileIndexV1WithTypeParams{})

			if gotStatusCode := w.Code; !reflect.DeepEqual(gotStatusCode, test.wantCode) {
				t.Errorf("Did not get expected status code, Got: %v, Expected: %v", gotStatusCode, test.wantCode)
				return
			}
		})
	}
}

// TestServeDevfileIndexV2 tests '/v2index' endpoint
func TestServeDevfileIndexV2(t *testing.T) {
	const wantStatusCode = http.StatusOK

	setupVars()

	gin.SetMode(gin.TestMode)

	server := &Server{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	server.ServeDevfileIndexV2(c, ServeDevfileIndexV2Params{})

	if gotStatusCode := w.Code; !reflect.DeepEqual(gotStatusCode, wantStatusCode) {
		t.Errorf("Did not get expected status code, Got: %v, Expected: %v", gotStatusCode, wantStatusCode)
		return
	}
}

// TestServeDevfileIndexV2 tests '/v2index/:type' endpoint
func TestServeDevfileIndexV2WithType(t *testing.T) {
	setupVars()
	tests := []struct {
		name     string
		params   gin.Params
		query    url.Values
		wantCode int
	}{
		{
			name: "GET /v2index/stack - Successful Response Test",
			params: gin.Params{
				gin.Param{Key: "type", Value: "stack"},
			},
			wantCode: http.StatusOK,
		},
		{
			name: "GET /v2index/sample - Successful Response Test",
			params: gin.Params{
				gin.Param{Key: "type", Value: "sample"},
			},
			wantCode: http.StatusOK,
		},
		{
			name: "GET /v2index/all - Successful Response Test",
			params: gin.Params{
				gin.Param{Key: "type", Value: "all"},
			},
			wantCode: http.StatusOK,
		},
		{
			name: "GET /v2index/all?minSchemaVersion=2.1.0&maxSChemaVersion=2.2 - Successful Response Test",
			params: gin.Params{
				gin.Param{Key: "type", Value: "all"},
			},
			query: url.Values{
				"minSchemaVersion": []string{"2.1.0"},
				"maxSchemaVersion": []string{"2.2"},
			},
			wantCode: http.StatusOK,
		},
		{
			name: "GET /v2index/all?minSchemaVersion=1.0&maxSChemaVersion=2.2 - Bad Request Response Test",
			params: gin.Params{
				gin.Param{Key: "type", Value: "all"},
			},
			query: url.Values{
				"minSchemaVersion": []string{"1.0"},
				"maxSchemaVersion": []string{"2.2"},
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "GET /v2index/all?minSchemaVersion=2.0.0.0&maxSChemaVersion=2.2 - Bad Request Response Test",
			params: gin.Params{
				gin.Param{Key: "type", Value: "all"},
			},
			query: url.Values{
				"minSchemaVersion": []string{"2.0.0.0"},
				"maxSchemaVersion": []string{"2.2"},
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "GET /v2index/all?minSchemaVersion=2.0.0&maxSChemaVersion=test - Bad Request Response Test",
			params: gin.Params{
				gin.Param{Key: "type", Value: "all"},
			},
			query: url.Values{
				"minSchemaVersion": []string{"2.0.0"},
				"maxSchemaVersion": []string{"test"},
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "GET /v2index/notatype - Type Not Found Response Test",
			params: gin.Params{
				gin.Param{Key: "type", Value: "notatype"},
			},
			wantCode: http.StatusNotFound,
		},
	}
	server := &Server{}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			gin.SetMode(gin.TestMode)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request = httptest.NewRequest(http.MethodGet, "/v2index", nil)
			c.Params = append(c.Params, test.params...)
			c.Request.URL.RawQuery = test.query.Encode()

			indexType, _ := c.Params.Get("type")

			server.ServeDevfileIndexV2WithType(c, indexType, ServeDevfileIndexV2WithTypeParams{})

			if gotStatusCode := w.Code; !reflect.DeepEqual(gotStatusCode, test.wantCode) {
				t.Errorf("Did not get expected status code, Got: %v, Expected: %v", gotStatusCode, test.wantCode)
				return
			}
		})
	}
}

// TestServeDevfile tests '/devfiles/:name' endpoint
func TestServeDevfile(t *testing.T) {
	tests := []struct {
		name              string
		params            gin.Params
		wantCode          int
		wantSchemaVersion string
		wantError         bool
	}{
		{
			name: "GET /devfiles/java-maven - Fetch Java Maven Devfile",
			params: gin.Params{
				gin.Param{Key: "name", Value: "java-maven"},
			},
			wantCode:          http.StatusOK,
			wantSchemaVersion: "2.2.0",
		},
		{
			name: "GET /devfiles/go - Fetch Go Devfile",
			params: gin.Params{
				gin.Param{Key: "name", Value: "go"},
			},
			wantCode:          http.StatusOK,
			wantSchemaVersion: "2.0.0",
		},
		{
			name: "GET /devfiles/not-exist - Fetch Non-Existent Devfile",
			params: gin.Params{
				gin.Param{Key: "name", Value: "not-exist"},
			},
			wantCode:  http.StatusNotFound,
			wantError: true,
		},
	}

	closeServer, err := setupMockOCIServer()
	if err != nil {
		t.Errorf("Did not setup mock OCI server properly: %v", err)
		return
	}
	defer closeServer()
	setupVars()
	server := &Server{}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			gin.SetMode(gin.TestMode)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Params = append(c.Params, test.params...)

			name, _ := c.Params.Get("name")
			server.ServeDevfile(c, name)

			if gotStatusCode := w.Code; !reflect.DeepEqual(gotStatusCode, test.wantCode) {
				t.Errorf("Did not get expected status code, Got: %v, Expected: %v", gotStatusCode, test.wantCode)
			} else if !test.wantError {
				bytes := w.Body.Bytes()
				content, err := parser.ParseFromData(bytes)
				if err != nil {
					t.Fatalf("Did not expect error: %v", err)
				}

				if gotSchemaVersion := content.Data.GetSchemaVersion(); !reflect.DeepEqual(gotSchemaVersion, test.wantSchemaVersion) {
					t.Errorf("Did not get expected schema version, Got: %v, Expected: %v", gotSchemaVersion, test.wantSchemaVersion)
				}
			}
		})
	}
}

// TestServeDevfileWithVersion tests '/devfiles/:name/:version' endpoint
func TestServeDevfileWithVersion(t *testing.T) {
	tests := []struct {
		name              string
		params            gin.Params
		query             url.Values
		wantCode          int
		wantSchemaVersion string
		wantError         bool
	}{
		{
			name: "GET /devfiles/go/default - Fetch Go Devfile With Default Stack Version",
			params: gin.Params{
				gin.Param{Key: "name", Value: "go"},
				gin.Param{Key: "version", Value: "default"},
			},
			wantCode:          http.StatusOK,
			wantSchemaVersion: "2.0.0",
		},
		{
			name: "GET /devfiles/go/latest - Fetch Go Devfile With Latest Stack Version",
			params: gin.Params{
				gin.Param{Key: "name", Value: "go"},
				gin.Param{Key: "version", Value: "latest"},
			},
			wantCode:          http.StatusOK,
			wantSchemaVersion: "2.1.0",
		},
		{
			name: "GET /devfiles/go/latest?minSchemaVersion=2.1 - Fetch Go Devfile With Latest Devfile 2.1.0 Stack Version",
			params: gin.Params{
				gin.Param{Key: "name", Value: "go"},
				gin.Param{Key: "version", Value: "latest"},
			},
			query: url.Values{
				"minSchemaVersion": []string{"2.1"},
			},
			wantCode:          http.StatusOK,
			wantSchemaVersion: "2.1.0",
		},
		{
			name: "GET /devfiles/go/latest?maxSchemaVersion=2.0.0 - Fetch Go Devfile With Latest Devfile 2.0.0 Stack Version",
			params: gin.Params{
				gin.Param{Key: "name", Value: "go"},
				gin.Param{Key: "version", Value: "latest"},
			},
			query: url.Values{
				"maxSchemaVersion": []string{"2.0.0"},
			},
			wantCode:          http.StatusOK,
			wantSchemaVersion: "2.0.0",
		},
		{
			name: "GET /devfiles/go/latest?maxSchemaVersion=1.0 - Invalid Schema Version Fetch Go Devfile With Latest Stack Version",
			params: gin.Params{
				gin.Param{Key: "name", Value: "go"},
				gin.Param{Key: "version", Value: "latest"},
			},
			query: url.Values{
				"maxSchemaVersion": []string{"1.0"},
			},
			wantCode:  http.StatusBadRequest,
			wantError: true,
		},
		{
			name: "GET /devfiles/go/latest?minSchemaVersion=test - Invalid Schema Version Fetch Go Devfile With Latest Stack Version",
			params: gin.Params{
				gin.Param{Key: "name", Value: "go"},
				gin.Param{Key: "version", Value: "latest"},
			},
			query: url.Values{
				"minSchemaVersion": []string{"test"},
			},
			wantCode:  http.StatusBadRequest,
			wantError: true,
		},
		{
			name: "GET /devfiles/go/1.2.0 - Fetch Go Devfile With Specific Version",
			params: gin.Params{
				gin.Param{Key: "name", Value: "go"},
				gin.Param{Key: "version", Value: "1.2.0"},
			},
			wantCode:          http.StatusOK,
			wantSchemaVersion: "2.1.0",
		},
		{
			name: "GET /devfiles/not-exist/latest - Fetch Non-Existent Devfile With Latest Stack Version",
			params: gin.Params{
				gin.Param{Key: "name", Value: "not-exist"},
				gin.Param{Key: "version", Value: "latest"},
			},
			wantCode:  http.StatusNotFound,
			wantError: true,
		},
		{
			name: "GET /devfiles/java-maven/not-exist - Fetch Java Maven Devfile With Non-Existent Stack Version",
			params: gin.Params{
				gin.Param{Key: "name", Value: "java-maven"},
				gin.Param{Key: "version", Value: "non-exist"},
			},
			wantCode:  http.StatusNotFound,
			wantError: true,
		},
	}

	closeServer, err := setupMockOCIServer()
	if err != nil {
		t.Errorf("Did not setup mock OCI server properly: %v", err)
		return
	}
	defer closeServer()
	setupVars()
	server := &Server{}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			gin.SetMode(gin.TestMode)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request = httptest.NewRequest(http.MethodGet, "/devfiles", nil)
			c.Params = append(c.Params, test.params...)
			c.Request.URL.RawQuery = test.query.Encode()

			name, _ := c.Params.Get("name")
			version, _ := c.Params.Get("version")
			server.ServeDevfileWithVersion(c, name, version)

			if gotStatusCode := w.Code; !reflect.DeepEqual(gotStatusCode, test.wantCode) {
				t.Errorf("Did not get expected status code, Got: %v, Expected: %v", gotStatusCode, test.wantCode)
			} else if !test.wantError {
				bytes := w.Body.Bytes()
				content, err := parser.ParseFromData(bytes)
				if err != nil {
					t.Fatalf("Did not expect error: %v", err)
				}

				if gotSchemaVersion := content.Data.GetSchemaVersion(); !reflect.DeepEqual(gotSchemaVersion, test.wantSchemaVersion) {
					t.Errorf("Did not get expected schema version, Got: %v, Expected: %v", gotSchemaVersion, test.wantSchemaVersion)
				}
			}
		})
	}
}

// TestServeDevfileStarterProject tests '/devfiles/:name/starter-projects/:starterProjectName' endpoint
func TestServeDevfileStarterProject(t *testing.T) {
	const wantContentType = starterProjectMediaType
	tests := []struct {
		name      string
		params    gin.Params
		wantCode  int
		wantError bool
	}{
		{
			name: "GET /devfiles/java-maven/starter-projects/springbootproject - Fetch Java Maven 'springbootproject' Starter Project",
			params: gin.Params{
				gin.Param{Key: "name", Value: "java-maven"},
				gin.Param{Key: "starterProjectName", Value: "springbootproject"},
			},
			wantCode: http.StatusAccepted,
		},
		{
			name: "GET /devfiles/go/starter-projects/go-starter - Fetch Go 'go-starter' Starter Project",
			params: gin.Params{
				gin.Param{Key: "name", Value: "go"},
				gin.Param{Key: "starterProjectName", Value: "go-starter"},
			},
			wantCode: http.StatusAccepted,
		},
		{
			name: "GET /devfiles/java-quarkus/starter-projects/community - Fetch Java Quarkus 'community' Starter Project",
			params: gin.Params{
				gin.Param{Key: "name", Value: "java-quarkus"},
				gin.Param{Key: "starterProjectName", Value: "community"},
			},
			wantCode: http.StatusAccepted,
		},
		{
			name: "GET /devfiles/java-wildfly/starter-projects/microprofile-config - Fetch Java Wildfly 'microprofile-config' Starter Project",
			params: gin.Params{
				gin.Param{Key: "name", Value: "java-wildfly"},
				gin.Param{Key: "starterProjectName", Value: "microprofile-config"},
			},
			wantCode: http.StatusAccepted,
		},
		{
			name: "GET /devfiles/java-wildfly/starter-projects/microprofile-jwt - Fetch Java Wildfly 'microprofile-jwt' Starter Project",
			params: gin.Params{
				gin.Param{Key: "name", Value: "java-wildfly"},
				gin.Param{Key: "starterProjectName", Value: "microprofile-jwt"},
			},
			wantCode: http.StatusAccepted,
		},
		{
			name: "GET /devfiles/not-exist/starter-projects/some - Fetch 'some' starter project from Non-Existent stack",
			params: gin.Params{
				gin.Param{Key: "name", Value: "not-exist"},
				gin.Param{Key: "starterProjectName", Value: "some"},
			},
			wantCode:  http.StatusNotFound,
			wantError: true,
		},
		{
			name: "GET /devfiles/java-maven/starter-projects/not-exist - Fetch Non-Existent starter project from Java Maven stack",
			params: gin.Params{
				gin.Param{Key: "name", Value: "java-maven"},
				gin.Param{Key: "starterProjectName", Value: "not-exist"},
			},
			wantCode:  http.StatusNotFound,
			wantError: true,
		},
	}

	closeServer, err := setupMockOCIServer()
	if err != nil {
		t.Errorf("Did not setup mock OCI server properly: %v", err)
		return
	}
	defer closeServer()
	setupVars()
	server := &Server{}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			gin.SetMode(gin.TestMode)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Params = append(c.Params, test.params...)

			name, _ := c.Params.Get("name")
			starterProject, _ := c.Params.Get("starterProjectName")

			server.ServeDevfileStarterProject(c, name, starterProject)

			if gotStatusCode := w.Code; !reflect.DeepEqual(gotStatusCode, test.wantCode) {
				t.Errorf("Did not get expected status code, Got: %v, Expected: %v", gotStatusCode, test.wantCode)
			} else if !test.wantError {
				gotContentType := http.DetectContentType(w.Body.Bytes())
				if !reflect.DeepEqual(gotContentType, wantContentType) {
					t.Errorf("Did not get expected content-type, Got: %v, Expected: %v", gotContentType, wantContentType)
				}
			}
		})
	}
}

// TestServeDevfileStarterProjectWithVersion tests '/devfiles/:name/:version/starter-projects/:starterProjectName' endpoint
func TestServeDevfileStarterProjectWithVersion(t *testing.T) {
	const wantContentType = starterProjectMediaType
	tests := []struct {
		name      string
		params    gin.Params
		wantCode  int
		wantError bool
	}{
		{
			name: "GET /devfiles/go/default/starter-projects/go-starter - Fetch Go 'go-starter' Starter Project With Default Version",
			params: gin.Params{
				gin.Param{Key: "name", Value: "go"},
				gin.Param{Key: "version", Value: "default"},
				gin.Param{Key: "starterProjectName", Value: "go-starter"},
			},
			wantCode: http.StatusAccepted,
		},
		{
			name: "GET /devfiles/go/latest/starter-projects/go-starter - Fetch Go 'go-starter' Starter Project With Latest Version",
			params: gin.Params{
				gin.Param{Key: "name", Value: "go"},
				gin.Param{Key: "version", Value: "latest"},
				gin.Param{Key: "starterProjectName", Value: "go-starter"},
			},
			wantCode: http.StatusAccepted,
		},
		{
			name: "GET /devfiles/go/1.2.0/starter-projects/go-starter - Fetch Go 'go-starter' Starter Project With Specific Version",
			params: gin.Params{
				gin.Param{Key: "name", Value: "go"},
				gin.Param{Key: "version", Value: "1.2.0"},
				gin.Param{Key: "starterProjectName", Value: "go-starter"},
			},
			wantCode: http.StatusAccepted,
		},
		{
			name: "GET /devfiles/not-exist/latest/starter-projects/some - " +
				"Fetch 'some' starter project from Non-Existent stack With Latest Version",
			params: gin.Params{
				gin.Param{Key: "name", Value: "not-exist"},
				gin.Param{Key: "version", Value: "latest"},
				gin.Param{Key: "starterProjectName", Value: "some"},
			},
			wantCode:  http.StatusNotFound,
			wantError: true,
		},
		{
			name: "GET /devfiles/java-maven/latest/starter-projects/not-exist - " +
				"Fetch Non-Existent starter project from Java Maven stack With Latest Version",
			params: gin.Params{
				gin.Param{Key: "name", Value: "java-maven"},
				gin.Param{Key: "version", Value: "latest"},
				gin.Param{Key: "starterProjectName", Value: "not-exist"},
			},
			wantCode:  http.StatusNotFound,
			wantError: true,
		},
		{
			name: "GET /devfiles/java-maven/not-exist/starter-projects/springbootproject - " +
				"Fetch Java Maven 'springbootproject' Starter Project With Non-Existent Version",
			params: gin.Params{
				gin.Param{Key: "name", Value: "java-maven"},
				gin.Param{Key: "version", Value: "non-exist"},
				gin.Param{Key: "starterProjectName", Value: "springbootproject"},
			},
			wantCode:  http.StatusNotFound,
			wantError: true,
		},
	}

	closeServer, err := setupMockOCIServer()
	if err != nil {
		t.Errorf("Did not setup mock OCI server properly: %v", err)
		return
	}
	defer closeServer()
	setupVars()
	server := &Server{}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			gin.SetMode(gin.TestMode)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Params = append(c.Params, test.params...)

			name, _ := c.Params.Get("name")
			version, _ := c.Params.Get("version")
			starterProject, _ := c.Params.Get("starterProjectName")

			server.ServeDevfileStarterProjectWithVersion(c, name, version, starterProject)

			if gotStatusCode := w.Code; !reflect.DeepEqual(gotStatusCode, test.wantCode) {
				t.Errorf("Did not get expected status code, Got: %v, Expected: %v", gotStatusCode, test.wantCode)
			} else if !test.wantError {
				gotContentType := http.DetectContentType(w.Body.Bytes())
				if !reflect.DeepEqual(gotContentType, wantContentType) {
					t.Errorf("Did not get expected content-type, Got: %v, Expected: %v", gotContentType, wantContentType)
				}
			}
		})
	}
}

// TestOCIServerProxy tests '/v2/*proxyPath' endpoint
func TestOCIServerProxy(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		url       string
		wantCode  int
		wantError bool
	}{
		{
			name:     "HEAD /v2/devfile-catalog/go/manifests/1.2.0",
			method:   http.MethodHead,
			url:      "/devfile-catalog/go/manifests/1.2.0",
			wantCode: 200,
		},
		{
			name:      "HEAD /v2/devfile-catalog/go/manifests/notfound",
			method:    http.MethodHead,
			url:       "/devfile-catalog/go/manifests/notfound",
			wantCode:  404,
			wantError: true,
		},
		{
			name:     "GET /v2/devfile-catalog/go/manifests/1.2.0",
			method:   http.MethodGet,
			url:      "/devfile-catalog/go/manifests/1.2.0",
			wantCode: 200,
		},
		{
			name:      "GET /v2/devfile-catalog/go/manifests/notfound",
			method:    http.MethodGet,
			url:       "/devfile-catalog/go/manifests/notfound",
			wantCode:  404,
			wantError: true,
		},
		{
			name:     "HEAD /v2/devfile-catalog/go/blobs/sha256:bb4c6b96292bbcd48f445436f7945399a4d314b111ee976d6235199e854bfb68",
			method:   http.MethodHead,
			url:      "/devfile-catalog/go/blobs/sha256:bb4c6b96292bbcd48f445436f7945399a4d314b111ee976d6235199e854bfb68",
			wantCode: 200,
		},
		{
			name:      "HEAD /v2/devfile-catalog/go/blobs/notfound",
			method:    http.MethodHead,
			url:       "/devfile-catalog/go/blobs/notfound",
			wantCode:  404,
			wantError: true,
		},
		{
			name:     "GET /v2/devfile-catalog/go/blobs/sha256:bb4c6b96292bbcd48f445436f7945399a4d314b111ee976d6235199e854bfb68",
			method:   http.MethodGet,
			url:      "/devfile-catalog/go/blobs/sha256:bb4c6b96292bbcd48f445436f7945399a4d314b111ee976d6235199e854bfb68",
			wantCode: 200,
		},
		{
			name:      "GET /v2/devfile-catalog/go/blobs/notfound",
			method:    http.MethodGet,
			url:       "/devfile-catalog/go/blobs/notfound",
			wantCode:  404,
			wantError: true,
		},
	}

	closeServer, err := setupMockOCIServer()
	if err != nil {
		t.Errorf("Did not setup mock OCI server properly: %v", err)
		return
	}
	defer closeServer()
	setupVars()

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			gin.SetMode(gin.TestMode)

			w := ocitest.NewProxyRecorder()
			c, _ := gin.CreateTestContext(w)
			url := fmt.Sprintf("%s://%s", scheme, filepath.Join(ociServerIP, "v2", test.url))

			c.Request, err = http.NewRequest(test.method, url, bytes.NewBuffer([]byte{}))
			if err != nil {
				t.Fatalf("Did not expect error: %v", err)
			}

			ServeOciProxy(c)

			// Force writes response headers to combat a response recording issue
			c.Writer.WriteHeaderNow()

			if gotStatusCode := w.Code; !reflect.DeepEqual(gotStatusCode, test.wantCode) {
				t.Errorf("Did not get expected status code, Got: %v, Expected: %v", gotStatusCode, test.wantCode)
			}
		})
	}
}

// TestServeHeadlessUI tests headless handle of the registry viewer endpoint
func TestServeHeadlessUI(t *testing.T) {
	const (
		wantCode = http.StatusBadRequest
		wantBody = "registry viewer is not available in headless mode"
	)

	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	headless = true
	ServeUI(c)

	if gotStatusCode, gotBody := w.Code, w.Body.String(); !reflect.DeepEqual(gotStatusCode, wantCode) {
		t.Errorf("Did not get expected status code, Got: %v, Expected: %v", gotStatusCode, wantCode)
	} else if !reflect.DeepEqual(gotBody, wantBody) {
		t.Errorf("Did not get expected response body, Got: %v, Expected: %v", gotBody, wantBody)
	}
}

// TestBuildProxyErrorResponse tests building the correct error response for proxy routes
func TestBuildProxyErrorResponse(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		proxyName string
		err       error
		wantCode  int
		wantBody  string
	}{
		{
			name:      "No Registry Viewer at Listening Address",
			path:      "/viewer",
			proxyName: "registry viewer",
			err:       errors.New("connection refused"),
			wantCode:  http.StatusBadGateway,
			wantBody:  "registry viewer is not accessible",
		},
		{
			name:      "Other Error",
			path:      "/viewer",
			proxyName: "registry viewer",
			err:       errors.New("something went wrong"),
			wantCode:  http.StatusInternalServerError,
			wantBody:  "internal server error",
		},
	}

	for _, test := range tests {
		w, r := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, test.path, nil)

		buildProxyErrorResponse(w, r, test.err, test.proxyName)

		if gotStatusCode, gotBody := w.Code, w.Body.String(); !reflect.DeepEqual(gotStatusCode, test.wantCode) {
			t.Errorf("Did not get expected status code, Got: %v, Expected: %v", gotStatusCode, test.wantCode)
		} else if !reflect.DeepEqual(gotBody, test.wantBody) {
			t.Errorf("Did not get expected response body, Got: %v, Expected: %v", gotBody, test.wantBody)
		}
	}
}
