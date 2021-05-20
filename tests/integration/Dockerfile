#
# Copyright (c) 2021 Red Hat, Inc.
# This program and the accompanying materials are made
# available under the terms of the Eclipse Public License 2.0
# which is available at https://www.eclipse.org/legal/epl-2.0/
#
# SPDX-License-Identifier: EPL-2.0
#
FROM quay.io/openshift-pipeline/golang:1.15-alpine

WORKDIR /registry-test

COPY . .

# Build the registry library and add to path
RUN cd /registry-test/registry-library && ./build.sh && cp /registry-test/registry-library/registry-library /usr/local/bin/registry-library

# Build the test binary
RUN /registry-test/build.sh

CMD /registry-test/devfileregistry-integration