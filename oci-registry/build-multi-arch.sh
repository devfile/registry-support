#!/bin/sh

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

buildfolder="$(basename "$(dirname "$0")")"
# Due to command differences between podman and docker we need to separate the process
# for creating and adding images to a multi-arch manifest
podman=${USE_PODMAN:-false}
# Stores all created image tags 
images=()
# Base Repository
BASE_REPO="quay.io/devfile/oci-registry"
BASE_TAG="next"
DEFAULT_MANIFEST="$BASE_REPO:$BASE_TAG"

function build {
    IMAGE="$BASE_REPO:$2"

    echo "Building: ${IMAGE}"
    $1 build -t $IMAGE --platform "linux/$2" "$buildfolder"

    echo "Tagging: ${IMAGE}"
    $1 tag "$IMAGE" "$IMAGE"

    echo "Pushing: ${IMAGE}"
    $1 push "$IMAGE"

    # Add image to list of all images to be added to a manifest
    images+=("${IMAGE}")
}

function engine-handler {
    for arch in amd64 arm64 ; do
        build "$1" "$arch"
    done
}


if [ ${podman} == true ]; then
  echo "Executing with podman"

  # Build and push multi-arch images
  engine-handler podman

  # Create manifest and add images
  podman manifest create oci-registry-manifest
  for img in "${images[@]}" ; do
    podman manifest add oci-registry-manifest "$img"
  done

  # Push and delete local manifest
  podman manifest push oci-registry-manifest "$DEFAULT_MANIFEST"
  podman manifest rm oci-registry-manifest

else
  echo "Executing with docker"

  # Build and push multi-arch images
  engine-handler docker
  
  # Create manifest and add images
  docker manifest create "$DEFAULT_MANIFEST" "${images[@]}"

  # Push and delete local manifest
  docker manifest push "$DEFAULT_MANIFEST"
  docker manifest rm "$$DEFAULT_MANIFEST"

fi