#
# Copyright (c) 2020 Red Hat, Inc.
# This program and the accompanying materials are made
# available under the terms of the Eclipse Public License 2.0
# which is available at https://www.eclipse.org/legal/epl-2.0/
#
# SPDX-License-Identifier: EPL-2.0
#

#!/bin/bash

# Enable extended globbing in the shell
shopt -s extglob

buildToolsFolder="$(dirname "$0")"
generatorFolder=$buildToolsFolder/../index/generator

display_usage() { 
  echo "A devfile registry repository folder must be passed in as an argument" 
  echo "usage: build.sh <path-to-registry-repository-folder>" 
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

# Copy the registry repository over to a temp folder
registryFolder=$(mktemp -d)
cp -rf $1/. $registryFolder/

cd $generatorFolder

# Build the index generator/validator
echo "Building index-generator tool"
./build.sh
if [ $? -ne 0 ]; then
  echo "Failed to build index-generator tool"
  cleanup_and_exit 1
fi
echo -e "Successfully built the index-generator tool\n"

cd "$OLDPWD"

# Generate the tar archive
for stackDir in $registryFolder/stacks/*/
do
  cd $stackDir
  # Find the files to add to the tar archive
  tarFiles=$(find . \( -not -name 'devfile.yaml' \
    -a -not -name "meta.yaml" \
    -a -not -name "*.vsx" \
    -a -not -name "." \
    -a -not -name "logo.svg" \
    -a -not -name "logo.png" \) -maxdepth 1)

  # There are files that need to be pulled into a tar archive
  if [[ ! -z $tarFiles ]]; then
    tar -cvf archive.tar $tarFiles
    rm -rf $tarFiles
  fi
  cd "$OLDPWD"
done

# Run the index generator tool
echo "Generate the devfile registry index"
$generatorFolder/index-generator $registryFolder/stacks $registryFolder/index.json
if [ $? -ne 0 ]; then
  echo "Failed to build the devfile registry index"
  cleanup_and_exit 1
fi
echo -e "Successfully built the devfile registry index\n"

# Build the Docker image containing the devfile stacks and index.json
echo "Building the devfile registry index container"
docker build -t devfile-index -f $buildToolsFolder/Dockerfile $registryFolder
if [ $? -ne 0 ]; then
  echo "Failed to build the devfile registry index container"
  cleanup_and_exit 1
fi

echo "Successfully built the devfile registry index container"
cleanup_and_exit 0