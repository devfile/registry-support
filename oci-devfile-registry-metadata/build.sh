#!/bin/sh

# Build the metadata container for the registry
buildfolder="$(basename "$(dirname "$0")")"
cp -rf $buildfolder/../index/generator $buildfolder
docker build -t devfile-registry-metadata:latest $buildfolder