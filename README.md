# registry-support

<div id="header">

![Go](https://img.shields.io/badge/Go-1.21-blue)
[![Apache2.0 License](https://img.shields.io/badge/license-Apache2.0-brightgreen.svg)](LICENSE)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/8257/badge)](https://www.bestpractices.dev/projects/8257)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/devfile/registry-support/badge)](https://securityscorecards.dev/viewer/?uri=github.com/devfile/registry-support)

</div>

Provide support for devfile registries

Issue tracking repo: https://github.com/devfile/api with label area/registry

## Build

### Prerequisite

The current release relies on [oapi-codegen 1.12.4](https://github.com/deepmap/oapi-codegen/tree/v1.12.4) for OpenAPI source generation. See the [Index Server README](index/server/README.md#source-generation) for more information.

To install, run:
`go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.12.4`

### Instructions
If you want to run the build scripts with Podman, set the environment variable
`export USE_PODMAN=true`

To build all of the components together (recommended) for dev/test, run `bash ./build_registry.sh` to build a Devfile Registry index image that is populated with the mock devfile registry data under `tests/registry/`.

By default `bash ./build_registry.sh` will build for `linux/amd64` architectures. To build for a different architecture pass in an argument to the script.
E.g. `bash ./build_registry.sh linux/arm64`.

Once the container has been pushed, you can push it to a container registry of your choosing with the following commands:

```
docker tag devfile-index <registry>/<username>/devfile-index:latest
```

followed by

```
docker push <registry>/<username>/devfile-index:latest
```

See the following for more on the component specific build process:

- [Building the OCI Registry](oci-registry/README.md#build)
- [Building the Index Server](index/server/README.md#build)

## Deploy

### Via the Devfile Registry Operator

We recommend using the [Devfile Registry Operator](https://github.com/devfile/registry-operator) to install a Devfile Registry on your Kubernetes or OpenShift cluster. Consult [its Readme for more information](https://github.com/devfile/registry-operator#running-the-controller-in-a-cluster).

### Via the Devfile Registry Helm Chart
Alternatively, a Helm chart is also provided if you do not wish to use an operator. You can find instructions below for installing via Helm to either a Kubernetes or OpenShift environment. You can find detailed instructions [here](deploy/chart/devfile-registry/README.md).

## Contributing

Please see our [contributing.md](./CONTRIBUTING.md).
