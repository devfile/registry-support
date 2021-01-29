# registry-support

Provide support for devfile registries

Issue tracking repo: https://github.com/devfile/api with label area/registry

## Architecture

### Registry

A standard OCI registry, currently using the [reference implementation from Docker](https://hub.docker.com/_/registry). Devfile artifacts are fed into it via the `devfile-registry-metadata` container upon startup.

![Registry Architecture](https://user-images.githubusercontent.com/606959/88183913-5e417280-cc32-11ea-9509-651bb44d9280.png)

### Metadata Container

The `oci-devfile-registry-metadata` container is a sidecar deployed alongside the OCI registry that has two purposes:
1) Pre-loading the devfiles into the registry upon startup
2) Generating and hosting the [index.json](https://raw.githubusercontent.com/odo-devfiles/registry/master/devfiles/index.json) for devfile consumers

## Deploy

### Via the Devfile Registry Operator

We recommend using the [Devfile Registry Operator](https://github.com/devfile/registry-operator) to install a Devfile Registry on your Kubernetes or OpenShift cluster. Consult [its Readme for more information](https://github.com/devfile/registry-operator#running-the-controller-in-a-cluster).

## Via the Devfile Registry Helm Chart

Alternatively, a Helm chart is also provided if you do not wish to use an operator. To install (with Helm 3) run:

```bash
$ helm install devfile-registry deploy/chart/devfile-registry/ --set global.ingress.domain=<ingress-domain>
```

For more information on the Helm chart, consult [its readme](deploy/kubernetes/devfile-registry/README.md).