# Devfile registry index server

## Overview

Provides REST API support for devfile registries and serves [devfile registry viewer](https://github.com/devfile/registry-viewer) client.

For more information on REST API docs: [registry-REST-API.adoc](registry-REST-API.adoc)

## Defining Endpoints

Edit the OpenAPI spec `openapi.yaml`, under `paths` you can define your endpoint, e.g. `GET /foo`:

```yaml
paths:
  /foo:
    get:
      summary: <short summary of what your endpoint does>
      description: <a long description of what your endpoint does>
      # 'serveFoo' points to handler function 'ServeFoo'
      operationId: serveFoo
      parameters: # the OpenAPI specifications of the endpoint parameters
        # spec for passing a bar query parameter /foo?bar=<string>
        - name: bar
          in: query
          description: <description for parameter>
          required: false
          schema:
            type: string
      responses: # the OpenAPI specifications for the endpoint responses
        default:
          description: <description of the response>
          content:
            # Content type(s)
            text/html: {}
```

See [swagger.io/docs](https://swagger.io/docs/specification/paths-and-operations) for more information.

## Build

The registry index server is built into a container image, `devfile-index-base:latest`, by running the following script:

```sh
bash build.sh
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
bash push.sh <new-image-tag>
```

For example, if the image repository is quay.io then use the pattern `quay.io/<user>/devfile-index-base`:

```sh
bash push.sh quay.io/someuser/devfile-index-base
```

### Source Generation

Index server build uses the CLI tool `oapi-codegen` to generate the schema types `pkg/server/types.gen.go` and endpoint definition `pkg/server/endpoint.gen.go` sources. When changing the OpenAPI specification, such as [defining endpoints](#defining-endpoints), it is required to regenerate these changes into the source.

The source generation can be done by manually building the index server with:

```bash
bash build.sh
```

or to just generate the source files by running:

```bash
bash codegen.sh
```

**Important**: When committing to this repository, it is _required_ to include the up to date source generation in your pull requests. Not including up to date source generation will lead to the PR check to fail.

### Enabling HTTP/2 on the Index Server

By default, http/2 on the index server is disabled due to [CVE-2023-44487](https://github.com/advisories/GHSA-qppj-fm5r-hxr3).

If you want to enable http/2, build with `ENABLE_HTTP2=true bash build.sh`.

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
