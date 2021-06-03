#!/bin/sh

# This script builds a devfile registry index container image based on the mock devfile registry data under tests/registry
# This can be useful if developing components within this repository (such as the index server or build tools)
# and want to test all of the components together

set -eux

# Build the index server base image
docker build -t devfile-index-base:latest ./index/server/

# Build the test devfile registry image
docker build -t devfile-index:latest -f .ci/Dockerfile .