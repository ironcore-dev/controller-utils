// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package envtestutils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

var _ = Describe("CRDPtrsFromCRDs", func() {
	It("returns CRD pointers from CRDs", func() {
		crdA := apiextensionsv1.CustomResourceDefinition{}
		crdB := apiextensionsv1.CustomResourceDefinition{}

		crds := []apiextensionsv1.CustomResourceDefinition{crdA, crdB}

		Expect(CRDPtrsFromCRDs(crds)).To(Equal([]*apiextensionsv1.CustomResourceDefinition{&crdA, &crdB}))
	})
})
