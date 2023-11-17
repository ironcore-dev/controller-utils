// Copyright 2021 IronCore authors
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

package envtestutils

import apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

// CRDPtrsFromCRDs generates a slice of CRD pointers from a slice of CRDs
func CRDPtrsFromCRDs(crds []apiextensionsv1.CustomResourceDefinition) (ptrs []*apiextensionsv1.CustomResourceDefinition) {
	for i := range crds {
		ptrs = append(ptrs, &crds[i])
	}
	return
}
