# Devfile Registry Build Tools

This folder contains tools for building up a Devfile Registry Repository and packaging it and its generated index.json into a container image for deployment on an OCI Devfile Registry, hosted on Kubernetes. As we expand the functionality of the build tools and index generator, this will grow to include stack validation as well.

## How to Run

### Prerequisites

- Golang 1.13.x or higher
- Docker 17.05 or higher
- Git
- [yq](https://github.com/mikefarah/yq) 4.x

### Building the Devfile Registry

To build a devfile registry repository, run: `./build_image.sh <path-to-devfile-registry-folder>`.

The build script will build the index generator, generate the index.json from the specified devfile registry, and build the stacks and index.json into a devfile index container image.