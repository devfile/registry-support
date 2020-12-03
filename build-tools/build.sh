#
# Copyright (c) 2020 Red Hat, Inc.
# This program and the accompanying materials are made
# available under the terms of the Eclipse Public License 2.0
# which is available at https://www.eclipse.org/legal/epl-2.0/
#
# SPDX-License-Identifier: EPL-2.0
#

#!/bin/sh

buildToolsFolder="$(basename "$(dirname "$0")")"
generatorFolder=$buildToolsFolder/../index/generator

display_usage() { 
  echo "A devfile registry repository folder must be passed in as an argument" 
  echo "usage: ./build.sh <path-to-registry-repository-folder>" 
} 

# Check if a registry repository folder was passed in, if not, exit
registryFolder=$1
if [ $# -ne 1 ]; then
  display_usage
  exit 1
fi

# Build the index generator/validator
echo "Building index-generator tool"
cd $generatorFolder
./build.sh
if [ ! $? -eq 0 ]; then
  echo "Failed to build index-generator tool"
  exit 1
fi
echo -e "Successfully built the index-generator tool\n"

cd "$OLDPWD"

# Run the index generator tool
echo "Generate the devfile registry index"
$generatorFolder/index-generator $registryFolder/stacks $registryFolder/index.json
if [ ! $? -eq 0 ]; then
  echo "Failed to build the devfile registry index"
  exit 1
fi
echo -e "Successfully built the devfile registry index\n"

# Build the Docker image containing the devfile stacks and index.json
echo "Building the devfile registry index container"
docker build -t devfile-index -f $buildToolsFolder/Dockerfile $registryFolder
if [ ! $? -eq 0 ]; then
  echo "Failed to build the devfile registry index container"
  exit 1
fi

echo "Successfully built the devfile registry index container"