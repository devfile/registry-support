//
// Copyright Red Hat
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tests

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/devfile/registry-support/tests/integration/pkg/config"
	"github.com/devfile/registry-support/tests/integration/pkg/util"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

// Integration/e2e test logic based on https://github.com/devfile/registry-operator/tree/master/test/integration

var _ = ginkgo.Describe("[Verify oci registry is working properly]", func() {
	ginkgo.It("/v2 endpoint should be available", func() {
		resp, err := http.Get(config.Registry + "/v2/")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(resp.StatusCode).To(gomega.Equal(http.StatusOK))
	})

	ginkgo.It("/v2/_catalog endpoint should return a list of OCI artifacts", func() {
		resp, err := http.Get(config.Registry + "/v2/_catalog")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(resp.StatusCode).To(gomega.Equal(http.StatusOK))

		body, err := ioutil.ReadAll(resp.Body)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		// Parse the oci catalog entries and verify it has at least 1 entry in it
		var ociCatalog util.OCICatalog
		err = json.Unmarshal(body, &ociCatalog)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(len(ociCatalog.Repositories) > 0).To(gomega.BeTrue())
	})

	ginkgo.It("POST requests should be denied", func() {
		responseBody := bytes.NewBuffer(nil)

		// OCI server proxy should return a 404 not found error for non-GET requests
		resp, err := http.Post(config.Registry+"/v2", "application/text", responseBody)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(resp.StatusCode).To(gomega.Equal(http.StatusNotFound))
	})
})
