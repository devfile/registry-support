#
#   Copyright 2020-2022 Red Hat, Inc.
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.

#!/bin/bash

# Set build folder
if [ -z "${1+x}" ]
then
    buildfolder="${PWD}/app"
else
    buildfolder=$1
fi

# Clone devfile-web repository
if [ -d $buildfolder ]
then
    rm -rf $buildfolder
fi
git clone https://github.com/devfile/devfile-web.git $buildfolder

# Set $buildfolder to working directory
cd $buildfolder

# Install project dependencies
$(npm get prefix)/bin/yarn install --frozen-lockfile
# Export static site
$(npm get prefix)/bin/yarn nx run registry-viewer:postexport --skip-nx-cache
