# registry-support

Provide support for devfile registries

Issue tracking repo: https://github.com/devfile/api with label area/registry

## Deploy

### Via the Devfile Registry Operator

We recommend using the [Devfile Registry Operator](https://github.com/devfile/registry-operator) to install a Devfile Registry on your Kubernetes or OpenShift cluster. Consult [its Readme for more information](https://github.com/devfile/registry-operator#running-the-controller-in-a-cluster).

### Via the Devfile Registry Helm Chart

Alternatively, a Helm chart is also provided if you do not wish to use an operator. To install (with Helm 3) run:

```bash
$ helm install devfile-registry deploy/chart/devfile-registry/ --set global.ingress.domain=<ingress-domain>
```

For more information on the Helm chart, consult [its readme](deploy/chart/devfile-registry/README.md).