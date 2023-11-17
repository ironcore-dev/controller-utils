// Copyright 2022 IronCore authors
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

package buildutils_test

import (
	. "github.com/ironcore-dev/controller-utils/buildutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Buildutils", func() {
	Describe("ModMode", func() {
		It("should set the mod property of build options", func() {
			mod := ModModeMod
			o := &BuildOptions{}
			mod.ApplyToBuild(o)

			Expect(o.Mod).To(HaveValue(Equal(mod)))
		})
	})

	Describe("ForceRebuild", func() {
		It("should set the force rebuild property of build options", func() {
			o := &BuildOptions{}
			ForceRebuild.ApplyToBuild(o)

			Expect(o.ForceRebuild).To(BeTrue())
		})
	})
})
