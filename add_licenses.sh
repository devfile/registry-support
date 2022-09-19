#!/bin/bash
# This script adds license headers that are missing from go files


if ! command -v addlicense 2> /dev/null
then
  echo "error addlicense must be installed for this rule: go get -u github.com/google/addlicense" && exit 1
else
  echo 'addlicense -v -f license_header.txt **/*.go'
  addlicense -v -f license_header.txt $(find . -not -path '*/\.*' -not -path '*/vendor/*' -name '*.go')
fi


