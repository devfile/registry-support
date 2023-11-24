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

# This script builds a devfile registry index container image based on the mock devfile registry data under tests/registry
# This can be useful if developing components within this repository (such as the index server or build tools)
# and want to test all of the components together
shopt -s expand_aliases
set -eux

# Set base registry support directory
BASE_DIR=$(dirname $0)

#set the docker alias if necessary
. ${BASE_DIR}/setenv.sh

# Build the index server base image
ENABLE_HTTP2="false" . ${BASE_DIR}/index/server/build.sh

# Build the test devfile registry image
docker build -t devfile-index:latest -f ${BASE_DIR}/.ci/Dockerfile ${BASE_DIR}
