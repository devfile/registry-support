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

# This script downloads and caches the devfile samples in the registry
# This is only called if extraDevfilesEntries.yaml exists and has entries for devfile samples
# The downloaded samples are cached under /registry/samples in the devfile registry container
set -eu

# clone_sample_repo clones a given git repository to the sampleDir. Before cloning it checks
# whether or not a revision exists.
# Parameters:
# 1: gitRepository  (git repository url)
# 2: sampleDir      (output directory of the clone command)
# 3: revision       (the prefered revision of the git repo we want to clone)
function clone_sample_repo() {
  local repoUrl="$1"
  local repoOutputDir="$2"
  local repoRevision="$3"
  git clone $repoUrl $repoOutputDir
  if [[ $repoRevision != "null" ]]; then
    cd $repoOutputDir && git checkout $repoRevision && cd -
  fi
}

# cache_sample takes in a given sample name (e.g. nodejs-basic), and git clones its corresponding repository
# Parameters:
# 1: Sample name (e.g. nodejs-basic)
# 2: Output directory
function cache_sample() {
    local sampleName="$1"
    local outputDir="$2"
    tempDir=$(mktemp -d)
    sampleDir=$tempDir/$sampleName

    # Git clone the sample project
    gitRepository="$(yq e '(.samples[] | select(.name == "'${sampleName}'")' $devfileEntriesFile | yq e '(.git.remotes.origin)' -)"
    revision="$(yq e '(.samples[] | select(.name == "'${sampleName}'")' $devfileEntriesFile | yq e '(.git.checkoutFrom.revision)' -)"
    if [[ $gitRepository == "null" ]]; then
        for version in $(yq e '(.samples[] | select(.name == "'${sampleName}'")' $devfileEntriesFile | yq e '(.versions[].version)' -); do
          gitRepository="$(yq e '(.samples[] | select(.name == "'${sampleName}'")' $devfileEntriesFile | yq e '(.versions[] | select(.version == "'${version}'")' -| yq e '.git.remotes.origin' -)"
          revision="$(yq e '(.samples[] | select(.name == "'${sampleName}'")' $devfileEntriesFile | yq e '(.versions[] | select(.version == "'${version}'")' -| yq e '.git.checkoutFrom.revision' -)"
          clone_sample_repo $gitRepository $sampleDir/$version $revision
          mkdir $outputDir/$version
          cache_devfile $sampleDir/$version $outputDir/$version $sampleName
        done
    else
      clone_sample_repo $gitRepository $sampleDir $revision
      cache_devfile $sampleDir $outputDir/ $sampleName
    fi

    # Cache the icon for the sample
    local iconPath="$(yq e '(.samples[] | select(.name == "'${sampleName}'")' $devfileEntriesFile | yq e '(.icon)' -)"
    if [[ $iconPath != "null" ]]; then
      urlRegex='(https?)://[-A-Za-z0-9\+&@#/%?=~_|!:,.;]*[-A-Za-z0-9\+&@#/%=~_|]'
      if [[ $iconPath =~ $urlRegex ]]; then
        (cd $outputDir && curl -O $iconPath)
      else
        if [[ ! -f $tempDir/$sampleName/$iconPath ]]; then
          echo "The specified icon does not exist for sample $sampleName"
          exit 1
        fi
        cp $sampleDir/$iconPath $outputDir/
      fi
    fi

    # Archive the sample project
    (cd $tempDir && zip -r sampleName.zip $sampleName/)
    cp $tempDir/sampleName.zip $outputDir/
    
}

function cache_devfile() {
    local srcDir="$1"
    local outputDir="$2"
    local sampleName="$3"
    # Cache the devfile for the sample
    if [[ -f "$srcDir/devfile.yaml" ]]; then
      cp $srcDir/devfile.yaml $outputDir/
    elif [[ -f "$srcDir/.devfile/devfile.yaml" ]]; then
      cp $srcDir/.devfile/devfile.yaml $outputDir/
    else
      echo "A devfile for sample $sampleName, version $(basename $srcDir) could not be found."
      echo "Please ensure a devfile exists in the root of the repository or under .devfile/"
      exit 1
    fi
}

devfileEntriesFile=$1
samplesDir=$2

for sample in $(yq e '(.samples[].name)' $devfileEntriesFile); do
  mkdir $samplesDir/$sample
  echo $sample
  cache_sample $sample $samplesDir/$sample
done