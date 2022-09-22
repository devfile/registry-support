#!/bin/bash
# This script runs license checks on go files to determine if they are missing

if ! command -v addlicense 2> /dev/null
then
  echo "error addlicense must be installed with this command: go install github.com/google/addlicense@latest" && exit 1
elif ! addlicense -check -f license_header.txt $(find . -not -path '*/\.*' -not -path '*/vendor/*' -name '*.go'); then
    echo "Licenses are not formatted; run './add_licenses.sh'"; exit 1 ;
fi

