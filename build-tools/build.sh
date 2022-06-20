#!/bin/bash
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


buildToolsFolder="$(dirname "$0")"
buildToolsDir="$PWD"
generatorFolder=$buildToolsFolder/../index/generator

display_usage() { 
  echo "usage: build.sh <path-to-registry-repository-folder> <output-dir>" 
} 

# cleanup_and_exit removes the build folder and exits with the exit code passed into it
cleanup_and_exit() {
  rm -rf $outputFolder
  exit $1
}

tar_files_and_cleanup() {
  # Find the files to add to the tar archive
  tarFiles=$(find . \( -not -name 'devfile.yaml' \
    -a -not -name "meta.yaml" \
    -a -not -name "*.vsx" \
    -a -not -name "." \
    -a -not -name "logo.svg" \
    -a -not -name "logo.png" \
    -a -not -name "*.zip" \) -maxdepth 1)

  # There are files that need to be pulled into a tar archive
  if [[ ! -z $tarFiles ]]; then
    tar -czvf archive.tar $tarFiles > /dev/null
    rm -rf $tarFiles
  fi
}

# build_registry <registry-folder> <output>
# Runs the steps to build the registry. Mainly:
# 1. Copying over registry repository to build folder
# 2. Building the index-generator tool -> ToDo: Download specific release of index-generator rather than building it
# 3. Create the tar archives for any miscellaneous files in each stack
# 4. Generate the index.json
build_registry() {
  # Copy the registry repository over to the destination folder
  cp -rf $registryRepository/. $outputFolder/

  cd $generatorFolder

  # Build the index generator/validator
  echo "Building index-generator tool"
  ./build.sh
  if [ $? -ne 0 ]; then
    echo "Failed to build index-generator tool"
    return 1
  fi
  echo "Successfully built the index-generator tool"

  cd "$OLDPWD"

  # Generate the tar archive
  for stackDir in $outputFolder/stacks/*
  do
    cd $stackDir
    if [[ -f "stack.yaml" ]]; then
      for versionDir in $stackDir/*
      do
        cd $versionDir
        tar_files_and_cleanup
      done
    else
      tar_files_and_cleanup
    fi
    cd "$OLDPWD"
  done
  cd "$buildToolsDir"

  # Cache any devfile samples if needed
  if [ -f $registryRepository/extraDevfileEntries.yaml ]; then
    mkdir $outputFolder/samples
    $buildToolsFolder/cache_samples.sh $registryRepository/extraDevfileEntries.yaml $outputFolder/samples
    if [ $? -ne 0 ]; then
      echo "Error caching the devfile samples"
      exit 1;
    fi
  fi

  # Run the index generator tool
  echo "Generating the devfile registry index"
  $generatorFolder/index-generator $outputFolder $outputFolder/index.json
  if [ $? -ne 0 ]; then
    echo "Failed to build the devfile registry index"
    return 1
  fi
  echo "Successfully built the devfile registry index"
}

# check_params validates that the arguments passed into the script are valid
# The first parameter must point to a valid devfile registry folder, containing a stacks folder
# The second parameter must point to an empty output folder, or a folder that does not yet exist.
check_params() {
  # If the output folder does not have a stacks folder, we cannot do the build, so exit out
  if [ ! -d "$registryRepository/stacks" ]; then
    echo "A valid devfile registry was not passed in. Please specify a devfile registry folder scontaining a stacks folder."
    display_usage
    exit 1
  fi

  # If the output registry folder does not exist, create it.
  if [ ! -d $outputFolder ]; then
    mkdir -p $outputFolder
  fi

  # If the speicifed output folder is not empty, exit.
  if [ ! -z "$(ls -A $outputFolder)" ]; then
    echo "The specified destination folder is not empty. Please specify an empty folder."
    display_usage
    exit 1
  fi
}

# Check if a registry repository folder and a output folder were passed in, if not, exit
if [ $# -ne 2 ]; then
  display_usage
  exit 1
fi
registryRepository=$1
outputFolder=$2

# Validate the script parameters
check_params

# Build the registry
build_registry
if [ $? -ne 0 ]; then
  echo "Error building the devfile registry"
  cleanup_and_exit 1
fi
