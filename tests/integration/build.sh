#!/bin/sh

CGO_ENABLED=0 go test -v -c -o devfileregistry-integration ./cmd/devfileregistry_test.go