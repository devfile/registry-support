# Devfile registry index server

## Overview

Provides REST API support for devfile registries and serves [devfile registry viewer](https://github.com/devfile/registry-viewer) client.

For more information on REST API docs: [registry-REST-API.adoc](registry-REST-API.adoc)

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
