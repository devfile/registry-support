# Devfile registry library

## Overview
Devfile registry library is used for interacting with devfile registry, consumers can use devfile registry library to list stacks and/or samples of devfile registry, download the stack devfile and the whole stack from devfile registry.

## What's included
`./library`: package `library` which contains devfile registry library constants, variables and functions, documentations can be found [here](https://pkg.go.dev/github.com/devfile/registry-support/registry-library/library)

`./build.sh`: build script to build `registry` binary to interact with devfile registry

`./registry`: `registry` binary to interact with devfile registry

## How to use it
1. Import devfile registry library
```
import (
    registryLibrary "github.com/devfile/registry-support/registry-library/library"
)
```
2. Invoke devfile registry library

    a. Get the index of devfile registry for various devfile types
    ```
    registryIndex, err := registryLibrary.GetRegistryIndex(registry, StackDevfileType, SampleDevfileType)
	if err != nil {
		return err
	}
    ```
    b. Get the indices of multiple devfile registries for various devfile types
    ```
    registryList := GetMultipleRegistriesIndices(registryURLs, StackDevfileType)
    ```
    c. Download the stack devfile from devfile registry
    ```
	err := registryLibrary.PullStackByMediaTypesFromRegistry(registry, stack, registryLibrary.DevfileMediaTypeList, destDir)
	if err != nil {
		return err
	}
    ```
    d. Download the whole stack from devfile registry
    ```
    err := registryLibrary.PullStackFromRegistry(registry, stack, destDir)
    if err != nil {
		return err
	}
    ```