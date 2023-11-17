// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package envtestutils

import apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

// CRDPtrsFromCRDs generates a slice of CRD pointers from a slice of CRDs
func CRDPtrsFromCRDs(crds []apiextensionsv1.CustomResourceDefinition) (ptrs []*apiextensionsv1.CustomResourceDefinition) {
	for i := range crds {
		ptrs = append(ptrs, &crds[i])
	}
	return
}
