# Devfile Registry

## Chart details
Installing this chart will deploy an OCI-based devfile registry on to your Kubernetes cluster, exposed over a single ingress endpoint.

## Prerequisites
- A Kubernetes or OpenShift cluster.
- Helm CLI, version 3 or higher
- Knowledge of your cluster's ingress domain

## Installing the Devfile Registry on Kubernetes

Run the following command to install the devfile registry on to your Kubernetes Cluster:

```
helm install <release-name> <path-to-chart> --set global.ingress.domain=<ingress-domain>
```

E.g. if your cluster's ingress domain is 192.168.1.0.nip.io, you would run:
```
helm install devfile-registry deploy/chart/devfile-registry --set global.ingress.domain=192.168.1.0.nip.io
```

### Kubernetes Installation Examples
```bash
$ helm install devfile-registry ./deploy/chart/devfile-registry \ 
    --set global.ingress.domain=<ingress-domain> \
	--set devfileIndex.image=<index-image> \
	--set devfileIndex.tag=<index-image-tag>
```

Where `<ingress-domain>` is the ingress domain for your cluster, `<index-image>` is the devfile index image you want to deploy, and `<index-image-tag>` is the corresponding image tag for the devfile index image.

For example, if you're installing your own custom devfile registry image for dev/test purposes on Minikube, you might run:

```bash
$ helm install devfile-registry ./deploy/chart/devfile-registry \
    --set global.ingress.domain="$(minikube ip).nip.io" \
	--set devfileIndex.image=quay.io/someuser/devfile-index \
	--set devfileIndex.tag=latest
```

You can deploy a devfile registry with a custom registry viewer image (uses `quay.io/devfile/registry-viewer:next` by default) by running the following:

```bash
$ helm install devfile-registry ./deploy/chart/devfile-registry \
    --set global.ingress.domain="$(minikube ip).nip.io" \
	--set devfileIndex.image=quay.io/someuser/devfile-index \
	--set devfileIndex.tag=latest \
	--set registryViewer.image=quay.io/someuser/registry-viewer \
	--set registryViewer.tag=latest
```

You can deploy a *headless* devfile registry (i.e. without the registry viewer) by specifying `--set global.headless=true` which will look like:

```bash
$ helm install devfile-registry ./deploy/chart/devfile-registry \
    --set global.ingress.domain="$(minikube ip).nip.io" \
	--set global.headless=true \
	--set devfileIndex.image=quay.io/someuser/devfile-index \
	--set devfileIndex.tag=latest
```

## Installing the Devfile Registry on OpenShift

If you're installing on OpenShift, you will first need to set `global.isOpenShift` to true, for example:
```
helm install devfile-registry deploy/chart/devfile-registry --set global.isOpenShift=true
```

There are 3 ways that you can install the registry on OpenShift:
#### 1: Via Installation Script With OpenShift Generated Route Hostname and Domain

If you wish to take advantage of OpenShift's generated hostname and domain, all you need to run is:
```
$ bash ./helm-openshift-install.sh
```
This will install the Devfile Registry to OpenShift for you with the generated route hostname and domain. If you wish to include additional arguments such as changing the image for `devfileIndex`, you can include those alongside the script call:
```
$ bash ./helm-openshift-install.sh \
--set devfileIndex.image=quay.io/someuser/devfile-index \
--set devfileIndex.tag=latest
```
#### 2: Via Installation Script With Custom Hostname

Similar to the above instructions, you can set your own custom domain as part of the arguments to the installation script.
```
$ bash ./helm-openshift-install.sh --set global.route.domain=<domain> <other arguments>
```

#### 3: Via Helm CLI With Custom Hostname

If you do not wish to use a helper script to install to OpenShift, you are able to mimic the Kubernetes installation with one slight change. Instead of `--set global.ingress.domain` you will swap it with `--set global.route.domain` as OpenShift utilizes Routes instead of Ingress.
```bash
$ helm install devfile-registry ./deploy/chart/devfile-registry \ 
    --set global.route.domain=<route-domain> \
	--set devfileIndex.image=<index-image> \
	--set devfileIndex.tag=<index-image-tag>
```

Installing to OpenShift follows the same process as Kubernetes, you just need to ensure that `--set global.isOpenShift=true` is an argument to the install. Additionally, instead of `--set global.ingress.domain=<domain>` for Kubernetes you will instead include `--set global.route.domain=<domain>` for OpenShift. All other arguments are available to either Kubernetes or OpenShift.

## Updating the Devfile Registry

If you wish to update the devfile registry (such as to add change the devfile index image or change some configurations), you can run the following command:

```bash
helm upgrade <release-name> <path-to-chart> [--set options]
```

For example, updating the devfile index image of the devfile registry `my-registry` might look like:
```bash
helm upgrade my-registry deploy/chart/devfile-registry --set devfileIndex.image=docker.io/myuser/devfile-index --set devfileIndex.tag=2.0
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
| `global.ingress.domain`                | Kubernetes Ingress domain for the devfile registry                                        | **MUST BE SET BY USER**     |
| `global.ingress.class`                 | Ingress class for the devfile registry                                         | `nginx` |
| `global.route.domain`                  | OpenShift Route domain for the devfile registry                                          | **MUST BE SET BY USER**    |
| `global.ingress.secretName`            | Name of an existing tls secret if using TLS                                    | ` '' ` |
| `global.isOpenShift  `                 | Set to true to use OpenShift routes instead of ingress                         | `false` |
| `global.tlsEnabled`                    | Set to true to use the devfile registry with TLS                               | `false` |
| `global.headless`                      | Set to true to run in headless mode (does not expect attached registry viewer) | `false` |
| `devfileIndex.image`                   | Image used for the devfile index image                                         | `quay.io/devfile/devfile-index` |
| `devfileIndex.tag`                     | Tag for devfile index image                                                    | `next` |
| `devfileIndex.imagePullpolicy`         | Image pull policy for devfile index image                                      | `Always` |
| `devfileIndex.memoryLimit`             | Memory for devfile index container                                             | `256Mi` |
| `registryViewer.image`                 | Image used for the registry viewer image                                       | `quay.io/devfile/registry-viewer` |
| `registryViewer.tag`                   | Tag for registry viewer image                                                  | `next` |
| `registryViewer.imagePullpolicy`       | Image pull policy for registry viewer image                                    | `Always` |
| `registryViewer.memoryLimit`           | Memory for registry viewer container                                           | `256Mi` |
| `ociRegistry.image`                    | Image used for the oci registry image                                          | `quay.io/devfile/oci-registry` |
| `ociRegistry.tag`                      | Tag for oci registry image                                                     | `next` |
| `ociRegistry.imagePullpolicy`          | Image pull policy for oci registry image                                       | `Always` |
| `ociRegistry.memoryLimit`              | Memory for oci registry container                                              | `256Mi` |
| `persistence.enabled`                  | Enable persistent storage for the registry                                     | `true` |
| `persistence.size`                     | The size of the persistent volume (if-enabled)                                 | `1Gi` |
| `telemetry.key`                        | The write key for the Segment instance                                         | **MUST BE SET BY USER**  |
| `telemetry.registryViewerWriteKey`     | The write key for the registry viewer                                          | **MUST BE SET BY USER**  |