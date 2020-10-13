#!/bin/sh

## Simple proof of concept bootstrap script to load devfiles into an oci registry
DEVFILES=/registry/devfiles

# Generate the index.json from the devfiles
cd /registry
./index-generator ./devfiles /usr/local/apache2/htdocs/devfiles/index.json

# Push the devfiles to the registry
# ToDo: Cleanup
cd /registry/devfiles
for devfileDir in "$DEVFILES"/*
do
  devfile="$devfileDir/devfile.yaml"
  stackName=`basename $devfileDir`
  # Push the devfile to the registry
  
  # ToDo:
  # 1) Discover the service name for the registry
  # 2) Getting the stack name (need to be reading the meta.yaml)
  # 3) Getting the proper stack version
  # 4) Not pushing over plain http
  # 5) Do in Golang
  echo "Pushing $stackName to $REGISTRY_HOST"
  cd $stackName
  oras push devfile-registry:5000/devfile-catalog/$stackName:latest --manifest-config /dev/null:application/vnd.devfileio.devfile.config.v2+json ./devfile.yaml:application/vnd.devfileio.devfile.layer.v1 --plain-http
  cd /registry/devfiles
done

# Launch the server hosting the index.json
echo $REGISTRY_HOST
exec "${@}"