#!/bin/sh
IMAGE_TAG=$1
docker tag devfile-index-base:latest $IMAGE_TAG
docker push $1