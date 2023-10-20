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
STACKS_DIR=/registry/stacks
# List of starter projects to use offline
offline_starter_projects=( "$@" )
# When no starter projects are specifed,
# all starter projects will be downloaded
download_all_starter_projects=false

if [[ -z "$@" ]]
then
    download_all_starter_projects=true
fi

# Downloads a starter project from a remote git repository and packages it as a zip archive
# to be used as an offline resource.
download_git_starter_project() {
    stack_root=$1
    name=$2
    remote_name=$(yq e ".starterProjects[] | select(.name == \"${name}\").git.checkoutFrom.remote" $stack_root/devfile.yaml)
    revision=$(yq e ".starterProjects[] | select(.name == \"${name}\").git.checkoutFrom.revision" $stack_root/devfile.yaml)
    subDir=$(yq e ".starterProjects[] | select(.name == \"${name}\").subDir" $stack_root/devfile.yaml)
    local_path=${stack_root}/${name}-offline

    if [ "${remote_name}" == "null" ]
    then
        remote_url=$(yq e ".starterProjects[] | select(.name == \"${name}\").git.remotes.origin" $stack_root/devfile.yaml)
    else
        remote_url=$(yq e ".starterProjects[] | select(.name == \"${name}\").git.remotes.${remote_name}" $stack_root/devfile.yaml)
    fi

    mkdir -p $local_path

    git clone $remote_url $local_path

    if [ "${revision}" != "null" ]
    then
        cd $local_path
        git checkout $revision
        cd -
    fi

    if [ "${subDir}" != "null" ]
    then
        cd $local_path/$subDir
        zip -q ../${name}-offline.zip * .[^.]*
        cd -
    else
        cd $local_path
        rm -rf ./.git
        zip -q ../${name}-offline.zip * .[^.]*
        cd -
    fi

    rm -rf $local_path
}

# Downloads a starter project from a remote zip archive source
# to be used as an offline resource.
download_zip_starter_project() {
    stack_root=$1
    name=$2
    remote_url=$(yq e ".starterProjects[] | select(.name == \"${name}\").zip.location" $stack_root/devfile.yaml)
    local_path=${stack_root}/${name}-offline

    curl -L $remote_url -o ${local_path}.zip
}

download_specific() {
    for starter_project in ${offline_starter_projects[@]}
    do
        for stack in ${stacks[@]}
        do
            stack_root=$STACKS_DIR/$stack
            stack_devfile=$stack_root/devfile.yaml
            # Read version list for stack
            read -r -a versions <<< "$(ls ${STACKS_DIR}/${stack} | grep -e '[0-9].[0-9].[0-9]' | tr '\n' ' ')"
            # If multi version stack
            if [[ ${#versions[@]} -gt 0 ]]
            then
                for version in ${versions[@]}
                do
                    stack_root=$STACKS_DIR/$stack/$version
                    stack_devfile=$stack_root/devfile.yaml
                    # If the specified starter project is found
                    if [ ! -z "$(yq e ".starterProjects[] | select(.name == \"${starter_project}\")" $stack_devfile)" ]
                    then
                        # Starter project has a git remote
                        if [ "$(yq e ".starterProjects[] | select(.name == \"${starter_project}\").git" $stack_devfile)" != "null" ]
                        then
                            echo "Downloading ${starter_project} starter project in stack ${stack} version ${version}.."
                            download_git_starter_project $stack_root $starter_project
                            echo "Downloading ${starter_project} starter project in stack ${stack} version ${version}..done!"
                        # Starter project has a zip remote
                        elif [ "$(yq e ".starterProjects[] | select(.name == \"${starter_project}\").zip" $stack_devfile)" != "null" ]
                        then
                            echo "Downloading ${starter_project} starter project in stack ${stack} version ${version}.."
                            download_zip_starter_project $stack_root $starter_project
                            echo "Downloading ${starter_project} starter project in stack ${stack} version ${version}..done!"
                        fi
                    fi
                done
            # If not multi version stack & the specified starter project is found
            elif [ ! -z "$(yq e ".starterProjects[] | select(.name == \"${starter_project}\")" $stack_devfile)" ]
            then
                # Starter project has a git remote
                if [ "$(yq e ".starterProjects[] | select(.name == \"${starter_project}\").git" $stack_devfile)" != "null" ]
                then
                    echo "Downloading ${starter_project} starter project in stack ${stack}.."
                    download_git_starter_project $stack_root $starter_project
                    echo "Downloading ${starter_project} starter project in stack ${stack}..done!"
                # Starter project has a zip remote
                elif [ "$(yq e ".starterProjects[] | select(.name == \"${starter_project}\").zip" $stack_devfile)" != "null" ]
                then
                    echo "Downloading ${starter_project} starter project in stack ${stack}.."
                    download_zip_starter_project $stack_root $starter_project
                    echo "Downloading ${starter_project} starter project in stack ${stack}..done!"
                fi
            fi
        done
    done
}

download_all() {
    for stack in ${stacks[@]}
    do
        stack_root=$STACKS_DIR/$stack
        stack_devfile=$stack_root/devfile.yaml
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

                for starter_project in $starter_projects
                do
                    # Starter project has a git remote
                    if [ "$(yq e ".starterProjects[] | select(.name == \"${starter_project}\").git" $stack_devfile)" != "null" ]
                    then
                        echo "Downloading ${starter_project} starter project in stack ${stack} version ${version}.."
                        download_git_starter_project $stack_root $starter_project
                        echo "Downloading ${starter_project} starter project in stack ${stack} version ${version}..done!"
                    # Starter project has a zip remote
                    elif [ "$(yq e ".starterProjects[] | select(.name == \"${starter_project}\").zip" $stack_devfile)" != "null" ]
                    then
                        echo "Downloading ${starter_project} starter project in stack ${stack} version ${version}.."
                        download_zip_starter_project $stack_root $starter_project
                        echo "Downloading ${starter_project} starter project in stack ${stack} version ${version}..done!"
                    fi
                done
            done
        # If not multi version stack
        else
            starter_projects="$(yq e ".starterProjects[].name" $stack_devfile)"
            for starter_project in $starter_projects
            do
                # Starter project has a git remote
                if [ "$(yq e ".starterProjects[] | select(.name == \"${starter_project}\").git" $stack_devfile)" != "null" ]
                then
                    echo "Downloading ${starter_project} starter project in stack ${stack}.."
                    download_git_starter_project $stack_root $starter_project
                    echo "Downloading ${starter_project} starter project in stack ${stack}..done!"
                # Starter project has a zip remote
                elif [ "$(yq e ".starterProjects[] | select(.name == \"${starter_project}\").zip" $stack_devfile)" != "null" ]
                then
                    echo "Downloading ${starter_project} starter project in stack ${stack}.."
                    download_zip_starter_project $stack_root $starter_project
                    echo "Downloading ${starter_project} starter project in stack ${stack}..done!"
                fi
            done
        fi
        echo
    done
}

# Read stacks list
read -r -a stacks <<< "$(ls ${STACKS_DIR} | tr '\n' ' ')"

echo "Downloading offline starter projects.."
if [ "$download_all_starter_projects" = true ]
then
    download_all
else
    download_specific
fi
echo "Downloading offline starter projects..done!"
