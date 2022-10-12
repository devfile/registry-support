#!/bin/sh

#set the docker alias if necessary
. ../../setenv.sh

IMAGE_TAG=$1
docker tag devfile-registry-integration $IMAGE_TAG
docker push $IMAGE_TAG