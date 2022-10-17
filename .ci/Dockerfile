#
#   Copyright 2021-2022 Red Hat, Inc.
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.
FROM quay.io/openshift-pipeline/golang:1.15-alpine AS registry-builder

# Install dependencies
RUN apk add --no-cache git bash curl zip

# Install yq
RUN curl -sL -O https://github.com/mikefarah/yq/releases/download/v4.9.5/yq_linux_amd64 -o /usr/local/bin/yq && mv ./yq_linux_amd64 /usr/local/bin/yq && chmod +x /usr/local/bin/yq
RUN yq

# Copy the registry build tools
COPY build-tools /build-tools
# Copy the index generator source
COPY index/ /index
COPY tests/registry /registry

# Download offline starter projects
RUN bash /build-tools/dl_starter_projects.sh go-starter community

# Run the registry build tools
RUN /build-tools/build.sh /registry /build

FROM registry.access.redhat.com/ubi8/nodejs-16-minimal AS viewer-builder
USER root
ARG public_analytics_write_key=
ARG headless=true
ARG static_files_path=

# error if headless is true and static files are specified
RUN if [ "${headless}" == "true" ] && [ ! -z "${static_files_path}" ]; then exit 1; fi

# Environment build variables
ENV NEXT_PUBLIC_BASE_PATH /viewer
ENV NEXT_PUBLIC_ANALYTICS_WRITE_KEY ${public_analytics_write_key}

# Install git build dependency
RUN if [ "${headless}" == "false" ] && [ -z "${static_files_path}" ]; then microdnf update -y && microdnf install git && rm -rf /var/cache/yum; fi

# Install yarn global dependency
RUN if [ "${headless}" == "false" ]; then npm install -g yarn; fi

# Copy the registry build tools
COPY --from=registry-builder /build-tools /build-tools

# Copy static files into export directory, if blank nothing will be copied
COPY ${static_files_path} /app/dist/apps/registry-viewer/exported

# Remove any junk files from export directory if headless
RUN if [ "${headless}" == "true" ]; then rm -rf /app/dist/apps/registry-viewer/exported/*; fi

# Build & export registry viewer script if static files path is not specified
WORKDIR /build-tools
RUN if [ "${headless}" == "false" ] && [ -z "${static_files_path}" ]; then bash build_viewer.sh /app; fi

FROM devfile-index-base

# Copy registry content
COPY --from=registry-builder /build /registry

# Copy viewer content (empty if headless)
COPY --from=viewer-builder /app/dist/apps/registry-viewer/exported /www/viewer
