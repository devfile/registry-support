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

TYPES=(
    "major"
    "minor"
)

UPSTREAM="https://github.com/devfile/registry-support.git"


# $VERSION has to be set by the user in [major].[minor].[patch] format
if [ -z "${VERSION}" ]; then 
    echo "Environment variable \$VERSION not set. Aborting ..."
    exit 1
fi

# RELEASE_CANDIDATE should be set to true for major version release candidates
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

exit 0

# Set upstream repo tracking if not already set
upstream_name=$(git remote -v | awk -v url="$UPSTREAM" '$2 == url {print $1; exit}')

# Setup upstream if not set
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
    echo "Major release and release-candidate"
    fetch_push_prior_commit
    git push $upstream_name $upstream_name/main:rc/$VERSION
    git tag $VERSION-rc
    git push $upstream_name $VERSION-rc
else
    # major/minor normal workflow
    echo "Major or Minor release"
    fetch_push_prior_commit
    # Create new tag in upstream
    git tag $VERSION
    git push $upstream_name $VERSION
fi

fetch_push_prior_commit () {
    git fetch $upstream_name --tags
    LATEST_TAG=$(git tag --sort=-v:refname | head -n 1)
    MODIFIED_TAG=$(echo "$LATEST_TAG" | awk -F. '{print $1 "." $2 ".x"}') # convert to [major].[minor].x format from [major].[minor].[patch]
    git checkout -b test-fetch-tag $LATEST_TAG #TODO: fix the test-fetch-tag to something better
    git push $upstream_name release/$MODIFIED_TAG

    # navigate back to main
    git checkout main
}