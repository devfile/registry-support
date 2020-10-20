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

### OpenShift

1) Run `oc apply -f deploy/openshift/`

### Kubernetes

1) Set the value of `$HOST` in `deploy/kubernetes/ingress.yaml` to your ingress domain (such as 192.168.1.1.nip.io)
2) Run `kubectl apply -f deploy/kubernetes/`