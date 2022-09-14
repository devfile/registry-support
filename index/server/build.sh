#!/bin/bash

# Build the index container for the registry
buildfolder="$(realpath $(dirname ${BASH_SOURCE[0]}))"
registryViewerPath=$1

if [ ! -z $registryViewer ]; then
    # Clone the registry-support repo
    registryViewerPath=$buildfolder/registry-viewer
    if [ -d $buildfolder/registry-viewer ]; then
        rm -rf $registryViewerPath
    fi

    git clone https://github.com/devfile/registry-viewer.git $registryViewerPath
fi

# Build the registry viewer
docker build -t registry-viewer --target builder --build-arg DEVFILE_VIEWER_ROOT=/viewer --build-arg DEVFILE_COMMUNITY_HOST=false $registryViewerPath

# Build the index server
docker build -t devfile-index-base:latest $buildfolder
