# registry-support CI tests

This readme outlines how the tests for the registry-support repo runs on GitHub actions.

## Background

The `registry-support` git repository contains a number of related, but separate components that form the basis for oci-based devfile registries:

- Index server base image
- OCI-registry container image
- Registry Go library
- Index.json generator and schema
- Registry build tools
- Helm chart for deploying the devfile registry

Some of the components (such as the index generator) have their own individual unit tests, but because of all of their interconnected parts, in order to properly test them, they need to be tested together as part of a single running devfile registry.

The `run_tests_minikube_linux.sh` script tries to do that, in an environment that can run via Minikube on GitHub Actions. It does the following:

1) Builds the index server base image
2) Dynamically builds a mock devfile registry image, whose base image is the index server base image built in step **1**
3) Installs the devfile registry using the Devfile registry helm chart in this repository, using the devfile registry image built in step **2**
4) Runs the [devfile registry integration tests](tests/README.md) against the test devfile registry

## Why Minikube and Not OpenShift?

At this moment, due to how the OpenShift CI builds docker images, this test flow (building a base image, then building a child image that depends on the base image, and deploying/testing the child image) isn't currently possible on its infrastructure. 

It _may_ be possible to separately publish the PR's "test" devfile index image to quay.io using GitHub Actions, and run tests against it in the OpenShift CI (without using the OpenShift CI to build the image), but that will require further investigation.
