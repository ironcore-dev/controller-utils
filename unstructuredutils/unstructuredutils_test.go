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

package unstructuredutils_test

import (
	"bytes"
	_ "embed"
	"github.com/onmetal/controller-utils/testdata"
	. "github.com/onmetal/controller-utils/unstructuredutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"strings"
)

var _ = Describe("Unstructuredutils", func() {
	var expectedObjs []unstructured.Unstructured
	setup := func() {
		expectedObjs = []unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Secret",
					"metadata": map[string]interface{}{
						"namespace": "default",
						"name":      "my-secret",
					},
					"stringData": map[string]interface{}{
						"foo": "bar",
					},
				},
			},
			{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]interface{}{
						"namespace": "kube-system",
						"name":      "my-configmap",
					},
					"data": map[string]interface{}{
						"baz": "qux",
					},
				},
			},
		}
	}
	setup()
	BeforeEach(setup)

	Describe("Read", func() {
		It("should read all objects from the YAML", func() {
			objs, err := Read(bytes.NewReader(testdata.ObjectsYAML))
			Expect(err).NotTo(HaveOccurred())
			Expect(objs).To(Equal(expectedObjs))
		})

		It("should error on malformed yaml", func() {
			_, err := Read(strings.NewReader(`malformed: "yes`))
			Expect(err).To(HaveOccurred())
		})

		It("should error if a value cannot be converted to an object", func() {
			_, err := Read(strings.NewReader(`no: "object"`))
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ReadFile", func() {
		It("should read all objects from the file", func() {
			objs, err := ReadFile("../testdata/objects.yaml")
			Expect(err).NotTo(HaveOccurred())
			Expect(objs).To(Equal(expectedObjs))
		})

		It("should error if there is an error opening the file", func() {
			_, err := ReadFile("nonexistent.yaml")
			Expect(err).To(HaveOccurred())
		})
	})
})
