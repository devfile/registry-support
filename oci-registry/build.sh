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

# Build the index container for the registry
buildfolder="$(basename "$(dirname "$0")")"

DEFAULT_ARCH="linux/amd64"

# Check if different architecture was passed for image build
# Will default to $DEFAULT_ARCH if unset
if [ ! -z "$1" ]
  then
    arch="$1"
else
    arch="$DEFAULT_ARCH"
fi

# set podman alias if necessary
. ${buildfolder}/../setenv.sh

# Copy license to include in image build
cp ${buildfolder}/../LICENSE ${buildfolder}/LICENSE

docker build -t oci-registry:next --platform "${arch}" "$buildfolder"

# Remove license from build directory
rm ${buildfolder}/LICENSE
