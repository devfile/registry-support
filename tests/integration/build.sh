#!/bin/sh

CGO_ENABLED=0 go test -v -c -o devfile-registry-integration ./cmd/devfileregistry_test.go