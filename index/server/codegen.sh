#!/bin/bash

# Base of the index server directory
projectfolder="$(realpath $(dirname ${BASH_SOURCE[0]}))"

# Generate source types from OpenAPI spec
echo "Generating type source.."
cd ${projectfolder} && ${GOPATH}/bin/oapi-codegen -config config/types.yaml openapi.yaml && cd - > /dev/null
if [ $? != 0 ]
then
    echo "error with generating type source "
    exit 1
fi

# Generate endpoint bindings source from OpenAPI spec
echo "Generating endpoint bindings source.."
cd ${projectfolder} && ${GOPATH}/bin/oapi-codegen -config config/endpoint.yaml openapi.yaml && cd - > /dev/null
if [ $? != 0 ]
then
    echo "error with generating endpoint bindings source "
    exit 1
fi

echo "Done."
