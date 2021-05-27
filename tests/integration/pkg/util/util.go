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
