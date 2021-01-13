#!/bin/sh
IMAGE_TAG=$1
docker tag oci-registry:next $IMAGE_TAG
docker push $IMAGE_TAG
