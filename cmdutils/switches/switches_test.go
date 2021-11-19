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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CMD Switches", func() {
	Context("Setting switches values", func() {
		It("should disable runner", func() {
			s := New([]string{"runner-a", Disable("runner-b")})
			Expect(s.Enabled("runner-a")).To(BeTrue())
			Expect(s.Enabled("runner-b")).To(BeFalse())
		})
		It("should reuse default settings", func() {
			s := New([]string{"runner-a", Disable("runner-b")})
			defaults := make(map[string]bool, len(s.settings))
			for k, v := range s.settings {
				defaults[k] = v
			}

			By("updating settings with empty value")
			Expect(s.Set("")).NotTo(HaveOccurred())
			Expect(s.settings).To(Equal(defaults))

			By("updating settings with *")
			Expect(s.Set("*")).NotTo(HaveOccurred())
			Expect(s.settings).To(Equal(defaults))
		})
		It("shouldn't reuse default settings", func() {
			s := New([]string{"runner-a", Disable("runner-b")})
			defaults := make(map[string]bool, len(s.settings))
			for k, v := range s.settings {
				defaults[k] = v
			}

			By("updating settings with new values")
			Expect(s.Set("-runner-a,runner-b")).NotTo(HaveOccurred())
			Expect(s.settings).ToNot(Equal(defaults))
		})
		It("overriding existing settings", func() {
			s := New([]string{"runner-a", Disable("runner-b")})
			defaults := make(map[string]bool, len(s.settings))
			for k, v := range s.settings {
				defaults[k] = v
			}

			By("overriding settings")
			Expect(s.Set("*,runner-b")).NotTo(HaveOccurred())
			Expect(s.settings).ToNot(Equal(defaults))
			defaults["runner-b"] = true
			Expect(s.settings).To(Equal(defaults))
		})
	})
})
