package library

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	indexSchema "github.com/devfile/registry-support/index/generator/schema"
)

func TestGetRegistryIndex(t *testing.T) {
	const serverIP = "127.0.0.1:8080"
	archFilteredIndex := []indexSchema.Schema{
		{
			Name:          "archindex1",
			Architectures: []string{"amd64, arm64"},
		},
		{
			Name: "archindex2",
		},
	}

	schemaVersionFilteredIndex := []indexSchema.Schema{
		{
			Name: "indexSchema2.1",
			Versions: []indexSchema.Version{
				{
					Version:       "1.0.0",
					SchemaVersion: "2.1.0",
				},
			},
		},
		{
			Name: "indexSchema2.2",
			Versions: []indexSchema.Version{
				{
					Version:       "1.1.0",
					SchemaVersion: "2.2.0",
				},
			},
		},
	}

	sampleFilteredIndex := []indexSchema.Schema{
		{
			Name: "sampleindex1",
		},
		{
			Name: "sampleindex2",
		},
	}

	sampleFilteredV2Index := []indexSchema.Schema{
		{
			Name: "samplev2index1",
		},
		{
			Name: "samplev2index2",
		},
	}

	stackFilteredIndex := []indexSchema.Schema{
		{
			Name: "stackindex1",
		},
		{
			Name: "stackindex2",
		},
	}

	stackFilteredV2Index := []indexSchema.Schema{
		{
			Name: "stackv2index1",
		},
		{
			Name: "stackv2index2",
		},
	}

	notFilteredIndex := []indexSchema.Schema{
		{
			Name: "index1",
		},
		{
			Name: "index2",
		},
	}

	notFilteredV2Index := []indexSchema.Schema{
		{
			Name: "v2index1",
		},
		{
			Name: "v2index2",
		},
	}

	// Mocking the registry REST endpoints on a very basic level
	testServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var data []indexSchema.Schema
		var err error

		if strings.Contains(r.URL.String(), "arch=amd64&arch=arm64") {
			data = archFilteredIndex
		} else if strings.Contains(r.URL.String(), "minSchemaVersion=2.1&maxSchemaVersion=2.2") {
			data = schemaVersionFilteredIndex
		} else if r.URL.Path == "/index/sample" {
			data = sampleFilteredIndex
		} else if r.URL.Path == "/v2index/sample" {
			data = sampleFilteredV2Index
		} else if r.URL.Path == "/index/stack" || r.URL.Path == "/index" {
			data = stackFilteredIndex
		} else if r.URL.Path == "/v2index/stack" || r.URL.Path == "/v2index" {
			data = stackFilteredV2Index
		} else if r.URL.Path == "/index/all" {
			data = notFilteredIndex
		} else if r.URL.Path == "/v2index/all" {
			data = notFilteredV2Index
		}

		bytes, err := json.MarshalIndent(&data, "", "  ")
		if err != nil {
			t.Errorf("Unexpected error while doing json marshal: %v", err)
			return
		}

		_, err = w.Write(bytes)
		if err != nil {
			t.Errorf("Unexpected error while writing data: %v", err)
		}
	}))
	// create a listener with the desired port.
	l, err := net.Listen("tcp", serverIP)
	if err != nil {
		t.Errorf("Unexpected error while creating listener: %v", err)
		return
	}

	// NewUnstartedServer creates a listener. Close that listener and replace
	// with the one we created.
	testServer.Listener.Close()
	testServer.Listener = l

	testServer.Start()
	defer testServer.Close()

	tests := []struct {
		name         string
		url          string
		options      RegistryOptions
		devfileTypes []indexSchema.DevfileType
		wantSchemas  []indexSchema.Schema
		wantErr      bool
	}{
		{
			name: "Get Devfile Schema Filtered Index",
			url:  "http://" + serverIP,
			options: RegistryOptions{
				NewIndexSchema: true,
				Filter: RegistryFilter{
					MinSchemaVersion: "2.1",
					MaxSchemaVersion: "2.2",
				},
			},
			wantSchemas: schemaVersionFilteredIndex,
		},
		{
			name: "Get Arch Filtered Index",
			url:  "http://" + serverIP,
			options: RegistryOptions{
				Filter: RegistryFilter{
					Architectures: []string{"amd64", "arm64"},
				},
			},
			devfileTypes: []indexSchema.DevfileType{indexSchema.SampleDevfileType},
			wantSchemas:  archFilteredIndex,
		},
		{
			name:         "Get Sample Filtered Index",
			url:          "http://" + serverIP,
			devfileTypes: []indexSchema.DevfileType{indexSchema.SampleDevfileType},
			wantSchemas:  sampleFilteredIndex,
		},
		{
			name: "Get Sample Filtered V2 Index",
			url:  "http://" + serverIP,
			options: RegistryOptions{
				NewIndexSchema: true,
			},
			devfileTypes: []indexSchema.DevfileType{indexSchema.SampleDevfileType},
			wantSchemas:  sampleFilteredV2Index,
		},
		{
			name:         "Get Stack Filtered Index",
			url:          "http://" + serverIP,
			devfileTypes: []indexSchema.DevfileType{indexSchema.StackDevfileType},
			wantSchemas:  stackFilteredIndex,
		},
		{
			name: "Get Stack Filtered V2 Index",
			url:  "http://" + serverIP,
			options: RegistryOptions{
				NewIndexSchema: true,
			},
			devfileTypes: []indexSchema.DevfileType{indexSchema.StackDevfileType},
			wantSchemas:  stackFilteredV2Index,
		},
		{
			name:         "Get all of the Indexes",
			url:          "http://" + serverIP,
			devfileTypes: []indexSchema.DevfileType{indexSchema.StackDevfileType, indexSchema.SampleDevfileType},
			wantSchemas:  notFilteredIndex,
		},
		{
			name: "Get all of the V2 Indexes",
			url:  "http://" + serverIP,
			options: RegistryOptions{
				NewIndexSchema: true,
			},
			devfileTypes: []indexSchema.DevfileType{indexSchema.StackDevfileType, indexSchema.SampleDevfileType},
			wantSchemas:  notFilteredV2Index,
		},
		{
			name:    "Not a URL",
			url:     serverIP,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotSchemas, err := GetRegistryIndex(test.url, test.options, test.devfileTypes...)
			if !test.wantErr && err != nil {
				t.Errorf("Unexpected err: %+v", err)
			} else if test.wantErr && err == nil {
				t.Errorf("Expected error but got nil")
			} else if !reflect.DeepEqual(gotSchemas, test.wantSchemas) {
				t.Errorf("Expected: %+v, \nGot: %+v", test.wantSchemas, gotSchemas)
			}
		})
	}
}
