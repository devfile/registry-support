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

	ginkgo.It("/devfiles/<devfile> endpoint should return a devfile", func() {
		parserArgs := parser.ParserArgs{
			URL: config.Registry + "/devfiles/nodejs",
		}
		_, _, err := devfilePkg.ParseDevfileAndValidate(parserArgs)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})
})
