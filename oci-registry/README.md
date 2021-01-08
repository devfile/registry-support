# oci-registry

This folder contains the Dockerfile for the OCI registry server. It is based off of the [reference implementation from Docker](https://github.com/docker/distribution), but using a UBI-8 base image rather than Alpine.

## Build

To build the image, run `build.sh`.

## Deploy

To deploy this image as part of a Devfile registry:

1. Build and push this image to an image registry.
2. Install the [Devfile Registry Operator](https://github.com/devfile/registry-operator) on a Kubernetes cluster.
3. Create a `DevfileRegistry` yaml file and set `spec.ociRegistryImage` to the name of your pushed image from the previous step.
4. Run `kubectl apply -f <devfile-registry yaml>`