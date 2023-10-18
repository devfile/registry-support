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

package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"time"

	indexSchema "github.com/devfile/registry-support/index/generator/schema"
	"github.com/onsi/gomega"
)

type OCICatalog struct {
	Repositories []string `json:"repositories,omitempty"`
}

// Probes a devfile registry to check if ready
func ProbeRegistry(registryUrl string, timeout int) error {
	endpointUrl := registryUrl + "/health"
	timeoutDuration := time.Duration(timeout) * time.Second
	probe := &http.Client{}

	// Checks if registry endpoint /health is ready
	if resp, err := probe.Get(endpointUrl); resp != nil && resp.StatusCode != http.StatusOK {
		start := time.Now()

		// Set initial request timeout to the timeout duration of probing
		probe.Timeout = timeoutDuration
		// Loop until exited or timeout is reached
		for time.Since(start) < timeoutDuration {
			// Reduce request timeout by the time elapsed probing
			probe.Timeout -= time.Since(start)
			resp, err = probe.Get(endpointUrl)
			// If request errors return error
			// Else if response is OK status then health check passes and registry is ready
			if err != nil {
				return err
			} else if resp != nil && resp.StatusCode == http.StatusOK {
				return nil
			}
			time.Sleep(time.Second)
		}
		err = fmt.Errorf("probe timeout: '%s' was not ready in time", endpointUrl)
		return err
	} else {
		return err
	}
}

// GetRegistryIndex downloads the registry index.json at the specified URL and returns it
func GetRegistryIndex(url string) []indexSchema.Schema {
	resp, err := http.Get(url)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(resp.StatusCode).To(gomega.Equal(http.StatusOK))

	body, err := ioutil.ReadAll(resp.Body)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	var registryIndex []indexSchema.Schema
	err = json.Unmarshal(body, &registryIndex)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return registryIndex
}
