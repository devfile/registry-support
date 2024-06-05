#!/bin/bash

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

# this script is commonly called from build_registry.sh and the podman alias is passed down from that script
# this can also affect pathing, to combat this we only run the setenv.sh script if build.sh is being run solo
if [ "$0" == "$BASH_SOURCE" ]; then
    . ../../setenv.sh
fi

DEFAULT_ARCH="linux/amd64"

# Check if different architecture was passed for image build
# Will default to $DEFAULT_ARCH if unset
if [ ! -z "$1" ]
  then
    arch="$1"
else
    arch="$DEFAULT_ARCH"
fi

echo "BUILDING: devfile-index-base for ${arch}"

# Build the index container for the registry
buildfolder="$(realpath $(dirname ${BASH_SOURCE[0]}))"

echo "RUNNING: bash ${buildfolders}/codegen.sh"
# Generate OpenAPI endpoint and type definitions
bash ${buildfolder}/codegen.sh

echo "RUNNING: docker build -t devfile-index-base:latest --platform ${arch} --build-arg ENABLE_HTTP2=${ENABLE_HTTP2} $buildfolder"
# Build the index server
docker build -t devfile-index-base:latest --platform "${arch}" --build-arg ENABLE_HTTP2=${ENABLE_HTTP2} $buildfolder
