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

# Index Server build stage
FROM registry.access.redhat.com/ubi8/go-toolset:1.24.4-1754273152@sha256:ddad134f9fe193a6b3c7a87fa73b17e80212cdb0aa10dbd5a7581df96c998f1b AS index-builder
USER root
WORKDIR /tools
COPY . .
RUN CGO_ENABLED=0 go build -mod=vendor -o index-server main.go

FROM registry.access.redhat.com/ubi8/ubi-minimal:8.10@sha256:000dd8e4a3046ac7c47e65bbe01efc48d7a568e5ee9946cca1d74a7abf042d36 AS license
ARG LICENSE_REPO="devfile/registry-support"
ARG LICENSE_REF="main"
USER root

# Fetch license
RUN mkdir -p /licenses
RUN curl -sL https://raw.githubusercontent.com/${LICENSE_REPO}/refs/heads/${LICENSE_REF}/LICENSE -o /licenses/LICENSE

# Application image
FROM registry.access.redhat.com/ubi8/ubi-minimal:8.10@sha256:000dd8e4a3046ac7c47e65bbe01efc48d7a568e5ee9946cca1d74a7abf042d36 AS runner
USER root

# Install and configure dependencies
RUN microdnf update -y && microdnf install shadow-utils findutils && rm -rf /var/cache/yum
COPY entrypoint.sh /
RUN chmod +x /entrypoint.sh

# Copy index server
COPY --from=index-builder /tools/index-server /registry/index-server
RUN chgrp -R 0 /registry && \
    chmod -R g=u /registry

# Create a non-root user to run the server as
RUN set -x ; \
    adduser www-data -u 1001 -G root && exit 0

# Modify the permissions on the necessary files to allow the container to properly run as a non-root UID
RUN mkdir -p /www/data && chmod -R g+rwx /www/data

# Add license
RUN mkdir -p /licenses
COPY --from=license /licenses/LICENSE /licenses/LICENSE

# disable http/2 on the index server by default
ARG ENABLE_HTTP2=false
ENV ENABLE_HTTP2=${ENABLE_HTTP2}

# Set env vars for the locations of the devfile stacks and index.json
ENV DEVFILE_STACKS /registry/stacks
ENV DEVFILE_SAMPLES /registry/samples
ENV DEVFILE_INDEX /registry/index.json
ENV DEVFILE_BASE64_INDEX /www/data/index_base64.json
ENV DEVFILE_SAMPLE_INDEX /www/data/sample_index.json
ENV DEVFILE_SAMPLE_BASE64_INDEX /www/data/sample_base64_index.json
ENV DEVFILE_STACK_INDEX /www/data/stack_index.json
ENV DEVFILE_STACK_BASE64_INDEX /www/data/stack_base64_index.json

USER 1001

EXPOSE 8080

ENTRYPOINT ["/entrypoint.sh"]
