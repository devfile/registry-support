#!/bin/bash

# This script builds a devfile registry index container image based on the mock devfile registry data under tests/registry
# This can be useful if developing components within this repository (such as the index server or build tools)
# and want to test all of the components together
shopt -s expand_aliases
set -eux

# Set base registry support directory
BASE_DIR=$(dirname $0)

#set the docker alias if necessary
. ${BASE_DIR}/setenv.sh

# Build the index server base image
. ${BASE_DIR}/index/server/build.sh

# Build the test devfile registry image
docker build -t devfile-index:latest -f ${BASE_DIR}/.ci/Dockerfile ${BASE_DIR}
