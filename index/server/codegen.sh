#!/bin/bash

# Base of the index server directory
projectfolder="$(realpath $(dirname ${BASH_SOURCE[0]}))"

# Generate source types from OpenAPI spec
echo "Generating type source.."
cd ${projectfolder} && ${GOPATH}/bin/oapi-codegen -config types.cfg.yaml openapi.yaml && cd - > /dev/null

# Generate endpoint bindings source from OpenAPI spec
echo "Generating endpoint bindings source.."
cd ${projectfolder} && ${GOPATH}/bin/oapi-codegen -config endpoint.cfg.yaml openapi.yaml && cd - > /dev/null

echo "Done."
