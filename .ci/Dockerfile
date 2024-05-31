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
FROM golang:1.21-alpine@sha256:3f8e3ad3e7c128d29ac3004ac8314967c5ddbfa5bfa7caa59b0de493fc01686a AS builder

# Allows for the proper yq to be downloaded
# This arg is automatically set if the dockerfile is built with --platform flag
ARG TARGETARCH=amd64

# Install dependencies
RUN apk add --no-cache git bash curl zip

# Install yq
ENV YQ_VERSION=v4.44.1
RUN curl -sL -O https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_linux_${TARGETARCH} -o /usr/local/bin/yq && mv ./yq_linux_${TARGETARCH} /usr/local/bin/yq && chmod +x /usr/local/bin/yq
RUN yq

# Copy the registry build tools
COPY build-tools /build-tools
# Copy the index generator source
COPY index/ /index
COPY tests/registry /registry

# Download all the offline parent devfiles
RUN bash /build-tools/dl_parent_devfiles.sh

# Download offline starter projects
RUN bash /build-tools/dl_starter_projects.sh go-starter community

# Run the registry build tools
RUN bash /build-tools/build.sh /registry /build

FROM devfile-index-base

COPY --from=builder /build /registry