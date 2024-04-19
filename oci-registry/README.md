# oci-registry

This folder contains the Dockerfile for the OCI registry server. It is based off of the [reference implementation from Docker](https://github.com/docker/distribution), but using a UBI-8 base image rather than Alpine.

## Build
The scripts in this project support both `Docker` and `Podman` container engines. By default the scripts will run using `Docker`, to use `Podman` first run `export USE_PODMAN=true`.

To build the image, run `bash build.sh`.

To push the image to a repository of your choice, you can run `bash push.sh <repository-tag>`.

## Deploy

To deploy this image as part of a Devfile registry:

1. Build and push this image to an image registry.
2. Install the [Devfile Registry Operator](https://github.com/devfile/registry-operator) on a Kubernetes cluster.
3. Create a `DevfileRegistry` yaml file and set `spec.ociRegistryImage` to the name of your pushed image from the previous step.
4. Run `kubectl apply -f <devfile-registry yaml>`