#!/bin/sh
IMAGE_TAG=$1
docker tag devfile-registry-metadata:latest $IMAGE_TAG
docker push $1