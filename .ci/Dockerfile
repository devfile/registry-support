#
# Copyright (c) 2021 Red Hat, Inc.
# This program and the accompanying materials are made
# available under the terms of the Eclipse Public License 2.0
# which is available at https://www.eclipse.org/legal/epl-2.0/
#
# SPDX-License-Identifier: EPL-2.0
#
FROM quay.io/openshift-pipeline/golang:1.15-alpine AS builder

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

FROM devfile-index-base

COPY --from=builder /build /registry