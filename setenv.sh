#!/bin/bash

# This script aliases the docker cli if the environment variable USE_PODMAN is set to true.

# default value is false if USE_PODMAN is unset or null
podman=${USE_PODMAN:-false}
if [ ${podman} == true ]; then
  alias docker=podman
  echo "setting alias docker=podman"
fi
