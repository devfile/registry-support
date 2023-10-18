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

package cmd

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/devfile/registry-support/tests/integration/pkg/config"
	"github.com/devfile/registry-support/tests/integration/pkg/util"

	_ "github.com/devfile/registry-support/tests/integration/pkg/tests"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

// Integration/e2e test logic based on https://github.com/devfile/devworkspace-operator/tree/main/test/e2e

// Create Constant file
const (
	testResultsDirectory = "/tmp/artifacts"
	jUnitOutputFilename  = "junit-devfileregistry-operator.xml"
)

// SynchronizedBeforeSuite blocks is executed before run all test suites
var _ = ginkgo.SynchronizedBeforeSuite(func() []byte {
	var err error

	fmt.Println("Starting to setup objects before run ginkgo suite")
	registry := os.Getenv("REGISTRY")
	if registry == "" {
		registry = "https://registry.devfile.io"
	}
	config.Registry = registry
	config.RegistryList = registry + "," + "https://registry.stage.devfile.io"
	err = os.Setenv("REGISTRY_LIST", config.RegistryList)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	// If the registry is the test one, set IsTestRegistry variable to run extra test cases
	if isTestRegistry, isSet := os.LookupEnv("IS_TEST_REGISTRY"); isSet {
		config.IsTestRegistry, err = strconv.ParseBool(isTestRegistry)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
	}

	// If timeout duration until readiness probe runs out is set, run readiness probe
	if timeout, isSet := os.LookupEnv("PROBE_TIMEOUT"); isSet {
		probeTimeout, err := strconv.Atoi(timeout)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}

		if err = util.ProbeRegistry(config.Registry, probeTimeout); err != nil {
			fmt.Println(err)
			panic(err)
		}
	}

	return nil
}, func(data []byte) {})

func TestDevfileRegistryController(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)

	fmt.Println("Running Devfile Registry integration tests...")
	ginkgo.RunSpecs(t, "Devfile Registry Tests")
}
