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

# Due to command differences between podman and docker we need to separate the process
# for creating and adding images to a multi-arch manifest
podman=${USE_PODMAN:-false}

function build {
    #IMAGE="quay.io/devfile/oci-registry:$2"
    IMAGE="quay.io/rh-ee-jdubrick/oci-registry:$2"

    echo "Building: ${IMAGE}"
    $1 build -t $IMAGE --platform "linux/$2" .

    echo "Pushing: ${IMAGE}"
    $1 push $IMAGE


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

  # Push and delete local manifest
else
    echo "Executing with docker"

    engine-handler docker
    # Build and push multi-arch images

    # Create manifest and add images

    # Push and delete local manifest
fi