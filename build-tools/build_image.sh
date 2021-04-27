#
# Copyright (c) 2020 Red Hat, Inc.
# This program and the accompanying materials are made
# available under the terms of the Eclipse Public License 2.0
# which is available at https://www.eclipse.org/legal/epl-2.0/
#
# SPDX-License-Identifier: EPL-2.0
#

#!/bin/bash

buildToolsFolder="$(dirname "$0")"
registryFolder=$(mktemp -d)

display_usage() { 
  echo "A devfile registry repository folder must be passed in as an argument" 
  echo "usage: build_image.sh <path-to-registry-repository-folder>" 
} 

# cleanup_and_exit removes the temp folder we created and exits with the exit code passed into it
cleanup_and_exit() {
  rm -rf $registryFolder
  exit $1
}

# Check if a registry repository folder was passed in, if not, exit
if [ $# -ne 1 ]; then
  display_usage
  exit 1
fi

$buildToolsFolder/build.sh $1 $registryFolder
if [ $? -ne 0 ]; then
  echo "Failed to build the devfile registry index"
  cleanup_and_exit 1
fi

# Build the Docker image containing the devfile stacks and index.json
echo "Building the devfile registry index container"
docker build -t devfile-index -f $buildToolsFolder/Dockerfile $registryFolder
if [ $? -ne 0 ]; then
  echo "Failed to build the devfile registry index container"
  cleanup_and_exit 1
fi

echo "Successfully built the devfile registry index container"
cleanup_and_exit 0