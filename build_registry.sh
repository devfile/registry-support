#!/bin/bash

# This script builds a devfile registry index container image based on the mock devfile registry data under tests/registry
# This can be useful if developing components within this repository (such as the index server or build tools)
# and want to test all of the components together
shopt -s expand_aliases
set -eux
#set the docker alias if necessary
. ./setenv.sh

# Check if headless arg is passed
if [ -z "${1+x}" ]
then
    headless=true
else
    headless=$1
fi

# Check if static-files-path is passed
if [ -z "${2+x}" ]
then
    static_files_path=
else
    static_files_path=$2
fi

# Build the index server base image
. ./index/server/build.sh

# Build the test devfile registry image
<<<<<<< HEAD
docker build -t devfile-index:latest -f .ci/Dockerfile .
=======
docker build -t devfile-index:latest --build-arg headless=$headless --build-arg static_files_path=$static_files_path -f .ci/Dockerfile .
>>>>>>> f586a44 (build_registry.sh script now includes build script arguments 'headless' & 'static_files_path'.)
