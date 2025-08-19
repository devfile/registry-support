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
FROM registry.access.redhat.com/ubi8/go-toolset:1.23.6-1.1744600118 as registry
ARG PKG="github.com/docker/distribution"
ARG PKG_VERSION="v2.8.3"
ENV GOPATH=/go
USER root

RUN git clone https://github.com/distribution/distribution /go/src/${PKG} -b ${PKG_VERSION}

WORKDIR /go/src/${PKG}
RUN VERSION=$(git describe --match 'v[0-9]*' --dirty='.m' --always --tags) REVISION=$(git rev-parse HEAD)$(if ! git diff --no-ext-diff --quiet --exit-code; then echo .m; fi); \
  echo "-X ${PKG}/version.Version=${VERSION#v} -X ${PKG}/version.Revision=${REVISION} -X ${PKG}/version.Package=${PKG}" | tee /tmp/.ldflags; \
  echo -n "${VERSION}" | tee /tmp/.version;
RUN set -x ; CGO_ENABLED=0 GO111MODULE=auto go build -tags "include_oss,include_gcs" -trimpath -ldflags "$(cat /tmp/.ldflags) -s -w" -o /usr/bin/registry ./cmd/registry

FROM registry.access.redhat.com/ubi8-minimal:8.2
RUN microdnf update -y && rm -rf /var/cache/yum && microdnf install ca-certificates httpd-tools

# Create a non-root user to run the server as
RUN set -x ; \
    adduser registry -u 1001 -G root && exit 0

# Copy OCI registry binary from container image
COPY --from=registry --chown=registry:0 /usr/bin/registry /bin/registry

# Add license
RUN mkdir -p /licenses
COPY LICENSE /licenses

USER 1001
EXPOSE 5000
ENTRYPOINT ["registry"]
CMD ["serve", "/etc/docker/registry/config.yml"]