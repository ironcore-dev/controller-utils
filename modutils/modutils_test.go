// Copyright 2022 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package modutils_test

import (
	"os/exec"
	"path/filepath"
	"time"

	. "github.com/onmetal/controller-utils/modutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Modutils", func() {
	Context("Executor", func() {
		var (
			executor *Executor
		)
		BeforeEach(func() {
			executor = NewExecutor(ExecutorOptions{Dir: "../testdata/testmod1"})
		})

		Describe("ListE", func() {
			It("should list all modules", func() {
				res, err := executor.ListE()
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(ConsistOf(
					HaveField("Path", "example.org/testmod1"),
					HaveField("Path", "example.org/testmod2"),
				))
			})
		})

		Describe("GetE", func() {
			It("should get the specified module", func() {
				res, err := executor.GetE("example.org/testmod2")
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(HaveField("Path", "example.org/testmod2"))
			})
		})

		Describe("DirE", func() {
			It("should get the directory of the specified module", func() {
				dir, err := executor.DirE("example.org/testmod2")
				Expect(err).NotTo(HaveOccurred())
				Expect(dir).To(BeADirectory())
				Expect(filepath.Join(dir, "go.mod")).To(BeARegularFile())
				Expect(filepath.Join(dir, "main.go")).To(BeARegularFile())
			})

			It("should get the directory of the specified module", func() {
				dir, err := executor.DirE("example.org/testmod2", "submain")
				Expect(err).NotTo(HaveOccurred())
				Expect(dir).To(BeADirectory())
				Expect(filepath.Join(dir, "main.go")).To(BeARegularFile())
			})
		})

		Describe("BuildE", func() {
			It("should build the module via another module", func() {
				dstFilename := filepath.Join(GinkgoT().TempDir(), "hello-world")
				Expect(executor.BuildE(dstFilename, "example.org/testmod2")).To(Succeed())

				session, err := gexec.Start(exec.Command(dstFilename), GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Expect(session.Wait(1 * time.Second).Out).To(gbytes.Say("Hello, World!"))
			})

			It("should build the module via another module and subpath", func() {
				dstFilename := filepath.Join(GinkgoT().TempDir(), "hello-world")
				Expect(executor.BuildE(dstFilename, "example.org/testmod2", "submain")).To(Succeed())

				session, err := gexec.Start(exec.Command(dstFilename), GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Expect(session.Wait(1 * time.Second).Out).To(gbytes.Say("Hello, Submain!"))
			})
		})
	})
})
