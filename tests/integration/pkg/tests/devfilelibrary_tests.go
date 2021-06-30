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
	"os"
	"path"
	"strings"

	devfilePkg "github.com/devfile/library/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/registry-support/tests/integration/pkg/config"
	"github.com/devfile/registry-support/tests/integration/pkg/util"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

const (
	nodejsStack  = "nodejs"
	quarkusStack = "java-quarkus"
	nodejsSample = "nodejs-basic"
)

var (
	userDevfileRegistry   string
	publicDevfileRegistry string
)

var _ = ginkgo.BeforeEach(func() {
	userDevfileRegistry = strings.Split(config.RegistryList, ",")[0]
	publicDevfileRegistry = strings.Split(config.RegistryList, ",")[1]
})

// Integration/e2e test logic based on https://github.com/devfile/registry-operator/tree/master/test/integration
// Tests use the CLI version of the registry-library to test.
// Note: Requires adding the CLI to the system path before running these tests
var _ = ginkgo.Describe("[Verify registry library works with registry]", func() {
	ginkgo.It("should properly list devfile stacks", func() {
		output := util.CmdShouldPass("registry-library", "list", "--type", "stack")
		gomega.Expect(output).To(gomega.ContainSubstring(nodejsStack))
		gomega.Expect(output).To(gomega.ContainSubstring(userDevfileRegistry))
		gomega.Expect(output).To(gomega.ContainSubstring(publicDevfileRegistry))
	})

	ginkgo.It("should properly list devfile samples", func() {
		output := util.CmdShouldPass("registry-library", "list", "--type", "sample")
		gomega.Expect(output).To(gomega.ContainSubstring(nodejsSample))
		gomega.Expect(output).To(gomega.ContainSubstring(userDevfileRegistry))
		gomega.Expect(output).To(gomega.ContainSubstring(publicDevfileRegistry))
	})

	ginkgo.It("should properly list both devfile stacks and samples", func() {
		output := util.CmdShouldPass("registry-library", "list", "--type", "all")
		gomega.Expect(output).To(gomega.ContainSubstring(nodejsSample))
		gomega.Expect(output).To(gomega.ContainSubstring(quarkusStack))
		gomega.Expect(output).To(gomega.ContainSubstring(userDevfileRegistry))
		gomega.Expect(output).To(gomega.ContainSubstring(publicDevfileRegistry))
	})

	ginkgo.It("should properly retrieve devfile stacks", func() {
		// Verify that the devfile library can properly pull a devfile stack from the registry
		tempDir := os.TempDir()
		util.CmdShouldPass("registry-library", "pull", publicDevfileRegistry, nodejsStack, "--context", tempDir)
		devfilePath := path.Join(tempDir, "devfile.yaml")
		_, err := os.Stat(devfilePath)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		// The devfile.yaml should be valid, verify that devfile parser doesn't throw an error
		parserArgs := parser.ParserArgs{
			Path: devfilePath,
		}
		_, _, err = devfilePkg.ParseDevfileAndValidate(parserArgs)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})
})
