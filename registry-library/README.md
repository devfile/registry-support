# Devfile registry library

## Overview
Devfile registry library is used for interacting with devfile registry, consumers can use devfile registry library to list stacks and/or samples of devfile registry, download the stack devfile and the whole stack from devfile registry.

## What's included
`./library`: package `library` which contains devfile registry library constants, variables and functions, documentations can be found [here](https://pkg.go.dev/github.com/devfile/registry-support/registry-library/library)

`./build.sh`: build script to build `registry` binary to interact with devfile registry

`./registry`: `registry` binary to interact with devfile registry

## How to use it
1. Import devfile registry library
```go
import (
    registryLibrary "github.com/devfile/registry-support/registry-library/library"
)
```
2. Invoke devfile registry library

    a. Get the index of devfile registry for various devfile types
    ```go
    registryIndex, err := registryLibrary.GetRegistryIndex(registryURL, false, telemetryClient, StackDevfileType)
	if err != nil {
		return err
	}
    ```
    b. Get the indices of multiple devfile registries for various devfile types
    ```go
    registryList := GetMultipleRegistryIndices(registryURLs, skipTLSVerify, telemetryClient, StackDevfileType)
    ```
    c. Download the stack devfile from devfile registry
    ```go
	err := registryLibrary.PullStackByMediaTypesFromRegistry(registry, stack, registryLibrary.DevfileMediaTypeList, destDir, skipTLSVerify, telemetryClient)
	if err != nil {
		return err
	}
    ```
    d. Download the whole stack from devfile registry
    ```go
    err := registryLibrary.PullStackFromRegistry(registry, stack, destDir, skipTLSVerify, telemetryClient)
    if err != nil {
		return err
	}
    ```