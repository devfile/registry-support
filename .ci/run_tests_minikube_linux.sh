#!/bin/bash

#
# Copyright Red Hat
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# NOTE: This script assumes that minikube is installed and running, and using the docker driver on Linux
# Due to networking issues with the docker driver and ingress on macOS/Windows, this script must be run on Linux

# Share docker env with Minikube
eval $(minikube docker-env)

# error on unset variables
set -u
# print each command before executing it
set -x

# Build the test devfile registry image
BASE_IMAGE=devfile-index-base bash ./build_registry.sh
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
  # Return the logs of the 3 containers in case the condition is not met
  echo "devfile-registry container logs:"
  kubectl logs -l app=devfile-registry --container devfile-registry
  echo "oci-registry container logs:"
  kubectl logs -l app=devfile-registry --container oci-registry
  echo "registry-viewer container logs:"
  kubectl logs -l app=devfile-registry --container registry-viewer
  # Return the description of every pod
  kubectl describe pods
  exit 1
fi

# Get the ingress URL for the registry
export REGISTRY=http://$(kubectl get ingress devfile-registry -o jsonpath="{.spec.rules[0].host}")

# Run the integration tests
cd tests/integration
bash ./docker-build.sh
docker run --env REGISTRY=$REGISTRY --env IS_TEST_REGISTRY=true devfile-registry-integration
