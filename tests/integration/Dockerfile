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
FROM registry.access.redhat.com/ubi8/go-toolset:1.25.9-1778693610@sha256:a9cc090ca7a3e5b059275faa53197fe4e7b13edcb2b665572493dff61c675bb0 AS builder
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

FROM registry.access.redhat.com/ubi8/ubi-minimal:8.10-1778735208@sha256:f0ed53122215c6cccdc992166caa9f6b56c08b202e8b1144dd1c2ad04579415d AS license
ARG LICENSE_REPO="devfile/registry-support"
ARG LICENSE_REF="main"
USER root

# Fetch license
RUN mkdir -p /licenses
RUN curl -sL https://raw.githubusercontent.com/${LICENSE_REPO}/refs/heads/${LICENSE_REF}/LICENSE -o /licenses/LICENSE

FROM registry.access.redhat.com/ubi8/ubi-minimal:8.10-1778735208@sha256:f0ed53122215c6cccdc992166caa9f6b56c08b202e8b1144dd1c2ad04579415d
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