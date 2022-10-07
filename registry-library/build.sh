#!/bin/sh
CGO_ENABLED=0 go build -mod=vendor -o registry-library .
