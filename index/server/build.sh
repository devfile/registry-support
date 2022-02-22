#!/bin/bash

# Build the index container for the registry
buildfolder="$(realpath $(dirname ${BASH_SOURCE[0]}))"

# Clone the registry-support repo
if [ -d $buildfolder/registry-viewer ]; then
	rm -rf $buildfolder/registry-viewer
fi
git clone https://github.com/devfile/registry-viewer.git $buildfolder/registry-viewer

# Build the registry viewer
docker build -t registry-viewer --target builder --build-arg DEVFILE_VIEWER_ROOT=/viewer --build-arg DEVFILE_COMMUNITY_HOST=false $buildfolder/registry-viewer

# Build the index server
docker build -t devfile-index-base:latest $buildfolder
