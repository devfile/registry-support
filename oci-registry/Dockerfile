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
FROM registry:2 as registry

FROM registry.access.redhat.com/ubi8-minimal:8.2
RUN microdnf update -y && rm -rf /var/cache/yum && microdnf install ca-certificates httpd-tools

# Create a non-root user to run the server as
RUN set -x ; \
    adduser registry -u 1001 -G root && exit 0

COPY --from=registry --chown=registry:0 /bin/registry /bin/registry
USER 1001
EXPOSE 5000
ENTRYPOINT ["registry"]
CMD ["serve", "/etc/docker/registry/config.yml"]