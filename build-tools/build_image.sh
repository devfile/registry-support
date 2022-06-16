#
#   Copyright 2020-2022 Red Hat, Inc.
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.

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