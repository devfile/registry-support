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
FROM registry.access.redhat.com/ubi8/go-toolset:1.24.4-1754273152@sha256:ddad134f9fe193a6b3c7a87fa73b17e80212cdb0aa10dbd5a7581df96c998f1b AS builder
ARG LIBRARY_REPO="https://github.com/devfile/registry-support"
ARG LIBRARY_REF="main"
USER root
WORKDIR /registry-test

COPY . .

# Fetch library repository
RUN git clone ${LIBRARY_REPO} -b ${LIBRARY_REF} /registry-support

# Build the registry library and add to builder path
RUN cd /registry-support/registry-library && ./build.sh && cp /registry-support/registry-library/registry-library /usr/local/bin/registry-library

# Build the test binary
RUN /registry-test/build.sh

FROM registry.access.redhat.com/ubi8/ubi-minimal:8.10@sha256:000dd8e4a3046ac7c47e65bbe01efc48d7a568e5ee9946cca1d74a7abf042d36 AS license
ARG LICENSE_REPO="devfile/registry-support"
ARG LICENSE_REF="main"
USER root

# Fetch license
RUN mkdir -p /licenses
RUN curl -sL https://raw.githubusercontent.com/${LICENSE_REPO}/refs/heads/${LICENSE_REF}/LICENSE -o /licenses/LICENSE

FROM registry.access.redhat.com/ubi8/ubi-minimal:8.10@sha256:000dd8e4a3046ac7c47e65bbe01efc48d7a568e5ee9946cca1d74a7abf042d36
USER root
WORKDIR /registry-test

# Add registry library to runner path
COPY --from=builder /usr/local/bin/registry-library /usr/local/bin/registry-library

# Add test binary to working directory of runner
COPY --from=builder /registry-test/devfile-registry-integration /registry-test/devfile-registry-integration

# Add license
RUN mkdir -p /licenses
COPY --from=license /licenses/LICENSE /licenses/LICENSE

USER 1001
CMD /registry-test/devfile-registry-integration