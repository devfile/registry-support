/*   Copyright 2020-2023 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tests

import (
	"io/ioutil"
	"net/http"

	devfilePkg "github.com/devfile/library/v2/pkg/devfile"
	"github.com/devfile/library/v2/pkg/devfile/parser"
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

	// v2index tests
	ginkgo.It("/v2index endpoint should return list of stacks", func() {
		registryIndex := util.GetRegistryIndex(config.Registry + "/v2index")
		for _, index := range registryIndex {
			gomega.Expect(index.Type).To(gomega.Equal(indexSchema.StackDevfileType))
		}
	})

	ginkgo.It("/v2index/sample endpoint should return list of samples", func() {
		registryIndex := util.GetRegistryIndex(config.Registry + "/v2index/sample")
		for _, index := range registryIndex {
			gomega.Expect(index.Type).To(gomega.Equal(indexSchema.SampleDevfileType))
		}
	})

	ginkgo.It("/v2index/all endpoint should return stacks and samples", func() {
		if config.IsTestRegistry {
			registryIndex := util.GetRegistryIndex(config.Registry + "/v2index/all")

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
		} else {
			ginkgo.Skip("cannot guarantee test outside of test registry, skipping test")
		}
	})

	ginkgo.It("/v2index/all?icon=base64 endpoint should return stacks and samples with encoded icon", func() {
		if config.IsTestRegistry {
			registryIndex := util.GetRegistryIndex(config.Registry + "/v2index/all?icon=base64")

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
		} else {
			ginkgo.Skip("cannot guarantee test outside of test registry, skipping test")
		}
	})

	ginkgo.It("/v2index/all?arch=amd64&arch=arm64 endpoint should return stacks and samples for arch amd64 and arm64", func() {
		if config.IsTestRegistry {
			registryIndex := util.GetRegistryIndex(config.Registry + "/v2index/all?arch=amd64&arch=arm64")

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
		} else {
			ginkgo.Skip("cannot guarantee test outside of test registry, skipping test")
		}
	})

	ginkgo.It("/v2index?arch=amd64&arch=arm64 endpoint should return stacks for devfile schema version 2.1.x", func() {
		registryIndex := util.GetRegistryIndex(config.Registry + "/v2index/all?minSchemaVersion=2.1&maxSchemaVersion=2.1")

		for _, index := range registryIndex {
			if len(index.Versions) != 0 {
				for _, version := range index.Versions {
					gomega.Expect(version.SchemaVersion).Should(gomega.HavePrefix("2.1"))
				}
			}
		}
	})

	ginkgo.It("/devfiles/<devfile> endpoint should return a devfile for stacks", func() {
		parserArgs := parser.ParserArgs{
			URL: config.Registry + "/devfiles/nodejs",
		}
		_, _, err := devfilePkg.ParseDevfileAndValidate(parserArgs)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	//workaround to avoid library parser error
	convertInURI := false
	ginkgo.It("/devfiles/<devfile> endpoint should return a devfile for samples", func() {
		parserArgs := parser.ParserArgs{
			URL:                           config.Registry + "/devfiles/code-with-quarkus",
			ConvertKubernetesContentInUri: &convertInURI,
		}
		_, _, err := devfilePkg.ParseDevfileAndValidate(parserArgs)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	ginkgo.It("/devfiles/<devfile> endpoint should return an error for a devfile that doesn't exist", func() {
		resp, err := http.Get(config.Registry + "/devfiles/fake-stack")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(resp.StatusCode).To(gomega.Equal(http.StatusNotFound))
	})

	ginkgo.It("/devfiles/<devfile>/<version> endpoint should return a devfile for stacks", func() {
		parserArgs := parser.ParserArgs{
			URL: config.Registry + "/devfiles/nodejs/latest",
		}
		_, _, err := devfilePkg.ParseDevfileAndValidate(parserArgs)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	ginkgo.It("/devfiles/<devfile>/<version> endpoint should return a devfile for samples", func() {
		if config.IsTestRegistry {
			parserArgs := parser.ParserArgs{
				URL:                           config.Registry + "/devfiles/code-with-quarkus/latest",
				ConvertKubernetesContentInUri: &convertInURI,
			}
			_, _, err := devfilePkg.ParseDevfileAndValidate(parserArgs)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		} else {
			ginkgo.Skip("cannot guarantee test outside of test registry, skipping test")
		}
	})

	ginkgo.It("/devfiles/<devfile>/starter-projects/<starterProject> endpoint should return a zip archive for devfile starter project", func() {
		resp, err := http.Get(config.Registry + "/devfiles/java-maven/starter-projects/springbootproject")
		var bytes []byte

		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		defer resp.Body.Close()

		bytes, err = ioutil.ReadAll(resp.Body)

		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(resp.StatusCode).To(gomega.Equal(http.StatusAccepted))
		gomega.Expect(bytes).ToNot(gomega.BeEmpty())
		gomega.Expect(bytes).To(gomega.Satisfy(func(file []byte) bool {
			return http.DetectContentType(file) == starterProjectMediaType
		}))
	})

	ginkgo.It("/devfiles/<devfile>/<version>/starter-projects/<starterProject> endpoint should return a zip archive for devfile starter project", func() {
		resp, err := http.Get(config.Registry + "/devfiles/java-maven/latest/starter-projects/springbootproject")
		var bytes []byte

		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		defer resp.Body.Close()

		bytes, err = ioutil.ReadAll(resp.Body)

		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(resp.StatusCode).To(gomega.Equal(http.StatusAccepted))
		gomega.Expect(bytes).ToNot(gomega.BeEmpty())
		gomega.Expect(bytes).To(gomega.Satisfy(func(file []byte) bool {
			return http.DetectContentType(file) == starterProjectMediaType
		}))
	})

	ginkgo.It("/devfiles/<devfile>/starter-projects/<starterProject> endpoint should return an offline zip archive for devfile starter project", func() {
		if config.IsTestRegistry {
			resp, err := http.Get(config.Registry + "/devfiles/go/starter-projects/go-starter-offline")
			var bytes []byte

			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			defer resp.Body.Close()

			bytes, err = ioutil.ReadAll(resp.Body)

			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(resp.StatusCode).To(gomega.Equal(http.StatusAccepted))
			gomega.Expect(bytes).ToNot(gomega.BeEmpty())
			gomega.Expect(bytes).To(gomega.Satisfy(func(file []byte) bool {
				return http.DetectContentType(file) == starterProjectMediaType
			}))
		} else {
			ginkgo.Skip("cannot guarantee test outside of test registry, skipping test")
		}
	})

	ginkgo.It("/devfiles/<devfile>/<version>/starter-projects/<starterProject> endpoint should return an offline zip archive for devfile starter project", func() {
		if config.IsTestRegistry {

			resp, err := http.Get(config.Registry + "/devfiles/go/1.2.0/starter-projects/go-starter-offline")
			var bytes []byte

			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			defer resp.Body.Close()

			bytes, err = ioutil.ReadAll(resp.Body)

			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(resp.StatusCode).To(gomega.Equal(http.StatusAccepted))
			gomega.Expect(bytes).ToNot(gomega.BeEmpty())
			gomega.Expect(bytes).To(gomega.Satisfy(func(file []byte) bool {
				return http.DetectContentType(file) == starterProjectMediaType
			}))
		} else {
			ginkgo.Skip("cannot guarantee test outside of test registry, skipping test")
		}
	})

	ginkgo.It("/devfiles/<devfile>/starter-projects/<starterProject> endpoint should return an offline zip archive of a subdir for devfile starter project", func() {
		if config.IsTestRegistry {

			resp, err := http.Get(config.Registry + "/devfiles/java-quarkus/starter-projects/community-offline")
			var bytes []byte

			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			defer resp.Body.Close()

			bytes, err = ioutil.ReadAll(resp.Body)

			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(resp.StatusCode).To(gomega.Equal(http.StatusAccepted))
			gomega.Expect(bytes).ToNot(gomega.BeEmpty())
			gomega.Expect(bytes).To(gomega.Satisfy(func(file []byte) bool {
				return http.DetectContentType(file) == starterProjectMediaType
			}))
		} else {
			ginkgo.Skip("cannot guarantee test outside of test registry, skipping test")
		}
	})

	ginkgo.It("/devfiles/<devfile>/starter-projects/<starterProject> endpoint should return an error for an offline starter project file location that doesn't exist", func() {
		if config.IsTestRegistry {
			resp, err := http.Get(config.Registry + "/devfiles/java-maven/starter-projects/springbootproject-offline")

			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(resp.StatusCode).To(gomega.Equal(http.StatusInternalServerError))
		} else {
			ginkgo.Skip("cannot guarantee test outside of test registry, skipping test")
		}
	})

	ginkgo.It("/devfiles/<devfile>/starter-projects/<starterProject> endpoint should return an error for a devfile that doesn't exist", func() {
		resp, err := http.Get(config.Registry + "/devfiles/fake-stack/starter-projects/springbootproject")

		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(resp.StatusCode).To(gomega.Equal(http.StatusNotFound))

	})

	ginkgo.It("/devfiles/<devfile>/starter-projects/<starterProject> endpoint should return an error for a starter project that doesn't exist", func() {

		resp, err := http.Get(config.Registry + "/devfiles/java-maven/starter-projects/fake-project")

		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(resp.StatusCode).To(gomega.Equal(http.StatusNotFound))
	})
})
