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

# This script runs license checks on go files to determine if they are missing

if ! command -v addlicense 2> /dev/null
then
  echo "error addlicense must be installed with this command: go install github.com/google/addlicense@latest" && exit 1
elif ! addlicense -check -f license_header.txt $(find . -not -path '*/\.*' -not -path '*/vendor/*' -name '*.go'); then
    echo "Licenses are not formatted; run './add_licenses.sh'"; exit 1 ;
fi

