// Copyright 2021 OnMetal authors
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
package switches

import (
	"flag"

	"github.com/spf13/pflag"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/kustomize/kyaml/sets"
)

var _ = Describe("CMD Switches", func() {
	Context("Testing Switches interface", func() {
		It("should disable runner", func() {
			s := New("runner-a", "runner-b")
			Expect(s.Set("*,-runner-b")).ToNot(HaveOccurred())
			Expect(s.Enabled("runner-a")).To(BeTrue())
			Expect(s.Enabled("runner-b")).To(BeFalse())
		})
		It("should return all items", func() {
			s := New("runner-a", "runner-b")
			Expect(s.Set("*,-runner-b")).ToNot(HaveOccurred())

			expected := make(sets.String)
			expected.Insert("runner-a", "runner-b")
			Expect(s.All()).To(Equal(expected))
		})
		It("should return all enabled items", func() {
			s := New("runner-a", Disable("runner-b"))

			expected := make(sets.String)
			expected.Insert("runner-a")
			Expect(s.EnabledByDefault()).To(Equal(expected))
		})
		It("should return all disabled items", func() {
			s := New("runner-a", Disable("runner-b"))

			expected := make(sets.String)
			expected.Insert("runner-b")
			Expect(s.DisabledByDefault()).To(Equal(expected))
		})
		It("should return string", func() {
			s := New("runner-a", "runner-b")
			Expect(s.Set("*,-runner-b")).ToNot(HaveOccurred())
			Expect(s.String()).To(Equal("runner-a,-runner-b"))
		})
	})

	Context("Testing flag package behavior", func() {
		It("should disable all controllers when no flag is passed", func() {
			fs := flag.NewFlagSet("", flag.ExitOnError)
			controllers := New("runner-a", Disable("runner-b"), "runner-c")
			fs.Var(controllers, "controllers", "")

			Expect(fs.Parse([]string{})).NotTo(HaveOccurred())
			Expect(controllers.Enabled("runner-a")).To(BeFalse())
			Expect(controllers.Enabled("runner-b")).To(BeFalse())
			Expect(controllers.Enabled("runner-c")).To(BeFalse())
		})
		It("should keep default settings when * is passed", func() {
			fs := flag.NewFlagSet("", flag.ExitOnError)
			controllers := New("runner-a", Disable("runner-b"), "runner-c")
			fs.Var(controllers, "controllers", "")

			Expect(fs.Parse([]string{"--controllers=*"})).NotTo(HaveOccurred())
			Expect(controllers.Enabled("runner-a")).To(BeTrue())
			Expect(controllers.Enabled("runner-b")).To(BeFalse())
			Expect(controllers.Enabled("runner-c")).To(BeTrue())
		})
		It("should override default settings", func() {
			fs := flag.NewFlagSet("", flag.ExitOnError)
			controllers := New("runner-a", Disable("runner-b"), "runner-c")
			fs.Var(controllers, "controllers", "")

			Expect(fs.Parse([]string{"--controllers=runner-a,-runner-c"})).NotTo(HaveOccurred())
			Expect(controllers.Enabled("runner-a")).To(BeTrue())
			Expect(controllers.Enabled("runner-b")).To(BeFalse())
			Expect(controllers.Enabled("runner-c")).To(BeFalse())
		})
		It("should override some of default settings", func() {
			fs := flag.NewFlagSet("", flag.ExitOnError)
			controllers := New("runner-a", Disable("runner-b"), "runner-c")
			fs.Var(controllers, "controllers", "")

			Expect(fs.Parse([]string{"--controllers=*,-runner-a"})).NotTo(HaveOccurred())
			Expect(controllers.Enabled("runner-a")).To(BeFalse())
			Expect(controllers.Enabled("runner-b")).To(BeFalse())
			Expect(controllers.Enabled("runner-c")).To(BeTrue())
		})
	})

	Context("Testing pflag package behavior", func() {
		It("should disable all controllers when no flag is passed", func() {
			fs := pflag.NewFlagSet("", pflag.ExitOnError)
			controllers := New("runner-a", Disable("runner-b"), "runner-c")
			fs.Var(controllers, "controllers", "")

			Expect(fs.Parse([]string{})).NotTo(HaveOccurred())
			Expect(controllers.Enabled("runner-a")).To(BeFalse())
			Expect(controllers.Enabled("runner-b")).To(BeFalse())
			Expect(controllers.Enabled("runner-c")).To(BeFalse())
		})
		It("should keep default settings when * is passed", func() {
			fs := pflag.NewFlagSet("", pflag.ExitOnError)
			controllers := New("runner-a", Disable("runner-b"), "runner-c")
			fs.Var(controllers, "controllers", "")

			Expect(fs.Parse([]string{"--controllers=*"})).NotTo(HaveOccurred())
			Expect(controllers.Enabled("runner-a")).To(BeTrue())
			Expect(controllers.Enabled("runner-b")).To(BeFalse())
			Expect(controllers.Enabled("runner-c")).To(BeTrue())
		})
		It("should override default settings", func() {
			fs := pflag.NewFlagSet("", pflag.ExitOnError)
			controllers := New("runner-a", Disable("runner-b"), "runner-c")
			fs.Var(controllers, "controllers", "")

			Expect(fs.Parse([]string{"--controllers=runner-a,-runner-c"})).NotTo(HaveOccurred())
			Expect(controllers.Enabled("runner-a")).To(BeTrue())
			Expect(controllers.Enabled("runner-b")).To(BeFalse())
			Expect(controllers.Enabled("runner-c")).To(BeFalse())
		})
		It("should override some of default settings", func() {
			fs := pflag.NewFlagSet("", pflag.ExitOnError)
			controllers := New("runner-a", Disable("runner-b"), "runner-c")
			fs.Var(controllers, "controllers", "")

			Expect(fs.Parse([]string{"--controllers=*,-runner-a"})).NotTo(HaveOccurred())
			Expect(controllers.Enabled("runner-a")).To(BeFalse())
			Expect(controllers.Enabled("runner-b")).To(BeFalse())
			Expect(controllers.Enabled("runner-c")).To(BeTrue())
		})
	})
})
