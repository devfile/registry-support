#!/bin/bash

# Build the index container for the registry
buildfolder="$(realpath $(dirname ${BASH_SOURCE[0]}))"

# Clone the devfile-web repo
if [ -d $buildfolder/devfile-web ]; then
	rm -rf $buildfolder/devfile-web
fi
git clone https://github.com/devfile/devfile-web.git $buildfolder/devfile-web

# Build the index server
docker build -t devfile-index-base:latest $buildfolder
