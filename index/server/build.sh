#!/bin/bash

# Build the index container for the registry
buildfolder="$(realpath $(dirname ${BASH_SOURCE[0]}))"

# Generate OpenAPI endpoint and type definitions
bash ${buildfolder}/codegen.sh

# Build the index server
docker build -t devfile-index-base:latest $buildfolder
