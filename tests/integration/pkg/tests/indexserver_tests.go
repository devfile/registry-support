//
// Copyright (c) 2020 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation
//

package tests

import (
	"io/ioutil"
	"net/http"

	devfilePkg "github.com/devfile/library/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser"
	indexSchema "github.com/devfile/registry-support/index/generator/schema"

	"github.com/devfile/registry-support/tests/integration/pkg/config"
	"github.com/devfile/registry-support/tests/integration/pkg/util"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

// Integration/e2e test logic based on https://github.com/devfile/registry-operator/tree/master/test/integration

var _ = ginkgo.Describe("[Verify index server is working properly]", func() {
	ginkgo.It("Root endpoint should be available", func() {
		resp, err := http.Get(config.Registry)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(resp.StatusCode).To(gomega.Equal(http.StatusOK))
	})

	ginkgo.It("Root endpoint should redirect to /viewer if text/html was requested", func() {
		client := &http.Client{}
		req, _ := http.NewRequest("GET", config.Registry, nil)
		req.Header.Set("Accept", "text/html")
		resp, err := client.Do(req)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(resp.StatusCode).To(gomega.Equal(http.StatusOK))

		// Check the path of the response. Should have been redirected to /viewer
		gomega.Expect(resp.Request.URL.Path).To(gomega.Equal("/viewer"))

		bytes, err := ioutil.ReadAll(resp.Body)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		body := string(bytes)
		gomega.Expect(body).To(gomega.ContainSubstring("<!DOCTYPE html>"))
	})

	ginkgo.It("/viewer should serve the registry viewer", func() {
		client := &http.Client{}
		req, _ := http.NewRequest("GET", config.Registry+"/viewer", nil)
		resp, err := client.Do(req)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(resp.StatusCode).To(gomega.Equal(http.StatusOK))

		// Check the path of the response. Should have been redirected to /viewer
		gomega.Expect(resp.Request.URL.Path).To(gomega.Equal("/viewer"))

		bytes, err := ioutil.ReadAll(resp.Body)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		body := string(bytes)

		gomega.Expect(body).To(gomega.ContainSubstring("<!DOCTYPE html>"))

		// Verify that the registry viewer's page has been properly generated
		gomega.Expect(body).To(gomega.ContainSubstring("A simple Hello World Node.js application"))
	})

	ginkgo.It("/index endpoint should return list of stacks", func() {
		registryIndex := util.GetRegistryIndex(config.Registry + "/index")
		for _, index := range registryIndex {
			gomega.Expect(index.Type).To(gomega.Equal(indexSchema.StackDevfileType))
		}
	})

	ginkgo.It("/index/sample endpoint should return list of samples", func() {
		registryIndex := util.GetRegistryIndex(config.Registry + "/index/sample")
		for _, index := range registryIndex {
			gomega.Expect(index.Type).To(gomega.Equal(indexSchema.SampleDevfileType))
		}
	})

	ginkgo.It("/index/all endpoint should return stacks and samples", func() {
		registryIndex := util.GetRegistryIndex(config.Registry + "/index/all")

		hasStacks := false
		hasSamples := false
		for _, index := range registryIndex {
			if index.Type == indexSchema.SampleDevfileType {
				hasSamples = true
			}
			if index.Type == indexSchema.StackDevfileType {
				hasStacks = true
			}
		}
		gomega.Expect(hasStacks && hasSamples).To(gomega.BeTrue())
	})

	ginkgo.It("/index/all?icon=base64 endpoint should return stacks and samples with encoded icon", func() {
		registryIndex := util.GetRegistryIndex(config.Registry + "/index/all?icon=base64")

		hasStacks := false
		hasSamples := false

		for _, index := range registryIndex {
			if index.Type == indexSchema.SampleDevfileType {
				hasSamples = true
			}
			if index.Type == indexSchema.StackDevfileType {
				hasStacks = true
			}
			if index.Icon != "" {
				gomega.Expect(index.Icon).To(gomega.HavePrefix("data:image"))
			}
		}
		gomega.Expect(hasStacks && hasSamples).To(gomega.BeTrue())
	})

	ginkgo.It("/index/all?arch=amd64&arch=arm64 endpoint should return stacks and samples for arch amd64 and arm64", func() {
		registryIndex := util.GetRegistryIndex(config.Registry + "/index/all?arch=amd64&arch=arm64")

		hasStacks := false
		hasSamples := false
		for _, index := range registryIndex {
			if index.Type == indexSchema.SampleDevfileType {
				hasSamples = true
			}
			if index.Type == indexSchema.StackDevfileType {
				hasStacks = true
			}
			if len(index.Architectures) != 0 {
				gomega.Expect(index.Architectures).Should(gomega.ContainElements("amd64", "arm64"))
			}
		}

		if len(registryIndex) > 0 {
			gomega.Expect(hasStacks && hasSamples).To(gomega.BeTrue())
		}
	})

	ginkgo.It("/devfiles/<devfile> endpoint should return a devfile for stacks", func() {
		parserArgs := parser.ParserArgs{
			URL: config.Registry + "/devfiles/nodejs",
		}
		_, _, err := devfilePkg.ParseDevfileAndValidate(parserArgs)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	ginkgo.It("/devfiles/<devfile> endpoint should return a devfile for samples", func() {
		parserArgs := parser.ParserArgs{
			URL: config.Registry + "/devfiles/code-with-quarkus",
		}
		_, _, err := devfilePkg.ParseDevfileAndValidate(parserArgs)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	ginkgo.It("/devfiles/<devfile> endpoint should return an error for a devfile that doesn't exist", func() {
		resp, err := http.Get(config.Registry + "/devfiles/fake-stack")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(resp.StatusCode).To(gomega.Equal(http.StatusNotFound))
	})
})
