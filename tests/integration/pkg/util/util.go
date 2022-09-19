//
// Copyright 2022 Red Hat, Inc.
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
	"io/ioutil"
	"net/http"

	indexSchema "github.com/devfile/registry-support/index/generator/schema"
	"github.com/onsi/gomega"
)

type OCICatalog struct {
	Repositories []string `json:"repositories,omitempty"`
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
