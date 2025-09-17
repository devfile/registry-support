#!/bin/bash

#
# Copyright Red Hat
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

buildToolsFolder="$(dirname "$0")"
registryFolder=$(mktemp -d)

# run setenv.sh script if invoking directly
if [ "$0" == "$BASH_SOURCE" ]; then
    . $buildToolsFolder/../setenv.sh
fi

display_usage() { 
  echo "A devfile registry repository folder must be passed in as an argument" 
  echo "usage: build_image.sh <path-to-registry-repository-folder> [offline: 0|1]" 
}

# Download all offline resources
download_offline() {
  export STACKS_DIR=$1

  # Download all the offline parent devfiles
  bash $buildToolsFolder/dl_parent_devfiles.sh

  # Download all the offline starter projects
  bash $buildToolsFolder/dl_starter_projects.sh

  # Update all devfiles to use offline starter projects
  bash $buildToolsFolder/update_devfiles_offline.sh

  return $?
}

# cleanup_and_exit removes the temp folder we created and exits with the exit code passed into it
cleanup_and_exit() {
  rm -rf $registryFolder
  exit $1
}

# Check if a registry repository folder was passed in, if not, exit
if [ $# -lt 1 ] || [ $# -gt 2 ]; then
  display_usage
  exit 1
fi

if [ ! -z $2 ] && [ $2 -eq 1 ]; then
  download_offline $1/stacks

  if [ $? -ne 0 ]; then
    exit $?
  fi
fi

bash $buildToolsFolder/build.sh $1 $registryFolder
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