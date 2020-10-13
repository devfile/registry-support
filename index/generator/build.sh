#!/bin/sh
MODULE="github.com/devfile/registry-support/index/generator"
BIN_DIR=$GOPATH/bin
MODULE_BIN=$BIN_DIR/generator

CGO_ENABLED=0 go install $MODULE
cp $MODULE_BIN ./index-generator
