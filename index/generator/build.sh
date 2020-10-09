#!/usr/bin/env bash
MODULE="github.com/devfile/registry-support/index/generator"
BIN_DIR=$GOPATH/bin
MODULE_BIN=$BIN_DIR/generator

go install $MODULE
cp $MODULE_BIN ./index-generator
