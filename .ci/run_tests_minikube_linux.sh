#!/bin/bash

# NOTE: This script assumes that minikube is installed and running, and using the docker driver on Linux
# Due to networking issues with the docker driver and ingress on macOS/Windows, this script must be run on Linux

# Share docker env with Minikube
eval $(minikube docker-env)

# error on unset variables
set -u
# print each command before executing it
set -x

# Build the test devfile registry image
./build_registry.sh
if [ $? -ne 0 ]; then
  echo "Error building devfile registry images"
  exit 1;
fi

# Deploy the devfile registry using the Helm chart
# Use the test registry image built in the previous step.
# Since minikube is running on Docker, we can specify a local image NOT pushed up to a registry
# This saves us a fair bit of hassle with having to dynamically push the test image to a container registry
helm install devfile-registry ./deploy/chart/devfile-registry --set global.ingress.domain="$(minikube ip).nip.io" \
	--set devfileIndex.image=devfile-index \
	--set devfileIndex.tag=latest \
	--set devfileIndex.imagePullPolicy=Never

# Wait for the registry to become ready
kubectl wait deploy/devfile-registry --for=condition=Available --timeout=600s
if [ $? -ne 0 ]; then
  kubectl get pods
  kubectl describe pods
  exit 1
fi

# Get the ingress URL for the registry
export REGISTRY=http://$(kubectl get ingress devfile-registry -o jsonpath="{.spec.rules[0].host}")

# Run the integration tests
cd tests/integration
./docker-build.sh
docker run --env REGISTRY=$REGISTRY devfile-registry-integration