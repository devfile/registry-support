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
	"os"
	"path"
	"strings"

	devfilePkg "github.com/devfile/library/v2/pkg/devfile"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	"github.com/devfile/registry-support/tests/integration/pkg/config"
	"github.com/devfile/registry-support/tests/integration/pkg/util"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

const (
	goStack          = "go"
	nodejsStack      = "nodejs"
	javaMavenStack   = "java-maven"
	quarkusStack     = "java-quarkus"
	nodejsSample     = "nodejs-basic"
	quarkusSample    = "code-with-quarkus"
	pythonSample     = "python-basic"
	javaMavenStarter = "springbootproject"
	goStarter        = "go-starter"
)

var (
	userDevfileRegistry   string
	publicDevfileRegistry string
)

var _ = ginkgo.BeforeEach(func() {
	userDevfileRegistry = strings.Split(config.RegistryList, ",")[0]
	publicDevfileRegistry = strings.Split(config.RegistryList, ",")[1]
})

// Integration/e2e test logic based on https://github.com/devfile/registry-operator/tree/main/tests/integration
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

	ginkgo.It("should properly list devfile stacks for the given arch", func() {
		if config.IsTestRegistry {
			output := util.CmdShouldPass("registry-library", "list", "--type", "stack", "--arch", "amd64", "--arch", "arm64")
			gomega.Expect(output).To(gomega.ContainSubstring(nodejsStack))
			gomega.Expect(output).To(gomega.ContainSubstring(javaMavenStack))
			gomega.Expect(output).To(gomega.ContainSubstring(userDevfileRegistry))
		} else {
			ginkgo.Skip("cannot guarantee test outside of test registry, skipping test")
		}
	})

	ginkgo.It("should properly list devfile samples for the given arch", func() {
		if config.IsTestRegistry {
			output := util.CmdShouldPass("registry-library", "list", "--type", "sample", "--arch", "amd64", "--arch", "arm64")
			gomega.Expect(output).To(gomega.ContainSubstring(nodejsSample))
			gomega.Expect(output).To(gomega.ContainSubstring(quarkusSample))
			gomega.Expect(output).To(gomega.ContainSubstring(pythonSample))
			gomega.Expect(output).To(gomega.ContainSubstring(userDevfileRegistry))
		} else {
			ginkgo.Skip("cannot guarantee test outside of test registry, skipping test")
		}
	})

	ginkgo.It("should properly list all devfiles for the given arch", func() {
		if config.IsTestRegistry {
			output := util.CmdShouldPass("registry-library", "list", "--type", "all", "--arch", "amd64", "--arch", "arm64", "--arch", "s390x")
			gomega.Expect(output).To(gomega.ContainSubstring(javaMavenStack))
			gomega.Expect(output).To(gomega.ContainSubstring(quarkusSample))
			gomega.Expect(output).To(gomega.ContainSubstring(nodejsSample))
			gomega.Expect(output).To(gomega.ContainSubstring(userDevfileRegistry))
		} else {
			ginkgo.Skip("cannot guarantee test outside of test registry, skipping test")
		}
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

	ginkgo.It("should properly download stack starter project", func() {
		tempDir := path.Join(os.TempDir(), javaMavenStarter)
		util.CmdShouldPass("registry-library", "download", publicDevfileRegistry, javaMavenStack, javaMavenStarter, "--context", tempDir)
		info, err := os.Stat(tempDir)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(info.IsDir()).To(gomega.Equal(true))
	})

	ginkgo.It("should properly download V2 stack starter project", func() {
		tempDir := path.Join(os.TempDir(), goStarter)
		util.CmdShouldPass("registry-library", "download", publicDevfileRegistry, goStack, goStarter, "--context", tempDir, "--new-index-schema")
		info, err := os.Stat(tempDir)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(info.IsDir()).To(gomega.Equal(true))
	})

	ginkgo.It("should properly download stack starter project with hostname url ending with '/'", func() {
		tempDir := path.Join(os.TempDir(), javaMavenStarter)
		util.CmdShouldPass("registry-library", "download", publicDevfileRegistry+"/", javaMavenStack, javaMavenStarter, "--context", tempDir)
		info, err := os.Stat(tempDir)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(info.IsDir()).To(gomega.Equal(true))
	})

	ginkgo.It("should properly download stack starter project with context set to relative path of WD (default)", func() {
		originalDir, err := os.Getwd()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		err = os.Chdir(os.TempDir())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		util.CmdShouldPass("registry-library", "download", publicDevfileRegistry+"/", javaMavenStack, javaMavenStarter)
		info, err := os.Stat(".")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(info.IsDir()).To(gomega.Equal(true))

		err = os.Chdir(originalDir)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})
})
