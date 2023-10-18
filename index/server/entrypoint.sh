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

set -eux

# Check if devfile stacks and index.json exist
if [ ! -d "$DEVFILE_STACKS" ]; then
    echo "The container does not contain any devfile stacks in $DEVFILE_STACKS. Exiting..."
    exit 1
fi
if [ ! -e "$DEVFILE_INDEX" ]; then
    echo "The container does not contain an index.json at $DEVFILE_INDEX. Exiting..."
    exit 1
fi

# Start the index server
/registry/index-server
