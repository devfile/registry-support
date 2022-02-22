# Devfile Registry Integration Tests

This folder contains the integration tests for the OCI-based Devfile Registry. The tests can be run against either a remote devfile registry (such as https://registry.stage.devfile.io), or a local devfile registry running your machine (such as in Minikube, or Docker Desktop).

## Build
If you want to run the build scripts with Podman, set the environment variable
`export USE_PODMAN=true`

The integration tests can be built to either run in a Docker container, or locally on your machine.

To build the test docker image, run `./docker-build.sh`

To build the test binary locally, run: `./build.sh`

## Custom Tests

Some tests like using the arch filter are registry specific. For example, the community registry may not have devfiles with archs mentioned but the test registry in this repo does. As such, run these specific tests by setting the env  `IS_TEST_REGISTRY=true`

## Run in a Container

The recommended way to run the tests is in a container, simply run the following after building the image in the previous step:
```
docker run --env REGISTRY=$REGISTRY --env IS_TEST_REGISTRY=true devfile-registry-integration
```

Where `$REGISTRY` is the hostname of the devfile registry that you wish to test against (such as https://registry.devfile.io or http://devfile-registry-default.10.101.108.46.nip.io)

## Run Locally

To run the tests locally, you must make sure that the CLI version of the `registry-library` is built and exists on your system path, as the integration tests rely on it. To do that:

1. Navigate to the `registry-library/` directory in the root of this repository
2. Run the `build.sh` script in that folder
3. Run `cp registry-library /usr/local/bin/registry-library` to add it to your system path

Then, to run the tests, navigate back to the `tests/integration` folder and run:
1. `export REGISTRY=$REGISTRY` where `$REGISTRY` is the hostname of the devfile registry that you wish to test against (such as https://registry.stage.devfile.io or http://devfile-registry-default.10.101.108.46.nip.io).
2. `./devfile-registry-integration` to run the tests 
