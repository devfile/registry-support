# oci-registry

This folder contains the Dockerfile for the OCI registry server. It is based off of the [reference implementation from Docker](https://github.com/docker/distribution), but using a UBI-8 base image rather than Alpine.

## Build
The scripts in this project support both `Docker` and `Podman` container engines. By default the scripts will run using `Docker`, to use `Podman` first run `export USE_PODMAN=true`.

The build script enables users to build for different architectures, by default running `bash build.sh` will build for `linux/amd64`. If you would like to build for a different architecture simply add it as an argument to the script. E.g. `bash build.sh linux/arm64` for `linux/arm64` builds.

To push the image to a repository of your choice, you can run `bash push.sh <repository-tag>`.

## Deploy

To deploy this image as part of a Devfile registry:

1. Build and push this image to an image registry.
2. Install the [Devfile Registry Operator](https://github.com/devfile/registry-operator) on a Kubernetes cluster.
3. Create a `DevfileRegistry` yaml file and set `spec.ociRegistryImage` to the name of your pushed image from the previous step.
4. Run `kubectl apply -f <devfile-registry yaml>`

## Troubleshooting

If you are trying to run `build-multi-arch.sh`, or you are trying to run `build.sh` for an architecture your machine is **not running** on, you may encounter an error similiar to this in your build:
```
Dockerfile:18
--------------------
  16 |     
  17 |     FROM registry.access.redhat.com/ubi8-minimal:8.2
  18 | >>> RUN microdnf update -y && rm -rf /var/cache/yum && microdnf install ca-certificates httpd-tools
  19 |     
  20 |     # Create a non-root user to run the server as
--------------------
ERROR: failed to solve: process "/bin/sh -c microdnf update -y && rm -rf /var/cache/yum && microdnf install ca-certificates httpd-tools" did not complete successfully: exit code: 1
```

This error can occur because your container engine is not properly supporting emulation. To check if this is the case you can run:
```
podman run --rm --privileged tonistiigi/binfmt
```
or if you are using docker:
```
docker run --rm --privileged tonistiigi/binfmt
```

This command will output something similar to the following where you can observe what architectures are currently supported for you:
```
"supported": [
    "linux/arm64",
    "linux/amd64",
    "linux/riscv64",
    "linux/ppc64le",
    "linux/s390x",
    "linux/386",
    "linux/mips64le",
    "linux/mips64"
  ]
```

If you do not see the architecture you are trying to build for in the output list you are probably missing emulation capabilities. The following command has been known to fix this issue and allow for emulation:
```
sudo apt-get install -y gcc-arm-linux-gnueabihf libc6-dev-armhf-cross qemu-user-static qemu-system-i386
```

The CI environments we call these scripts in are incorporating the use of QEMU. If the above command does not solve the issue for you it would be best to investigate into if QEMU is working properly in your environment.