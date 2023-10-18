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
	"fmt"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

// From https://github.com/openshift/odo/blob/main/tests/helper/helper_cmd_wrapper.go

type CmdWrapper struct {
	Cmd     *exec.Cmd
	program string
	args    []string
	writer  *gexec.PrefixedWriter
	session *gexec.Session
}

func Cmd(program string, args ...string) *CmdWrapper {
	prefix := fmt.Sprintf("[%s] ", filepath.Base(program))
	prefixWriter := gexec.NewPrefixedWriter(prefix, GinkgoWriter)
	command := exec.Command(program, args...)
	return &CmdWrapper{
		Cmd:     command,
		program: program,
		args:    args,
		writer:  prefixWriter,
	}
}

func (cw *CmdWrapper) ShouldPass() *CmdWrapper {
	fmt.Fprintln(GinkgoWriter, runningCmd(cw.Cmd))
	session, err := gexec.Start(cw.Cmd, cw.writer, cw.writer)
	Expect(err).NotTo(HaveOccurred())
	cw.session = session
	return cw
}

func (cw *CmdWrapper) WithEnv(args ...string) *CmdWrapper {
	cw.Cmd.Env = args
	return cw
}

func (cw *CmdWrapper) OutAndErr() (string, string) {
	return string(cw.session.Wait().Out.Contents()), string(cw.session.Wait().Err.Contents())
}

func (cw *CmdWrapper) Out() string {
	return string(cw.session.Wait().Out.Contents())
}
