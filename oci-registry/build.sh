#!/bin/sh

# Build the index container for the registry
buildfolder="$(basename "$(dirname "$0")")"
docker build -t oci-registry:next $buildfolder
