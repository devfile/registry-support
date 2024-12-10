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

fetch_push_prior_release () {
    git fetch $upstream_name --tags
    LATEST_TAG=$(git tag --sort=-v:refname | head -n 1)
    MODIFIED_TAG=$(echo "$LATEST_TAG" | awk -F. '{print $1 "." $2 ".x"}') # convert to [major].[minor].x format from [major].[minor].[patch]
    git branch release/$MODIFIED_TAG $LATEST_TAG
    git push $upstream_name release/$MODIFIED_TAG
    git branch -D release/$MODIFIED_TAG
}

# append rc for release-candidate if necessary
tag_and_push () {
    final_version="v$VERSION"
    if [ "$1" == "rc" ]; then
        final_version+="-rc"
    fi
    git tag $final_version
    git push $upstream_name $final_version
}

TYPES=(
    "major"
    "minor"
    "patch"
)

UPSTREAM="https://github.com/devfile/registry-support.git"

# $VERSION has to be set by the user in [major].[minor].[patch] format
if [ -z "${VERSION}" ]; then 
    echo "Environment variable \$VERSION not set. Aborting ..."
    exit 1
fi

if [[ ! $VERSION =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Environment variable \$VERSION set to "$VERSION". Must be in [major].[minor].[patch] format ..."
    exit 1
fi

# RELEASE_CANDIDATE should be set to true only for major version release candidates
if [ -z "${RELEASE_CANDIDATE}" ]; then 
    echo "Environment variable \$RELEASE_CANDIDATE not set. Defaulting to false ..."
    RELEASE_CANDIDATE=false
fi

# RELEASE_TYPE should be one of $TYPES defined above
if [ -z "${RELEASE_TYPE}" ]; then 
    echo "Environment variable \$RELEASE_TYPE not set. Aborting ..."
    exit 1
else
    found=false
    for type in "${TYPES[@]}"; do
        if [ "$type" == "$RELEASE_TYPE" ]; then
            found=true
            break
        fi
    done

    if [ "$found" == "false" ]; then
        echo "Environment variable \$RELEASE_TYPE set to: "${RELEASE_TYPE}". Must be one of: "${TYPES[@]}" ..."
        exit 1
    fi
fi

# Set upstream repo tracking if not already set
upstream_name=$(git remote -v | awk -v url="$UPSTREAM" '$2 == url {print $1; exit}')

if [ -n "$upstream_name" ]; then
  echo "Upstream repo found ... Name = $upstream_name, url=$UPSTREAM"
else
  echo "Upstream not set ..."
  echo "Setting upstream to ... Name = release-upstream, url=$UPSTREAM"
  git remote add release-upstream $UPSTREAM
  upstream_name="release-upstream"
fi

if [ "${RELEASE_TYPE}" == "major" ] && [ "${RELEASE_CANDIDATE}" == "true" ]; then
    # the release associated with this tag will be a pre-release, and we should be moving the code to a rc/<name> branch alongside the prev release
    fetch_push_prior_release
    git push $upstream_name $upstream_name/main:refs/heads/rc/v$VERSION
    tag_and_push rc
elif [ "${RELEASE_TYPE}" == "patch" ]; then
    tag_and_push
else
    # major/minor normal workflow
    fetch_push_prior_release
    tag_and_push
fi