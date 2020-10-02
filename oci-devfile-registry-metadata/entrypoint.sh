#!/bin/sh

## Simple proof of concept bootstrap script to load devfiles into an oci registry
DEVFILES=/registry/devfiles

# Generate the index.json from the devfiles
cd /registry
./index -devfiles-dir ./devfiles -index /usr/local/apache2/htdocs/index.json

# Push the devfiles to the registry
# ToDo: Cleanup
for devfileDir in "$DEVFILES"/*
do
  devfile="$devfileDir/devfile.yaml"
  stackName=`basename $devfileDir`
  # Push the devfile to the registry
  
  # ToDo: Fix
  # 1) Discover the service name for the registry
  # 2) Getting the stack name (need to be reading the meta.yaml)
  # 3) Getting the stack version
  # 4) Not pushing over plain http
  # Might be easier to do in Golang... Maybe put this into the index util?
  echo "Pushing $stackName to $REGISTRY_HOST"
  oras push devfile-registry:5000/devfile-catalog/$stackName:latest --manifest-config /dev/null:application/vnd.devfileio.devfile.config.v2+json ./devfile.yaml:application/vnd.devfileio.devfile.layer.v1 --plain-http
done

# Launch the server hosting the index.json
echo $REGISTRY_HOST
exec "${@}"