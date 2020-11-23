#!/bin/sh

# Build the index container for the registry
buildfolder="$(basename "$(dirname "$0")")"
docker build -t devfile-index-base:latest $buildfolder
