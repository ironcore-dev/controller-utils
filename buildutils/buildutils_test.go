// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

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
