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

package memorystore_test

import (
	"context"

	"github.com/onmetal/controller-utils/memorystore"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Store", func() {
	var (
		s         *memorystore.Store
		ctx       context.Context
		namespace string

		pod    *corev1.Pod
		podKey client.ObjectKey
		podGR  schema.GroupResource
		podGK  schema.GroupKind

		cm1    *corev1.ConfigMap
		cm1Key client.ObjectKey
		cm2    *corev1.ConfigMap
		cm2Key client.ObjectKey
		cmGR   schema.GroupResource
		cmGK   schema.GroupKind
	)
	f := func() {
		s = memorystore.New(scheme.Scheme)
		ctx = context.Background()
		namespace = corev1.NamespaceDefault

		pod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: "my-pod"},
		}
		podKey = client.ObjectKeyFromObject(pod)
		podGR = schema.GroupResource{Group: "", Resource: "Pod"}
		podGK = schema.GroupKind{Group: "", Kind: "Pod"}

		cm1 = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      "my-cm",
				Labels: map[string]string{
					"foo": "bar",
				},
			},
		}
		cm1Key = client.ObjectKeyFromObject(cm1)
		cm2 = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "other-namespace",
				Name:      "other-cm",
			},
		}
		cmGR = schema.GroupResource{Group: "", Resource: "ConfigMap"}
		cmGK = schema.GroupKind{Group: "", Kind: "ConfigMap"}

		_ = cm2Key
		_ = podKey
		_ = podGR
	}
	f()
	BeforeEach(f)

	Describe("Objects", func() {
		It("should return the stored objects", func() {
			Expect(s.Objects()).To(BeEmpty())

			Expect(s.Create(ctx, cm1)).To(Succeed())
			Expect(s.Create(ctx, pod)).To(Succeed())

			Expect(s.Objects()).To(ConsistOf(cm1, pod))
		})
	})

	Describe("GroupKindObjects", func() {
		It("should return the stored objects for the given group kind", func() {
			Expect(s.GroupKindObjects(podGK)).To(BeEmpty())
			Expect(s.GroupKindObjects(cmGK)).To(BeEmpty())

			Expect(s.Create(ctx, cm1)).To(Succeed())
			Expect(s.Create(ctx, pod)).To(Succeed())

			Expect(s.GroupKindObjects(podGK)).To(ConsistOf(pod))
			Expect(s.GroupKindObjects(cmGK)).To(ConsistOf(cm1))
		})
	})

	Describe("GroupKinds", func() {
		It("should return the stored group kinds", func() {
			Expect(s.GroupKinds()).To(BeEmpty())

			Expect(s.Create(ctx, cm1)).To(Succeed())
			Expect(s.Create(ctx, pod)).To(Succeed())

			Expect(s.GroupKinds()).To(ConsistOf(podGK, cmGK))
		})
	})

	Describe("Create", func() {
		It("should create and retrieve the given object", func() {
			Expect(s.Create(ctx, cm1)).To(Succeed())
			Expect(s.Objects()).To(ConsistOf(cm1))
		})

		It("should return already exists if the object already exists", func() {
			Expect(s.Create(ctx, cm1)).To(Succeed())
			Expect(s.Create(ctx, cm1)).To(MatchError(apierrors.NewAlreadyExists(cmGR, cm1Key.String())))
		})

		DescribeTable("unsupported create options",
			func(opts ...client.CreateOption) {
				Expect(s.Create(ctx, cm1, opts...)).To(HaveOccurred())
			},
			Entry("dry run", client.DryRunAll),
			Entry("raw", &client.CreateOptions{Raw: &metav1.CreateOptions{}}),
		)
	})

	Describe("Get", func() {
		It("should get the specified object", func() {
			Expect(s.Create(ctx, cm1)).To(Succeed())

			otherCM := &corev1.ConfigMap{}
			Expect(s.Get(ctx, cm1Key, otherCM)).To(Succeed())
			Expect(otherCM).To(Equal(cm1))
		})

		It("should error if the specified object does not exist", func() {
			Expect(s.Get(ctx, cm1Key, &corev1.ConfigMap{})).To(Equal(apierrors.NewNotFound(cmGR, cm1Key.String())))
		})
	})

	Describe("List", func() {
		DescribeTable("successful listing",
			func(opts []client.ListOption, elems ...interface{}) {
				Expect(s.Create(ctx, cm1)).To(Succeed())
				Expect(s.Create(ctx, cm2)).To(Succeed())
				list := &corev1.ConfigMapList{}
				Expect(s.List(ctx, list, opts...)).To(Succeed())
				Expect(list.Items).To(ConsistOf(elems...))
			},
			Entry("no options", nil, *cm1, *cm2),
			Entry("in namespace", []client.ListOption{client.InNamespace(namespace)}, *cm1),
			Entry("matching labels", []client.ListOption{client.MatchingLabels{"foo": "bar"}}, *cm1),
		)

		DescribeTable("unsupported options",
			func(opts ...client.ListOption) {
				Expect(s.List(ctx, &corev1.ConfigMapList{}, opts...)).To(HaveOccurred())
			},
			Entry("raw", &client.ListOptions{Raw: &metav1.ListOptions{}}),
			Entry("continue", client.Continue("foo")),
			Entry("limit", client.Limit(1)),
			Entry("field selector", client.MatchingFields{"foo": "bar"}),
		)
	})

	Describe("Status", func() {
		It("should return the store itself", func() {
			Expect(s.Status()).Should(Equal(s))
		})
	})

	Describe("Delete", func() {
		It("should delete the specified object", func() {
			Expect(s.Create(ctx, cm1)).To(Succeed())
			Expect(s.Delete(ctx, cm1)).To(Succeed())
			Expect(s.Objects()).To(BeEmpty())
		})

		It("should error if the specified object does not exist", func() {
			Expect(s.Delete(ctx, cm1)).To(Equal(apierrors.NewNotFound(cmGR, cm1Key.String())))
		})

		DescribeTable("unsupported options",
			func(opts ...client.DeleteOption) {
				Expect(s.Delete(ctx, &corev1.ConfigMap{}, opts...)).To(HaveOccurred())
			},
			Entry("dry run", client.DryRunAll),
			Entry("grace period seconds", client.GracePeriodSeconds(1)),
			Entry("preconditions", client.Preconditions{}),
			Entry("propagation policy", client.PropagationPolicy(metav1.DeletePropagationOrphan)),
		)
	})

	Describe("DeleteAllOf", func() {
		DescribeTable("successful delete all of",
			func(opts []client.DeleteAllOfOption, expected ...interface{}) {
				Expect(s.Create(ctx, cm1)).To(Succeed())
				Expect(s.Create(ctx, cm2)).To(Succeed())
				Expect(s.DeleteAllOf(ctx, &corev1.ConfigMap{}, opts...))
				Expect(s.Objects()).To(ConsistOf(expected...))
			},
			Entry("all config maps", nil, nil),
			Entry("in namespace", []client.DeleteAllOfOption{client.InNamespace(namespace)}, cm2),
			Entry("matching labels", []client.DeleteAllOfOption{client.MatchingLabels{"foo": "bar"}}, cm2),
		)

		DescribeTable("unsupported options",
			func(opts ...client.DeleteAllOfOption) {
				Expect(s.DeleteAllOf(ctx, &corev1.ConfigMap{}, opts...)).To(HaveOccurred())
			},
			Entry("raw", &client.DeleteAllOfOptions{ListOptions: client.ListOptions{Raw: &metav1.ListOptions{}}}),
			Entry("field selector", client.MatchingFields{"foo": "bar"}),
			Entry("dry run", client.DryRunAll),
			Entry("grace period seconds", client.GracePeriodSeconds(1)),
			Entry("preconditions", client.Preconditions{}),
			Entry("propagation policy", client.PropagationPolicy(metav1.DeletePropagationOrphan)),
		)
	})

	Describe("Update", func() {
		It("should update the object", func() {
			Expect(s.Create(ctx, cm1)).To(Succeed())
			Expect(s.Update(ctx, cm1)).To(Succeed())
		})

		It("should error if the object does not exist", func() {
			Expect(s.Update(ctx, cm1)).To(Equal(apierrors.NewNotFound(cmGR, cm1Key.String())))
		})

		DescribeTable("unsupported options",
			func(opts ...client.UpdateOption) {
				Expect(s.Update(ctx, &corev1.ConfigMap{}, opts...)).To(HaveOccurred())
			},
			Entry("dry run", client.DryRunAll),
			Entry("raw", &client.UpdateOptions{Raw: &metav1.UpdateOptions{}}),
			Entry("field manager", client.FieldOwner("foo")),
		)
	})

	Describe("Patch", func() {
		It("should not be supported", func() {
			Expect(s.Patch(ctx, &corev1.ConfigMap{}, client.Apply, client.FieldOwner("foo"))).To(HaveOccurred())
		})
	})

	Describe("Scheme", func() {
		It("should return the used scheme", func() {
			Expect(s.Scheme()).To(Equal(scheme.Scheme))
		})
	})

	Describe("RESTMapper", func() {
		It("should return nil", func() {
			Expect(s.RESTMapper()).To(BeNil())
		})
	})
})
