#!/bin/sh

#set the docker alias if necessary
. ../../setenv.sh
# Get the registry-library
cp -rf ../../registry-library ./

# Build the container image
docker build -t devfile-registry-integration ./

rm -rf ./registry-library/