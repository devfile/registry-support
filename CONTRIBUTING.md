# Contributing

Thank you for your interest in contributing to the Devfile Registry! We welcome your additions to this project.

## Code of Conduct

Before contributing to this repository for the first time, please review our project's [Code of Conduct](https://github.com/devfile/api/blob/main/CODE_OF_CONDUCT.md)

## Certificate of Origin

By contributing to this project you agree to the Developer Certificate of
Origin (DCO). This document was created by the Linux Kernel community and is a
simple statement that you, as a contributor, have the legal right to make the
contribution. See the [DCO](./DCO) file for details.

## How to Contribute:

### Issues

If you spot a problem with the devfile registry, [search if an issue already exists](https://github.com/devfile/api/issues). If a related issue doesn't exist, you can open a new issue using a relevant [issue form](https://github.com/devfile/api/issues/new/choose).

You can tag Devfile Registry related issues with the `/area registry` text in your issue.

### Development

#### Repository Format

The `registry-support` repository for the devfile registry is a mono-repo of sorts and there are a number of individual components contained within.

- [Devfile Registry Build Tools](./build-tools) - Tools for building devfile registries into container images.
- [Devfile Registry Helm Chart](./deploy/chart/devfile-registry) - Helm chart for deploying the devfile registry on to Kubernetes.
- [Index Generator](./index/generator) - The index generator tool for generation index.json files from registry data
- [Index Server](./index/server) - The index server, one of two servers that the devfile registry runs
- [OCI Registry Server](./oci-registry) - The OCI registry server, the second of two servers that the devfile registry runs.

Each of these individual components will have their own readme with instructions for consuming or developing the component, and it's recommended to consult them before development.


#### Building All Components Together

This repository contains a handy script that will build all of the individual components of the devfile registry (build-tools, index generator, index server, registry viewer, etc) and produce a deployable
container image containing test devfile data.

### Testing your Changes

All changes delivered for the Devfile Registry are expected to be sufficiently tested. This may include validating that existing tests pass, updating tests, or adding new tests.
Some of the components in this repository may have their own tests, others may just be covered by the repository's integration tests.

#### Unit Tests

Unit tests for each component are denoted by files with the `_test.go` suffix.

#### Integration Tests

The integration tests for this repository are located under the `tests/integration` folder and contain tests that validate the Operator's functionality when running on an OpenShift cluster.

To run these tests, consult the integration test's [readme](./tests/integration).

### Submitting Pull Request

**Note:** All commits must be signed off with the footer:
```
Signed-off-by: First Lastname <email@email.com>
```

You can easily add this footer to your commits by adding `-s` when running `git commit`. When you think the code is ready for review, create a pull request and link the issue associated with it.

Owners of the repository will watch out for and review new PRs. 

By default for each change in the PR, GitHub Actions and OpenShift CI will run checks against your changes (linting, unit testing, and integration tests).

If comments have been given in a review, they have to be addressed before merging.

After addressing review comments, don’t forget to add a comment in the PR afterward, so everyone gets notified by Github and know to re-review.


# Contact us

If you have questions, please visit us on `#devfile` in the [Kubernetes Slack](https://slack.k8s.io).
