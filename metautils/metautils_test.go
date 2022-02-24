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

package metautils_test

import (
	"reflect"

	. "github.com/onmetal/controller-utils/metautils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ = Describe("Metautils", func() {
	Describe("ListElementType", func() {
		It("should return the element type of an object list", func() {
			t, err := ListElementType(&appsv1.DeploymentList{})
			Expect(err).NotTo(HaveOccurred())
			Expect(t).To(Equal(reflect.TypeOf(appsv1.Deployment{})))
		})

		It("should error if the list is not a valid object list", func() {
			_, err := ListElementType(&appsv1.Deployment{})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GVKForList", func() {
		It("should return the GVK for the list", func() {
			gvk, err := GVKForList(scheme.Scheme, &appsv1.DeploymentList{})
			Expect(err).NotTo(HaveOccurred())
			Expect(gvk).To(Equal(appsv1.SchemeGroupVersion.WithKind("Deployment")))
		})

		It("should error if no gvk could be obtained for the given object", func() {
			_, err := GVKForList(scheme.Scheme, &unstructured.UnstructuredList{})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ConvertAndSetList", func() {
		It("should convert the given objects for the list and insert them", func() {
			list := &corev1.ConfigMapList{}
			Expect(ConvertAndSetList(scheme.Scheme, list, []runtime.Object{
				&corev1.ConfigMap{},
				&unstructured.Unstructured{Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
				}},
			})).NotTo(HaveOccurred())
			Expect(list).To(Equal(&corev1.ConfigMapList{
				Items: []corev1.ConfigMap{{}, {}},
			}))
		})

		It("should error if the given list is not a list", func() {
			Expect(ConvertAndSetList(scheme.Scheme, &corev1.ConfigMap{}, nil)).To(HaveOccurred())
		})

		It("should error if an element could not be converted", func() {
			Expect(ConvertAndSetList(
				scheme.Scheme,
				&corev1.ConfigMap{},
				[]runtime.Object{&corev1.Secret{}},
			)).To(HaveOccurred())
		})
	})

	Describe("IsControlledBy", func() {
		It("should report true if the object is controlled by another", func() {
			By("making a controlling object")
			owner := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: corev1.NamespaceDefault,
					Name:      "owner",
					UID:       types.UID("owner-uuid"),
				},
			}

			By("making an object to be controlled")
			owned := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: corev1.NamespaceDefault,
					Name:      "owned",
					UID:       types.UID("owned-uuid"),
				},
			}

			By("setting the controller reference")
			Expect(controllerutil.SetControllerReference(owner, owned, scheme.Scheme)).To(Succeed())

			By("asserting the object reports as controlled")
			Expect(IsControlledBy(scheme.Scheme, owner, owned)).To(BeTrue(), "object should be controlled by owner, object: %#v, owner: %#v", owned, owner)
		})

		It("should report false if the object is not controlled by another", func() {
			By("making two regular objects")
			obj1 := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: corev1.NamespaceDefault,
					Name:      "obj1",
					UID:       types.UID("obj1-uuid"),
				},
			}
			obj2 := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: corev1.NamespaceDefault,
					Name:      "obj2",
					UID:       types.UID("obj2-uuid"),
				},
			}

			By("asserting the object does not report as controlled")
			Expect(IsControlledBy(scheme.Scheme, obj1, obj2)).To(BeFalse(), "object should not be controlled, obj1: %#v, obj2: %#v", obj1, obj2)
		})

		It("should error if it cannot determine the gvk of an object", func() {
			By("creating an object whose type is not registered in the default scheme")
			obj1 := &struct{ corev1.ConfigMap }{
				ConfigMap: corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: corev1.NamespaceDefault,
						Name:      "obj1",
						UID:       types.UID("obj1-uuid"),
					},
				},
			}

			By("making a controlling object")
			owner := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: corev1.NamespaceDefault,
					Name:      "owner",
					UID:       types.UID("owner-uuid"),
				},
			}

			By("making a regular object")
			owned := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: corev1.NamespaceDefault,
					Name:      "owned",
					UID:       types.UID("owned-uuid"),
				},
			}

			By("setting the controller for owned")
			Expect(controllerutil.SetControllerReference(owner, owned, scheme.Scheme)).To(Succeed())

			_, err := IsControlledBy(scheme.Scheme, obj1, owned)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("FilterOwnedBy", func() {
		It("should filter the objects via the owner", func() {
			owner := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "owner",
					UID:       types.UID("owner-uuid"),
				},
			}
			cm1 := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "n1",
				},
			}
			Expect(controllerutil.SetControllerReference(&owner, &cm1, scheme.Scheme)).To(Succeed())
			cm2 := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "n2",
				},
			}
			cm3 := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "n3",
				},
			}
			Expect(controllerutil.SetControllerReference(&owner, &cm3, scheme.Scheme)).To(Succeed())

			filtered, err := FilterControlledBy(scheme.Scheme, &owner, []client.Object{&cm1, &cm2, &cm3})
			Expect(err).NotTo(HaveOccurred())
			Expect(filtered).To(Equal([]client.Object{&cm1, &cm3}))
		})
	})

	Describe("ExtractList", func() {
		It("should extract a list's items", func() {
			cm1 := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "n1",
				},
			}
			cm2 := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "n2",
				},
			}
			cmList := &corev1.ConfigMapList{
				Items: []corev1.ConfigMap{cm1, cm2},
			}

			items, err := ExtractList(cmList)
			Expect(err).NotTo(HaveOccurred())
			Expect(items).To(Equal([]client.Object{&cm1, &cm2}))
		})
	})

	Describe("SetList", func() {
		It("should set a list's items", func() {
			cm1 := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "n1",
				},
			}
			cm2 := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "n2",
				},
			}
			cmList := &corev1.ConfigMapList{
				Items: []corev1.ConfigMap{cm1, cm2},
			}

			Expect(SetList(cmList, []client.Object{&cm1, &cm2})).To(Succeed())
			Expect(cmList.Items).To(Equal([]corev1.ConfigMap{cm1, cm2}))
		})
	})

	Describe("ExtractObjectSlice", func() {
		It("should extract an object slice", func() {
			cm1 := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "n1",
				},
			}
			cm2 := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "n2",
				},
			}
			slice := []corev1.ConfigMap{cm1, cm2}

			res, err := ExtractObjectSlice(slice)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal([]client.Object{&cm1, &cm2}))
		})

		It("should error if the value is not a slice", func() {
			_, err := ExtractObjectSlice(1)
			Expect(err).To(HaveOccurred())
		})

		It("should error if the slice elements cannot be converted to client.Object", func() {
			_, err := ExtractObjectSlice([]string{"foo", "bar"})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ExtractObjectSlicePointer", func() {
		It("should extract an object slice pointer", func() {
			cm1 := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "n1",
				},
			}
			cm2 := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "n2",
				},
			}
			slice := []corev1.ConfigMap{cm1, cm2}

			res, err := ExtractObjectSlicePointer(&slice)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal([]client.Object{&cm1, &cm2}))
		})

		It("should error if the value is a non-slice non-pointer", func() {
			_, err := ExtractObjectSlicePointer(1)
			Expect(err).To(HaveOccurred())
		})

		It("should error if the value is a non-pointer", func() {
			_, err := ExtractObjectSlicePointer([]corev1.ConfigMap{})
			Expect(err).To(HaveOccurred())
		})

		It("should error if the slice elements cannot be converted to client.Object", func() {
			slice := []string{"foo", "bar"}
			_, err := ExtractObjectSlicePointer(&slice)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("SetObjectSlice", func() {
		It("should set the slice values from the given client.Object slice", func() {
			var s []corev1.ConfigMap
			cm1 := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "n1",
				},
			}
			cm2 := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "n2",
				},
			}

			Expect(SetObjectSlice(&s, []client.Object{&cm1, &cm2})).To(Succeed())
			Expect(s).To(Equal([]corev1.ConfigMap{cm1, cm2}))
		})

		It("should error if the given value is a non-pointer slice", func() {
			Expect(SetObjectSlice([]client.Object{}, nil)).To(HaveOccurred())
		})

		It("should error if the given value is a pointer non-slice", func() {
			Expect(SetObjectSlice(&corev1.ConfigMap{}, nil)).To(HaveOccurred())
		})

		It("should error if the given value is a non-pointer non-slice", func() {
			Expect(SetObjectSlice(1, nil)).To(HaveOccurred())
		})
	})
})
