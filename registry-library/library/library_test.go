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

	sampleFilteredIndex := []indexSchema.Schema{
		{
			Name: "sampleindex1",
		},
		{
			Name: "sampleindex2",
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

	notFilteredIndex := []indexSchema.Schema{
		{
			Name: "index1",
		},
		{
			Name: "index2",
		},
	}

	// Mocking the registry REST endpoints on a very basic level
	testServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var data []indexSchema.Schema
		var err error

		if strings.Contains(r.URL.String(), "arch=amd64&arch=arm64") {
			data = archFilteredIndex
		} else if r.URL.Path == "/index/sample" {
			data = sampleFilteredIndex
		} else if r.URL.Path == "/index/stack" || r.URL.Path == "/index" {
			data = stackFilteredIndex
		} else if r.URL.Path == "/index/all" {
			data = notFilteredIndex
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
			name:         "Get Stack Filtered Index",
			url:          "http://" + serverIP,
			devfileTypes: []indexSchema.DevfileType{indexSchema.StackDevfileType},
			wantSchemas:  stackFilteredIndex,
		},
		{
			name:         "Get all of the Indexes",
			url:          "http://" + serverIP,
			devfileTypes: []indexSchema.DevfileType{indexSchema.StackDevfileType, indexSchema.SampleDevfileType},
			wantSchemas:  notFilteredIndex,
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
