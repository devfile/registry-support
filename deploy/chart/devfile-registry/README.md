# Devfile Registry

## Chart details
Installing this chart will deploy an OCI-based devfile registry on to your Kubernetes cluster, exposed over a single ingress endpoint.

## Prerequisites
- A Kubernetes or OpenShift cluster.
- Helm CLI, version 3 or higher
- Knowledge of your cluster's ingress domain

## Installing the Devfile Registry

Run the following command to install the devfile registry on to your cluster:

```
helm install <release-name> <path-to-chart> --set global.ingress.domain=<ingress-domain>
```

E.g. if your cluster's ingress domain is 192.168.1.0.nip.io, you would run:
```
helm install devfile-registry deploy/chart/devfile-registry --set global.ingress.domain=192.168.1.0.nip.io
```

## Updating the Devfile Registry

If you wish to update the devfile registry (such as to add change the devfile index image or change some configurations), you can run the following command:

```bash
helm upgrade <release-name> <path-to-chart> [--set options]
```

For example, updating the devfile index image of the devfile registry `my-registry` might look like:
```bash
helm upgrade my-registry deploy/chart/devfile-registry --set devfileIndex.image=docker.io/myuser/devfile-index --set defileIndex.tag=2.0
```

Alternatively to using `--set`, you can change the fields in `values.yaml` and then run the `helm upgrade` command.

## Uninstalling the Devfile Registry

To uninstall or delete the devfile registry Helm release, run:

```bash
helm uninstall <release-name>
```

## Configuration

The following fields can be configured in the Helm chart, either via the `values.yaml` file or with the `--values` flag in the `helm` CLI.

| Parameter                              | Description                                     | Default                                                    |
| -----------------------                | ---------------------------------------------   | ---------------------------------------------------------- |
| `global.ingress.domain`                | Ingress domain for the devfile registry         | **MUST BE SET BY USER**     |
| `global.ingress.class`                 | Ingress class for the devfile registry          | `nginx` |
| `global.ingress.secretName`            | Name of an existing tls secret if using TLS     | ` '' ` |
| `global.isOpenShift  `                 | Set to true to use OpenShift routes instead of ingress   | `false` |
| `devfileIndex.image`                   | Image used for the devfile index image          | `quay.io/devfile/devfile-index` |
| `devfileIndex.tag`                     | Tag for devfile index image                     | `next` |
| `devfileIndex.imagePullpolicy`         | Image pull policy for devfile index image       | `Always` |
| `devfileIndex.memoryLimit`             | Memory for devfile index container              | `256Mi` |
| `ociRegistry.image`                    | Image used for the oci registry image           | `quay.io/devfile/oci-registry` |
| `ociRegistry.tag`                      | Tag for oci registry image                      | `next` |
| `ociRegistry.imagePullpolicy`          | Image pull policy for oci registry image        | `Always` |
| `ociRegistry.memoryLimit`              | Memory for oci registry container               | `256Mi` |
| `persistence.enabled`                  | Enable persistent storage for the registry      | `true` |
| `persistence.size`                     | The size of the persistent volume (if-enabled)  | `1Gi` |