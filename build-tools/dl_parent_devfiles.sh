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
STACKS_DIR=${STACKS_DIR:-/registry/stacks}

# Downloads the parent devfile to be used as an offline resource
download_parent_devfile() {
    local stack_root=$1
    local name=$2
    local parent_devfile_uri=$3
    parent_devfile=${name}-parent.devfile.yaml

    if [ ! -f $stack_root/$parent_devfile ]; then
        curl -L $parent_devfile_uri -o $stack_root/$parent_devfile || return 1
    fi
}

# Updates the uri to the downloaded offline parent devfile
replace_parent_devfile() {
    local stack_root=$1
    local name=$2
    local parent_devfile_uri=$3
    stack_devfile=$stack_root/devfile.yaml
    parent_devfile=../${name}-parent.devfile.yaml

    if [ -f $stack_root/$parent_devfile ]; then
        export PARENT_DEVFILE=$parent_devfile
        yq e -i ".parent.uri=env(PARENT_DEVFILE)" $stack_devfile
    fi
}

download_and_replace() {
    for stack in ${stacks[@]}
    do
        if [ $stack == "OWNERS" ]; then
            continue
        fi
        stack_root=$STACKS_DIR/$stack
        stack_devfile=$stack_root/devfile.yaml
        # Read version list for stack
        versions=($([ -f ${STACKS_DIR}/${stack}/stack.yaml ] && yq e '.versions.[].version' ${STACKS_DIR}/${stack}/stack.yaml))
        # Multi version stack
        if [[ ${#versions[@]} -gt 0 ]]
        then
            for version in ${versions[@]}
            do
                stack_root=$STACKS_DIR/$stack/$version
                stack_devfile=$stack_root/devfile.yaml
                name="$(yq e ".metadata.name" $stack_devfile)"
                parent_devfile_uri="$(yq e ".parent.uri" $stack_devfile)"

                if [ "$parent_devfile_uri" != "null" ]
                then
                    echo "Downloading parent devfile in stack ${stack} version ${version}.."
                    download_parent_devfile $stack_root $name $parent_devfile_uri
                    if [ $? -eq 0 ]; then
                        replace_parent_devfile $stack_root $name $parent_devfile_uri
                    fi
                    echo "Downloading parent devfile in stack ${stack} version ${version}..done!"
                fi
            done
        # Not a multi version stack
        else
            name="$(yq e ".metadata.name" $stack_devfile)"
            parent_devfile_uri="$(yq e ".parent.uri" $stack_devfile)"

            if [ "$parent_devfile_uri" != "null" ]
            then
                echo "Downloading parent devfile in stack ${stack}.."
                download_parent_devfile $stack_root $name $parent_devfile_uri
                if [ $? -eq 0 ]; then
                    replace_parent_devfile $stack_root $name $parent_devfile_uri
                fi
                echo "Downloading parent devfile in stack ${stack}..done!"
            fi
        fi
    done
}

# Read stacks list
read -r -a stacks <<< "$(ls ${STACKS_DIR} | tr '\n' ' ')"

echo "Downloading parent devfiles.."
download_and_replace
echo "Downloading parent devfiles..done!"
