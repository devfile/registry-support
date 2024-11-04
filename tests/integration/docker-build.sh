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

#set the docker alias if necessary
. ../../setenv.sh
# Get the registry-library
cp -rf ../../registry-library ./

# Copy license to include in image build
cp ../../LICENSE LICENSE

# Build the container image
docker build -t devfile-registry-integration ./

# Remove license from build directory
rm LICENSE

rm -rf ./registry-library/