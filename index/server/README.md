# Devfile registry index server

## Overview

Provides REST API support for devfile registries and serves [devfile registry viewer](https://github.com/devfile/registry-viewer) client.

For more information on REST API docs: [registry-REST-API.adoc](registry-REST-API.adoc)

## Build

The registry index server is built into a container image, `devfile-index-base:latest`, by running the following script:

```sh
bash index/server/build.sh
```

You retag it with one of the two command:

**Docker CLI**

```sh
docker tag devfile-index-base:latest <new-image-tag>
```

**Podman CLI**

```sh
podman tag devfile-index-base:latest <new-image-tag>
```

Push your image into the an image repository with the following:

```sh
bash index/server/push.sh <new-image-tag>
```

For example, if the image repository is quay.io then use the pattern `quay.io/<user>/devfile-index-base`:

```sh
bash index/server/push.sh quay.io/someuser/devfile-index-base
```

## Testing

Endpoint unit testing is defined under `pkg/server/endpoint_test.go` and can be performed by running the following:

```sh
go test pkg/server/endpoint_test.go
```

or by running all tests:

```sh
go test ./...
```

**Environment Variables**

- `DEVFILE_REGISTRY`: Optional environment variable for specifying testing registry path
    - default: `../../tests/registry`
