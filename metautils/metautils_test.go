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
	"strings"

	"github.com/golang/mock/gomock"
	. "github.com/onmetal/controller-utils/metautils"
	mockmetautils "github.com/onmetal/controller-utils/mock/controller-utils/metautils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type BadList struct {
	metav1.ListMeta
	metav1.TypeMeta
}

func (b *BadList) DeepCopyObject() runtime.Object {
	return &BadList{
		*b.ListMeta.DeepCopy(),
		b.TypeMeta,
	}
}

var _ = Describe("Metautils", func() {
	var (
		ctrl *gomock.Controller
	)
	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		DeferCleanup(ctrl.Finish)
	})

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

		It("should error if it cannot extract a list's items", func() {
			_, err := ExtractList(&BadList{})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("MustExtractList", func() {
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

			items := MustExtractList(cmList)
			Expect(items).To(Equal([]client.Object{&cm1, &cm2}))
		})

		It("should panic if it cannot extract a list's items", func() {
			Expect(func() { MustExtractList(&BadList{}) }).To(Panic())
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

		It("should error if it cannot set a list's items", func() {
			Expect(SetList(&BadList{}, nil)).To(HaveOccurred())
		})
	})

	Describe("MustSetList", func() {
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

			MustSetList(cmList, []client.Object{&cm1, &cm2})
			Expect(cmList.Items).To(Equal([]corev1.ConfigMap{cm1, cm2}))
		})

		It("should panic if it cannot set a list's items", func() {
			Expect(func() { MustSetList(&BadList{}, nil) }).To(Panic())
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

	Describe("MustExtractObjectSlice", func() {
		It("should panic if it cannot extract the given object slice", func() {
			Expect(func() { MustExtractObjectSlice([]string{"foo", "bar"}) }).To(Panic())
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

	Describe("MustExtractObjectSlicePointer", func() {
		It("should panic if it cannot extract the object slice", func() {
			slice := []string{"foo", "bar"}
			Expect(func() { MustExtractObjectSlicePointer(&slice) }).To(Panic())
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

	Describe("MustSetObjectSlice", func() {
		It("should panic if it cannot set the object slice", func() {
			Expect(func() { MustSetObjectSlice(1, nil) }).To(Panic())
		})
	})

	Describe("NewListForGVK", func() {
		It("should create a new list for the given gvk", func() {
			list, err := NewListForGVK(scheme.Scheme, corev1.SchemeGroupVersion.WithKind("Secret"))
			Expect(err).NotTo(HaveOccurred())

			Expect(list).To(Equal(&corev1.SecretList{}))
		})

		It("should error if it cannot instantiate the list", func() {
			_, err := NewListForGVK(scheme.Scheme, schema.GroupVersionKind{})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("NewListForObject", func() {
		It("should create a new list for the given object", func() {
			list, err := NewListForObject(scheme.Scheme, &corev1.Secret{})
			Expect(err).NotTo(HaveOccurred())

			Expect(list).To(Equal(&corev1.SecretList{}))
		})

		It("should error if it cannot determine the gvk for the object", func() {
			_, err := NewListForObject(scheme.Scheme, &unstructured.Unstructured{})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("EachListItem", func() {
		It("should traverse over each list item", func() {
			list := &corev1.SecretList{
				Items: []corev1.Secret{
					{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "bar"}},
				},
			}

			f := mockmetautils.NewMockEachListItemFunc(ctrl)
			gomock.InOrder(
				f.EXPECT().Call(&list.Items[0]),
				f.EXPECT().Call(&list.Items[1]),
			)

			Expect(EachListItem(list, f.Call)).To(Succeed())
		})
	})

	DescribeTable("HasLabel",
		func(initLabels map[string]string, key string, expected bool) {
			obj := &metav1.ObjectMeta{Labels: initLabels}
			Expect(HasLabel(obj, key)).To(Equal(expected))
		},
		Entry("label present", map[string]string{"foo": ""}, "foo", true),
		Entry("nil labels", nil, "foo", false),
		Entry("different label present", map[string]string{"bar": ""}, "foo", false),
	)

	DescribeTable("SetLabel",
		func(initLabels map[string]string, key, value string, expected map[string]string) {
			obj := &metav1.ObjectMeta{Labels: initLabels}
			SetLabel(obj, key, value)
			Expect(obj.Labels).To(Equal(expected))
		},
		Entry("nil labels", nil, "foo", "bar", map[string]string{"foo": "bar"}),
		Entry("key present w/ different value", map[string]string{"foo": "baz"}, "foo", "bar", map[string]string{"foo": "bar"}),
		Entry("other keys present", map[string]string{"bar": "baz"}, "foo", "bar", map[string]string{"bar": "baz", "foo": "bar"}),
	)

	DescribeTable("SetLabels",
		func(initLabels map[string]string, set, expected map[string]string) {
			obj := &metav1.ObjectMeta{Labels: initLabels}
			SetLabels(obj, set)
			Expect(obj.Labels).To(Equal(expected))
		},
		Entry("nil labels", nil, map[string]string{"foo": "bar"}, map[string]string{"foo": "bar"}),
		Entry("key present w/ different value", map[string]string{"foo": "baz"}, map[string]string{"foo": "bar"}, map[string]string{"foo": "bar"}),
		Entry("other keys present", map[string]string{"bar": "baz"}, map[string]string{"foo": "bar"}, map[string]string{"bar": "baz", "foo": "bar"}),
		Entry("partial other keys, same key", map[string]string{"foo": "baz", "bar": "baz"}, map[string]string{"foo": "bar"}, map[string]string{"foo": "bar", "bar": "baz"}),
	)

	DescribeTable("DeleteLabel",
		func(initLabels map[string]string, key string, expected map[string]string) {
			obj := &metav1.ObjectMeta{Labels: initLabels}
			DeleteLabel(obj, key)
			Expect(obj.Labels).To(Equal(expected))
		},
		Entry("key present", map[string]string{"foo": "bar"}, "foo", map[string]string{}),
		Entry("different key present", map[string]string{"bar": "baz"}, "foo", map[string]string{"bar": "baz"}),
		Entry("nil map", nil, "foo", nil),
	)

	DescribeTable("DeleteLabels",
		func(initLabels map[string]string, keys []string, expected map[string]string) {
			obj := &metav1.ObjectMeta{Labels: initLabels}
			DeleteLabels(obj, keys)
			Expect(obj.Labels).To(Equal(expected))
		},
		Entry("keys present", map[string]string{"foo": "bar", "bar": "baz"}, []string{"foo", "bar"}, map[string]string{}),
		Entry("some keys present", map[string]string{"foo": "bar", "bar": "baz"}, []string{"foo"}, map[string]string{"bar": "baz"}),
		Entry("no keys present", map[string]string{"foo": "bar", "bar": "baz"}, []string{"qux"}, map[string]string{"foo": "bar", "bar": "baz"}),
		Entry("nil map", nil, []string{"foo", "bar"}, nil),
	)

	DescribeTable("HasAnnotation",
		func(initAnnotations map[string]string, key string, expected bool) {
			obj := &metav1.ObjectMeta{Annotations: initAnnotations}
			Expect(HasAnnotation(obj, key)).To(Equal(expected))
		},
		Entry("annotation present", map[string]string{"foo": ""}, "foo", true),
		Entry("nil annotations", nil, "foo", false),
		Entry("different annotation present", map[string]string{"bar": ""}, "foo", false),
	)

	DescribeTable("SetAnnotation",
		func(initAnnotations map[string]string, key, value string, expected map[string]string) {
			obj := &metav1.ObjectMeta{Annotations: initAnnotations}
			SetAnnotation(obj, key, value)
			Expect(obj.Annotations).To(Equal(expected))
		},
		Entry("nil annotations", nil, "foo", "bar", map[string]string{"foo": "bar"}),
		Entry("key present w/ different value", map[string]string{"foo": "baz"}, "foo", "bar", map[string]string{"foo": "bar"}),
		Entry("other keys present", map[string]string{"bar": "baz"}, "foo", "bar", map[string]string{"bar": "baz", "foo": "bar"}),
	)

	DescribeTable("SetAnnotations",
		func(initAnnotations map[string]string, set, expected map[string]string) {
			obj := &metav1.ObjectMeta{Annotations: initAnnotations}
			SetAnnotations(obj, set)
			Expect(obj.Annotations).To(Equal(expected))
		},
		Entry("nil annotations", nil, map[string]string{"foo": "bar"}, map[string]string{"foo": "bar"}),
		Entry("key present w/ different value", map[string]string{"foo": "baz"}, map[string]string{"foo": "bar"}, map[string]string{"foo": "bar"}),
		Entry("other keys present", map[string]string{"bar": "baz"}, map[string]string{"foo": "bar"}, map[string]string{"bar": "baz", "foo": "bar"}),
		Entry("partial other keys, same key", map[string]string{"foo": "baz", "bar": "baz"}, map[string]string{"foo": "bar"}, map[string]string{"foo": "bar", "bar": "baz"}),
	)

	DescribeTable("DeleteAnnotation",
		func(initAnnotations map[string]string, key string, expected map[string]string) {
			obj := &metav1.ObjectMeta{Annotations: initAnnotations}
			DeleteAnnotation(obj, key)
			Expect(obj.Annotations).To(Equal(expected))
		},
		Entry("key present", map[string]string{"foo": "bar"}, "foo", map[string]string{}),
		Entry("different key present", map[string]string{"bar": "baz"}, "foo", map[string]string{"bar": "baz"}),
		Entry("nil map", nil, "foo", nil),
	)

	DescribeTable("DeleteAnnotations",
		func(initAnnotations map[string]string, keys []string, expected map[string]string) {
			obj := &metav1.ObjectMeta{Annotations: initAnnotations}
			DeleteAnnotations(obj, keys)
			Expect(obj.Annotations).To(Equal(expected))
		},
		Entry("keys present", map[string]string{"foo": "bar", "bar": "baz"}, []string{"foo", "bar"}, map[string]string{}),
		Entry("some keys present", map[string]string{"foo": "bar", "bar": "baz"}, []string{"foo"}, map[string]string{"bar": "baz"}),
		Entry("no keys present", map[string]string{"foo": "bar", "bar": "baz"}, []string{"qux"}, map[string]string{"foo": "bar", "bar": "baz"}),
		Entry("nil map", nil, []string{"foo", "bar"}, nil),
	)

	Describe("FilterList", func() {
		It("should filter the list with the given function", func() {
			list := &corev1.SecretList{
				Items: []corev1.Secret{
					{ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"}},
					{ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "baz"}},
					{ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "qux"}},
					{ObjectMeta: metav1.ObjectMeta{Namespace: "bar", Name: "bar"}},
				},
			}

			Expect(FilterList(list, func(obj client.Object) bool {
				return obj.GetNamespace() == "foo" && strings.HasPrefix(obj.GetName(), "b")
			})).To(Succeed())

			Expect(list).To(Equal(&corev1.SecretList{
				Items: []corev1.Secret{
					{ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"}},
					{ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "baz"}},
				},
			}))
		})
	})
})
