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

# Path of stacks directory in the registry
STACKS_DIR=${STACKS_DIR:-/build/stacks}

extract() {
    local stack_root=$1
    if [[ -f "$stack_root/archive.tar" ]]
    then
        tar -xf "$stack_root/archive.tar" -C "$stack_root"
        echo "Successfully extracted archive.tar"
    else
        echo "Skipping... no archive.tar found"
    fi
}

registry=$1

read -r -a stacks <<< "$(ls ${STACKS_DIR} | tr '\n' ' ')"

for stack in ${stacks[@]}
do
    stack_root=$STACKS_DIR/$stack
    stack_archive=$stack_root/archive.tar

    # Read version list for stack
    versions=($([ -f ${STACKS_DIR}/${stack}/stack.yaml ] && yq e '.versions.[].version' ${STACKS_DIR}/${stack}/stack.yaml))

    # Multi version stack
    if [[ ${#versions[@]} -gt 0 ]]
    then
        for version in ${versions[@]}
            do
                echo "Extracting archive.tar in stack ${stack} version ${version}.."
                extract "$stack_root/$version"
            done
    # Not a multi version
    else
        echo "Extracting archive.tar in stack ${stack}.."
        extract $stack_root
    fi
done
