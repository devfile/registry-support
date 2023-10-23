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

# Base of the index server directory
projectfolder="$(realpath $(dirname ${BASH_SOURCE[0]}))"

# Generate source types from OpenAPI spec
echo "Generating type source.."
cd ${projectfolder} && ${GOPATH}/bin/oapi-codegen -config config/types.yaml openapi.yaml && cd - > /dev/null
if [ $? != 0 ]
then
    echo "error with generating type source "
    exit 1
fi

# Generate endpoint bindings source from OpenAPI spec
echo "Generating endpoint bindings source.."
cd ${projectfolder} && ${GOPATH}/bin/oapi-codegen -config config/endpoint.yaml openapi.yaml && cd - > /dev/null
if [ $? != 0 ]
then
    echo "error with generating endpoint bindings source "
    exit 1
fi

echo "Done."
