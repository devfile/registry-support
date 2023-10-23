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

# This script will modify the devfiles in each stack to use offline resources.
# Stack devfiles will not be modified if there are no starterProjects or if the
# staterProjects have already been modified.
#   - starterProjects will be commented out and replaced with a zip block with
#     the location to the offline project in the stack root

# Path of stacks directory in the registry
STACKS_DIR=/registry/stacks
# Automated comment used to check whether the devfile has already been modified
MODIFIED_MESSAGE="# AUTOMATED MODIFICATION -"

cleanup() {
    rm -f "$devfile.tmp"
    rm -f "$offline_starter_projects"
}

trap cleanup EXIT

comment_out_starter_project() {
    stack_root=$1
    name=$2
    devfile=$3

    has_starterProjects=$(yq e '.starterProjects' $devfile 2> /dev/null)
    if [[ $has_starterProjects != null ]]
    then
        # the first line of the diff needs to be removed; example of the diff:
        # 44a45,52
        # > starterProjects:
        # >   - name: go-starter
        # ...
        #
        # 's/^../#/' replaces the first 2 characters of each line with #
        diff="$(diff -b <(yq e 'del(.starterProjects)' $devfile) $devfile | tail -n +2)"
        starter_projects="$(printf '%s\n' "$diff" | sed 's/^../#/')"

        # comment out the starter projects
        yq e 'del(.starterProjects)' "$devfile" > "$devfile.tmp"
        echo "# Commented out original starter projects" >> "$devfile.tmp"
        echo "$starter_projects" >> "$devfile.tmp"
        mv "$devfile.tmp" "$devfile"
    fi
}

# Updates a devfile to use an offline starter project
new_starter_projects() {
    stack_root=$1
    name=$2
    devfile=$3
    offline_starter_projects=$4

    project="
    - name: ${name}
      zip:
        location: ${name}-offline.zip"

    echo -n "$project" >> $offline_starter_projects
}

# Read stacks list
read -r -a stacks <<< "$(ls ${STACKS_DIR} | tr '\n' ' ')"

echo "Updating devfiles.."
for stack in ${stacks[@]}
do
    stack_root=$STACKS_DIR/$stack
    stack_devfile=$stack_root/devfile.yaml
    # Exit early on failure to avoid bad overwriting of the devfile
    offline_starter_projects=$(mktemp) || exit 1
    # Overwrite the temp file for each stack
    echo -n "" > $offline_starter_projects
    # Read version list for stack
    read -r -a versions <<< "$(ls ${STACKS_DIR}/${stack} | grep -e '[0-9].[0-9].[0-9]' | tr '\n' ' ')"

    # If multi version stack
    if [[ ${#versions[@]} -gt 0 ]]
    then
        for version in ${versions[@]}
        do
            stack_root=$STACKS_DIR/$stack/$version
            stack_devfile=$stack_root/devfile.yaml
            starter_projects="$(yq e ".starterProjects[].name" $stack_devfile)"
            echo -n "starterProjects:" > $offline_starter_projects

            if [[ $starter_projects == "" ]]
            then
                echo "Skipping stack ${stack} version ${version}: no starter projects found."
                continue
            fi

            for starter_project in $starter_projects
            do
                if ! grep -q "$MODIFIED_MESSAGE" "$stack_devfile";
                then
                    echo "Updating the ${starter_project} starter project in stack ${stack} version ${version}.."
                    new_starter_projects $stack_root $starter_project $stack_devfile $offline_starter_projects
                    comment_out_starter_project $stack_root $starter_project $stack_devfile
                    echo "Updating the ${starter_project} starter project in stack ${stack} version ${version}..done!"
                else
                    echo "The ${starter_project} starter project in stack ${stack} version ${version} has already been modified."
                fi
            done

            # Only write to the devfile if starter projects have been commented out
            has_starterProjects=$(yq e '.starterProjects' $stack_devfile 2> /dev/null)
            if [[ $has_starterProjects == null ]] && ! grep -q "$MODIFIED_MESSAGE" "$stack_devfile"
            then
                    echo "${MODIFIED_MESSAGE} Updated starterProjects to use offline versions" >> $stack_devfile
                    cat $offline_starter_projects >> $stack_devfile
            fi
        done
    # If not multi version stack
    else
        starter_projects="$(yq e ".starterProjects[].name" $stack_devfile)"
        echo -n "starterProjects:" > $offline_starter_projects

        if [[ $starter_projects == "" ]]
        then
            echo "Skipping stack ${stack}: no starter projects found."
            continue
        fi

        for starter_project in $starter_projects
        do
            if ! grep -q "$MODIFIED_MESSAGE" "$stack_devfile";
            then
                echo "Updating the ${starter_project} starter project in stack ${stack}.."
                new_starter_projects $stack_root $starter_project $stack_devfile $offline_starter_projects
                comment_out_starter_project $stack_root $starter_project $stack_devfile
                echo "Updating the ${starter_project} starter project in stack ${stack}..done!"
            else
                echo "The ${starter_project} starter project in stack ${stack} has already been modified."
            fi
        done

        # Only write to the devfile if starter projects have been commented out
        has_starterProjects=$(yq e '.starterProjects' $stack_devfile 2> /dev/null)
        if [[ $has_starterProjects == null ]] && ! grep -q "$MODIFIED_MESSAGE" "$stack_devfile"
        then
            echo "${MODIFIED_MESSAGE} Updated starterProjects to use offline versions" >> $stack_devfile
            cat $offline_starter_projects >> $stack_devfile
        fi
    fi
done
echo "Updating devfiles....done!"
