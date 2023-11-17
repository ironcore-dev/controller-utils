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

package kustomizeutils

import (
	"github.com/ironcore-dev/controller-utils/testdata"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/kustomize/api/hasher"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
)

var _ = Describe("Kustomizeutils", func() {
	Describe("BuildKustomization", func() {
		It("should build the kustomization", func() {
			resMap, err := RunKustomize("../testdata")
			Expect(err).NotTo(HaveOccurred())
			Expect(resMap.Size()).To(Equal(1))
			resources := resMap.Resources()
			Expect(resources).To(HaveLen(1))
			resource := resources[0]
			data, err := resource.AsYAML()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal(testdata.ConfigMapYAML))
		})
	})

	Describe("BuildKustomizationIntoList", func() {
		It("should build the kustomization into a list", func() {
			list := &corev1.ConfigMapList{}
			Expect(RunKustomizeIntoList("../testdata", scheme.Codecs.UniversalDeserializer(), list)).To(Succeed())
			Expect(list.Items).To(ConsistOf(corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-config",
				},
				Data: map[string]string{"foo": "bar"},
			}))
		})
	})

	Describe("DecodeResource", func() {
		It("should decode the resource into the object", func() {
			res, err := resource.NewFactory(&hasher.Hasher{}).FromBytes([]byte(testdata.ConfigMapYAML))
			Expect(err).NotTo(HaveOccurred())

			cm := &corev1.ConfigMap{}
			_, _, err = DecodeResource(scheme.Codecs.UniversalDeserializer(), res, nil, cm)
			Expect(err).NotTo(HaveOccurred())
			Expect(cm).To(Equal(&corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-config",
				},
				Data: map[string]string{"foo": "bar"},
			}))
		})
	})

	Describe("DecodeResMapIntoList", func() {
		It("should decode the resmap into a list", func() {
			res, err := resource.NewFactory(&hasher.Hasher{}).FromBytes([]byte(testdata.ConfigMapYAML))
			Expect(err).NotTo(HaveOccurred())
			resMap := resmap.New()
			Expect(resMap.Append(res)).To(Succeed())

			list := &corev1.ConfigMapList{}
			Expect(DecodeResMapIntoList(scheme.Codecs.UniversalDeserializer(), resMap, list)).To(Succeed())
			Expect(list.Items).To(ConsistOf(corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-config",
				},
				Data: map[string]string{"foo": "bar"},
			}))
		})
	})

	Describe("DecodeResMapUnstructureds", func() {
		res, err := resource.NewFactory(&hasher.Hasher{}).FromBytes([]byte(testdata.ConfigMapYAML))
		Expect(err).NotTo(HaveOccurred())
		resMap := resmap.New()
		Expect(resMap.Append(res)).To(Succeed())

		unstructureds, err := DecodeResMapUnstructureds(resMap)
		Expect(err).NotTo(HaveOccurred())
		Expect(unstructureds).To(ConsistOf(unstructured.Unstructured{Object: map[string]interface{}{
			"kind":       "ConfigMap",
			"apiVersion": "v1",
			"metadata": map[string]interface{}{
				"name": "my-config",
			},
			"data": map[string]interface{}{
				"foo": "bar",
			},
		}}))
	})
})
