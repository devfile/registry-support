#!/bin/bash

if [[ -z "$@" ]]
then
    echo "No starter projects specified."
    exit 0
fi

# Path of stacks directory in the registry
STACKS_DIR=/registry/stacks
# List of starter projects to use offline
offline_starter_projects=( "$@" )

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
        cd $local_path && git checkout $revision && cd -
    fi

    if [ "${subDir}" != "null" ]
    then
        cd $local_path/$subDir && zip -q ${local_path}.zip * .[^.]* && cd -
    else
        cd $local_path && rm -rf ./.git && zip -q ${local_path}.zip * .[^.]* && cd -
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

# Read stacks list
read -r -a stacks <<< "$(ls ${STACKS_DIR} | tr '\n' ' ')"

echo "Downloading offline starter projects.."
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
echo "Downloading offline starter projects..done!"
