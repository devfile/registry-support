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
<<<<<<< HEAD
FROM golang:1.19-alpine@sha256:0ec0646e208ea58e5d29e558e39f2e59fccf39b7bda306cb53bbaff91919eca5
=======
FROM golang:1.21-alpine@sha256:3f8e3ad3e7c128d29ac3004ac8314967c5ddbfa5bfa7caa59b0de493fc01686a
>>>>>>> 9b9d99c6 (Bump up integration generator server and ci go version)

WORKDIR /registry-test

COPY . .

# Build the registry library and add to path
RUN cd /registry-test/registry-library && ./build.sh && cp /registry-test/registry-library/registry-library /usr/local/bin/registry-library

# Build the test binary
RUN /registry-test/build.sh

CMD /registry-test/devfile-registry-integration