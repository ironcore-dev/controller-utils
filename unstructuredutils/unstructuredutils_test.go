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
	"path/filepath"
	"strings"

	"github.com/onmetal/controller-utils/testdata"
	. "github.com/onmetal/controller-utils/unstructuredutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = Describe("Unstructuredutils", func() {
	Describe("Read", func() {
		It("should read all objects from the YAML", func() {
			objs, err := Read(bytes.NewReader(testdata.ObjectsYAML))
			Expect(err).NotTo(HaveOccurred())
			Expect(objs).To(Equal(testdata.UnstructuredObjects()))
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
			objs, err := ReadFile("../testdata/bases/objects.yaml")
			Expect(err).NotTo(HaveOccurred())
			Expect(objs).To(Equal(testdata.UnstructuredObjects()))
		})

		It("should error if there is an error opening the file", func() {
			_, err := ReadFile("nonexistent.yaml")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ReadFiles", func() {
		It("should read all objects from the folder", func() {
			objs, err := ReadFiles("../testdata/bases/*.yaml")
			Expect(err).NotTo(HaveOccurred())
			Expect(objs).To(Equal([]unstructured.Unstructured{*testdata.UnstructuredMyConfigMap(), *testdata.UnstructuredSecret(), *testdata.UnstructuredConfigMap()}))
		})

		It("should empty result and no error if there is no folder presents", func() {
			objs, err := ReadFiles("nonexistent-folder")
			Expect(err).NotTo(HaveOccurred())
			Expect(objs).NotTo(Equal([]unstructured.Unstructured{}))
		})

		It("should result an ErrBadPattern error if pattern is wrong", func() {
			_, err := ReadFiles("nonexistent-folder[")
			Expect(err).Should(Equal(filepath.ErrBadPattern))

		})
	})

	Describe("UnstructuredSliceToObjectSliceNoCopy", func() {
		It("should return nil if the unstructureds are nil", func() {
			Expect(UnstructuredSliceToObjectSliceNoCopy(nil)).To(BeNil())
		})

		It("should transform the list of unstructureds to objects without copying", func() {
			uObjs := testdata.UnstructuredObjects()
			cObjs := UnstructuredSliceToObjectSliceNoCopy(uObjs)
			Expect(cObjs).To(HaveLen(len(uObjs)))
			for i := range uObjs {
				Expect(cObjs[i]).To(Equal(&uObjs[i]))
			}
		})
	})

	Describe("UnstructuredSliceToObjectSlice", func() {
		It("should return nil if the unstructureds are nil", func() {
			Expect(UnstructuredSliceToObjectSlice(nil)).To(BeNil())
		})

		It("should transform the list of unstructureds to objects without copying", func() {
			uObjs := testdata.UnstructuredObjects()
			cObjs := UnstructuredSliceToObjectSlice(uObjs)
			Expect(cObjs).To(HaveLen(len(uObjs)))
			for i := range uObjs {
				Expect(cObjs[i]).To(PointTo(Equal(uObjs[i])))
			}
		})
	})
})
