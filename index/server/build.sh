#!/bin/bash

# Build the index container for the registry
buildfolder="$(realpath $(dirname ${BASH_SOURCE[0]}))"

# Build the index server
docker build -t devfile-index-base:latest $buildfolder
