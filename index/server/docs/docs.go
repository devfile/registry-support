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

// Package classification Devfile registry REST API.
//
// Documentation of devfile registry REST API.
//
// The devfile registry REST API is used for interacting with devfile registry. In this documentation, the host is serving the public devfile registry, you can change it to your own host if you want to use private devfile registry.
//
//     Schemes: http, https
//     Version: 1.0.0
//     Host: preview-devfile-registry-stage.apps.app-sre-stage-0.k3s7.p1.openshiftapps.com
//
//     Consumes:
//     - application/json
//     - application/yaml
//
//     Produces:
//     - application/json
//     - application/yaml
//
//     Security:
//     - basic
//
//    SecurityDefinitions:
//    basic:
//      type: basic
//
// swagger:meta
package docs

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	indexSchema "github.com/devfile/registry-support/index/generator/schema"
)

// getIndex swagger:route GET /index devfile getIndex
//
// Get index.
//
// Return the index content of devfile registry.
//
//     Produces:
//     - application/json
//
//     Responses:
//         200: indexResponse
//         404: description: Page not found.

// getDevfile swagger:route GET /devfiles/{stack} devfile getDevfile
//
// Get devfile by stack name.
//
// Return the specific stack devfile content of devfile registry.
//
//     Produces:
//     - application/yaml
//     - application/json
//
//     Responses:
//         200: devfileResponse
//         404: devfileNotFoundResponse
//         500: devfileErrorResponse

// getStatus swagger:route GET /health server getStatus
//
// Get health status.
//
// Return the devfile registry health status.
//
//     Produces:
//     - application/json
//
//     Responses:
//         200: healthResponse
//         404: description: Page not found.

// Successful operation.
//
// Index content.
//
// swagger:response indexResponse
type IndexResponse struct {
	// in: body
	Payload []indexSchema.Schema
}

// A Stack parameter model.
//
// This is used for operations that want the stack name in the path.
//
// swagger:parameters getDevfile
type Stack struct {
	// The stack name
	//
	// in: path
	// required: true
	Stack string `json:"stack"`
}

// Successful operation.
//
// Stack devfile content.
//
// swagger:response devfileResponse
type DevfileResponse struct {
	// in: body
	Payload v1alpha2.Devfile
}

// Failed to find the devfile.
//
// swagger:response devfileNotFoundResponse
type DevfileNotFoundResponse struct {
	// in: body
	Payload struct {
		Status string `json:"status"`
	}
}

// Failed to get the devfile.
//
// swagger:response devfileErrorResponse
type DevfileErrorResponse struct {
	// in: body
	Payload struct {
		Error  string `json:"error"`
		Status string `json:"status"`
	}
}

// Successful operation.
//
// Health status.
//
// swagger:response healthResponse
type HealthResponse struct {
	// in: body
	Payload struct {
		Message string `json:"message"`
	}
}
