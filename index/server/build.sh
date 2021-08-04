#!/bin/bash

# Build the index container for the registry
buildfolder="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Clone the registry-support repo
if [ -d $buildfolder/registry-viewer ]; then
	rm -rf $buildfolder/registry-viewer
fi
git clone https://github.com/devfile/registry-viewer.git $buildfolder/registry-viewer

docker build -t devfile-index-base:latest $buildfolder
