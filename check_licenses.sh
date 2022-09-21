#!/bin/bash
# This script runs license checks on go files to determine if they are missing

if ! command -v addlicense 2> /dev/null
then
  echo "error addlicense must be installed for this rule: go install -u github.com/google/addlicense" && exit 1
elif ! addlicense -check -f license_header.txt $(find . -not -path '*/\.*' -not -path '*/vendor/*' -name '*.go'); then
    echo "Licenses are not formatted; run './add_licenses.sh'"; exit 1 ;
fi

