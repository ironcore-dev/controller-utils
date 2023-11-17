// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package clientutils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("ObjectRef", func() {
	var (
		namespace string
		cm        *corev1.ConfigMap
		cmGK      schema.GroupKind
		cmRef     ObjectRef

		pod    *corev1.Pod
		podGK  schema.GroupKind
		podRef ObjectRef

		emptyU *unstructured.Unstructured
	)
	BeforeEach(func() {
		namespace = corev1.NamespaceDefault
		cm = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      "my-cm",
			},
		}
		cmGK = schema.GroupKind{
			Group: corev1.GroupName,
			Kind:  "ConfigMap",
		}
		cmRef = ObjectRef{
			GroupKind: cmGK,
			Key:       client.ObjectKeyFromObject(cm),
		}

		pod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      "my-pod",
			},
		}
		podGK = schema.GroupKind{
			Group: corev1.GroupName,
			Kind:  "Pod",
		}
		podRef = ObjectRef{
			GroupKind: podGK,
			Key:       client.ObjectKeyFromObject(pod),
		}

		emptyU = &unstructured.Unstructured{}
	})

	Describe("ObjectRefFromObject", func() {
		It("should create an object reference from the given object", func() {
			ref, err := ObjectRefFromObject(scheme.Scheme, cm)
			Expect(err).NotTo(HaveOccurred())
			Expect(ref).To(Equal(cmRef))
		})

		It("should error if it cannot determine the group kind of an object", func() {
			_, err := ObjectRefFromObject(scheme.Scheme, emptyU)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ObjectRefsFromObjects", func() {
		It("should create an object reference from the given object", func() {
			refs, err := ObjectRefsFromObjects(scheme.Scheme, []client.Object{cm, pod})
			Expect(err).NotTo(HaveOccurred())
			Expect(refs).To(Equal([]ObjectRef{cmRef, podRef}))
		})

		It("should error if it cannot determine the group kind of an object", func() {
			_, err := ObjectRefsFromObjects(scheme.Scheme, []client.Object{cm, emptyU, pod})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ObjectRefFromGetRequest", func() {
		It("should create an object reference from the given object", func() {
			ref, err := ObjectRefFromGetRequest(scheme.Scheme, GetRequestFromObject(cm))
			Expect(err).NotTo(HaveOccurred())
			Expect(ref).To(Equal(cmRef))
		})

		It("should error if it cannot determine the group kind of an object", func() {
			_, err := ObjectRefFromGetRequest(scheme.Scheme, GetRequestFromObject(emptyU))
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ObjectRefsFromGetRequests", func() {
		It("should create an object reference from the given object", func() {
			refs, err := ObjectRefsFromGetRequests(scheme.Scheme, []GetRequest{
				GetRequestFromObject(cm),
				GetRequestFromObject(pod),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(refs).To(Equal([]ObjectRef{cmRef, podRef}))
		})

		It("should error if it cannot determine the group kind of an object", func() {
			_, err := ObjectRefsFromGetRequests(scheme.Scheme, []GetRequest{
				GetRequestFromObject(cm),
				GetRequestFromObject(emptyU),
				GetRequestFromObject(pod),
			})
			Expect(err).To(HaveOccurred())
		})
	})

	Context("ObjectRefSet", func() {
		Describe("NewObjectRefSet", func() {
			It("should create a new object ref set with the given items", func() {
				s := NewObjectRefSet(cmRef)
				Expect(s).To(Equal(ObjectRefSet{
					cmRef: struct{}{},
				}))
			})
		})

		Describe("Has", func() {
			It("should determine whether the given item is in the set", func() {
				s := NewObjectRefSet(cmRef)
				Expect(s.Has(cmRef)).To(BeTrue())
				Expect(s.Has(podRef)).To(BeFalse())
			})
		})

		Describe("Delete", func() {
			It("should delete the item from the set if present", func() {
				s := NewObjectRefSet(cmRef)
				s.Delete(cmRef)
				s.Delete(podRef)
				Expect(s).To(Equal(ObjectRefSet{}))
			})
		})

		Describe("Insert", func() {
			It("should insert the item if not yet present", func() {
				s := NewObjectRefSet(cmRef)
				s.Insert(cmRef)
				s.Insert(cmRef)
				Expect(s).To(Equal(ObjectRefSet{cmRef: struct{}{}}))
			})
		})

		Describe("Len", func() {
			It("should report the correct length", func() {
				s := NewObjectRefSet(cmRef)
				Expect(s.Len()).To(Equal(1))
				s.Insert(cmRef)
				Expect(s.Len()).To(Equal(1))
				s.Insert(podRef)
				Expect(s.Len()).To(Equal(2))
				s.Delete(cmRef)
				Expect(s.Len()).To(Equal(1))
				s.Delete(podRef)
				Expect(s.Len()).To(Equal(0))
			})
		})

		Describe("ObjectRefSetReferencesObject", func() {
			It("should report whether the object is referenced by the set", func() {
				s := NewObjectRefSet(cmRef)
				ok, err := ObjectRefSetReferencesObject(scheme.Scheme, s, cm)
				Expect(err).NotTo(HaveOccurred())
				Expect(ok).To(BeTrue())

				ok, err = ObjectRefSetReferencesObject(scheme.Scheme, s, pod)
				Expect(err).NotTo(HaveOccurred())
				Expect(ok).To(BeFalse())
			})

			It("should error if it cannot obtain a reference from the object", func() {
				s := NewObjectRefSet(cmRef)
				_, err := ObjectRefSetReferencesObject(scheme.Scheme, s, emptyU)
				Expect(err).To(HaveOccurred())
			})
		})

		Describe("ObjectRefSetReferencesGetRequest", func() {
			It("should report whether the object is referenced by the set", func() {
				s := NewObjectRefSet(cmRef)
				ok, err := ObjectRefSetReferencesGetRequest(scheme.Scheme, s, GetRequestFromObject(cm))
				Expect(err).NotTo(HaveOccurred())
				Expect(ok).To(BeTrue())

				ok, err = ObjectRefSetReferencesGetRequest(scheme.Scheme, s, GetRequestFromObject(pod))
				Expect(err).NotTo(HaveOccurred())
				Expect(ok).To(BeFalse())
			})

			It("should error if it cannot obtain a reference from the request", func() {
				s := NewObjectRefSet(cmRef)
				_, err := ObjectRefSetReferencesGetRequest(scheme.Scheme, s, GetRequestFromObject(emptyU))
				Expect(err).To(HaveOccurred())
			})
		})

		Describe("ObjectRefSetFromObjects", func() {
			It("should create an ObjectRefSet from the given get request set", func() {
				s, err := ObjectRefSetFromObjects(scheme.Scheme, []client.Object{cm, pod})
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(ObjectRefSet{
					cmRef:  struct{}{},
					podRef: struct{}{},
				}))
			})

			It("should error if it cannot obtain a reference from an object", func() {
				_, err := ObjectRefSetFromObjects(scheme.Scheme, []client.Object{cm, emptyU, pod})
				Expect(err).To(HaveOccurred())
			})
		})

		Describe("ObjectRefSetFromGetRequestSet", func() {
			It("should create an ObjectRefSet from the given get request set", func() {
				s2 := NewGetRequestSet(GetRequestFromObject(cm), GetRequestFromObject(pod))
				s, err := ObjectRefSetFromGetRequestSet(scheme.Scheme, s2)
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(ObjectRefSet{
					cmRef:  struct{}{},
					podRef: struct{}{},
				}))
			})

			It("should error if it cannot obtain a reference from a request", func() {
				s2 := NewGetRequestSet(GetRequestFromObject(cm), GetRequestFromObject(emptyU), GetRequestFromObject(pod))
				_, err := ObjectRefSetFromGetRequestSet(scheme.Scheme, s2)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
