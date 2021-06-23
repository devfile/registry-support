#!/bin/bash

# This script downloads and caches the devfile samples in the registry
# This is only called if extraDevfilesEntries.yaml exists and has entries for devfile samples
# The downloaded samples are cached under /registry/samples in the devfile registry container
set -eu

# download_sample takes in a given sample name (e.g. nodejs-basic), and git clones its corresponding repository
# Parameters:
# 1: Sample name (e.g. nodejs-basic)
# 2: Path to extraDevfileEntries.yaml
# 3: Output directory
function cache_sample() {
    local sampleName="$1"
    local outputDir="$2"
    tempDir=$(mktemp -d)

    # Git clone the sample project
    local gitRepository="$(yq e '(.samples[] | select(.name == "'${sampleName}'")' $devfileEntriesFile | yq e '(.git.remotes.origin)' -)"
    git clone "$gitRepository" "$tempDir/$sampleName"

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
        cp $tempDir/$sampleName/$iconPath $outputDir/
      fi
    fi

    # Cache the devfile for the sample
    if [[ -f "$tempDir/$sampleName/devfile.yaml" ]]; then
      cp $tempDir/$sampleName/devfile.yaml $outputDir/
    elif [[ -f "$tempDir/$sampleName/.devfile/devfile.yaml" ]]; then
      cp $tempDir/$sampleName/.devfile/devfile.yaml $outputDir/
    else
      echo "A devfile for sample $sampleName could not be found."
      echo "Please ensure a devfile exists in the root of the repository or under .devfile/"
      exit 1
    fi

    
    # Archive the sample project
    (cd $tempDir && zip -r sampleName.zip $sampleName/)
    cp $tempDir/sampleName.zip $outputDir/
    
}

devfileEntriesFile=$1
samplesDir=$2

for sample in $(yq e '(.samples[].name)' $devfileEntriesFile); do
  mkdir $samplesDir/$sample
  echo $sample
  cache_sample $sample $samplesDir/$sample
done