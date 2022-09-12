# Devfile registry library

## Overview
The Devfile registry library is used to interact with the devfile registry to perform the following actions:

* List the indices of stacks and/or samples from a single registry or across multiple registries
* Download a stack with specific media types or all supported media types
* Send telemetry to the Devfile Registry service
* Filter stacks based on architectures

## What's included
`./library`: package `library` which contains devfile registry library constants, variables and functions. Documentation can be found [here](https://pkg.go.dev/github.com/devfile/registry-support/registry-library/library)

`./build.sh`: build script to build the `registry-library` binary to interact with devfile registry

`./registry-library`: `registry-library` binary to interact with devfile registry

## How to use it
1. Import the devfile registry library and index schema
   ```go
   import (
       registryLibrary "github.com/devfile/registry-support/registry-library/library"
       indexSchema "github.com/devfile/registry-support/index/generator/schema"
   )
   ```

### List the indices of stacks and/or samples
1. Get the index for stack devfile types from a single registry

    ```go
    registryURL := "https://registry.devfile.io"
    options := registryLibrary.RegistryOptions{} //leave empty if telemetry and architecture types are not relevant
    registryIndex, err := registryLibrary.GetRegistryIndex(registryURL, options, indexSchema.StackDevfileType)
    if err != nil {
        return err
    }
    ```
2. Get the index for all devfile types from a single registry
    ```go
   devfileTypes := []indexSchema.DevfileType{indexSchema.StackDevfileType, indexSchema.SampleDevfileType}
   registryIndex, err := registryLibrary.GetRegistryIndex(registryURL, options, devfileTypes...)
   if err != nil {
   return err
   }
    ```

3. Get the indices for various devfile stacks from multiple devfile registries
    ```go
    registryList := GetMultipleRegistryIndices(registryURLs, options, indexSchema.StackDevfileType)
    ```
#### Download the stack 
Supported devfile media types can be found in the latest version of [library.go](https://github.com/devfile/registry-support/blob/main/registry-library/library/library.go)
1. Download a stack devfile with a given media type from the devfile registry
    ```go
    stack := "java-springboot"
    destDir := "."
    err := registryLibrary.PullStackByMediaTypesFromRegistry(registryURL, stack, registryLibrary.DevfileMediaTypeList, destDir, options)
    if err != nil {
        return err
    }
    ```
   
2. Download a stack devfile with a given version and media type from the devfile registry
    ```go
    stack := "java-springboot:1.0.0" // java-springboot is stack name, 1.0.0 is stack version
    destDir := "."
    err := registryLibrary.PullStackByMediaTypesFromRegistry(registryURL, stack, registryLibrary.DevfileMediaTypeList, destDir, options)
    if err != nil {
        return err
    }
    ```
   
3. Download a stack with all supported media types from the devfile registry
    ```go
    err := registryLibrary.PullStackFromRegistry(registryURL, stack, destDir, options)
    if err != nil {
        return err
    }
    ```

#### Specify Registry Options
1. Test a pre-prod registry installed with self-signed certificates
    ```go
    options := registryLibrary.RegistryOptions{
        SkipTLSVerify: "true",
    }
    ```
2. Filter Devfiles based on a set of architectures found in [header.go](https://github.com/devfile/api/blob/main/pkg/devfile/header.go)
    ```go
    architectures := []string{"amd64", "arm64"}
    options := registryLibrary.RegistryOptions{
        Filter: registryLibrary.RegistryFilter{
            Architectures: architectures,
        },
    }
    ```
3. Send Telemetry data to the Devfile Registry
    ```go
    options := registryLibrary.RegistryOptions{
        Telemetry: registryLibrary.TelemetryData{
            User:   "user-name" //this can be a generated UUID
            Locale: "en_US" // set the OS or browser locale
            Client: "client-name" //the name of the client
        }
    } 
   ```
4. Get v2index with versions information from the Devfile Registry
    ```go
    options := registryLibrary.RegistryOptions{
         NewIndexSchema: true
    }
   ```
5. Filter Devfiles based the min and max devfile schema version provided
    ```go
    options := registryLibrary.RegistryOptions{
    	NewIndexSchema: true,
        Filter: registryLibrary.RegistryFilter{
             // devfile schema version range is [2.1, 2.2], inclusive
             MinSchemaVersion: "2.1",
             MaxSchemaVersion: "2.2"
        },
    }
    ```
6.  Override the HTTP request and response timeout values
   ```go
   customTimeout := 20
   options := registryLibrary.RegistryOptions{
      HTTPTimeout: &customTimeout
   }
   ```

#### Download the starter project

1. Download starter project in-memory
```go
var bytes []byte
var err error

...

starterProject := "springbootproject"
bytes, err = registryLibrary.DownloadStarterProjectAsBytes(registryURL, 
    stack, starterProject, options)
if err != nil {
    return err
}
```
2. Download starter project archive to filesystem path
```go
starterProject := "springbootproject"
path := fmt.Sprintf("%s.zip", starterProject)
err := registryLibrary.DownloadStarterProject(path, registryURL, stack, starterProject, options)
if err != nil {
    return err
}
```

```sh
ls # => devfile.yaml springbootproject.zip
```
3. Download starter project archive and extract to filesystem path
```go
starterProject := "springbootproject"
path := "."
err := registryLibrary.DownloadStarterProjectAsDir(path, registryURL, stack, starterProject, options)
if err != nil {
    return err
}
```

```sh
ls # => devfile.yaml pom.xml HELP.md mvnw mvnw.cmd src 
```
