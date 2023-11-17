// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package clientutils_test

import (
	. "github.com/ironcore-dev/controller-utils/clientutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("ObjectKey", func() {
	Context("ObjectKeySet", func() {
		var (
			k1, k2, k3, k4, k5, k6 client.ObjectKey
		)
		BeforeEach(func() {
			k1 = client.ObjectKey{
				Namespace: "n1",
				Name:      "foo",
			}
			k2 = client.ObjectKey{
				Namespace: "n1",
				Name:      "bar",
			}
			k3 = client.ObjectKey{
				Namespace: "n2",
				Name:      "foo",
			}
			k4 = client.ObjectKey{
				Namespace: "n2",
				Name:      "bar",
			}
			k5 = client.ObjectKey{
				Name: "cluster1",
			}
			k6 = client.ObjectKey{
				Name: "cluster2",
			}
		})

		Describe("New", func() {
			It("should initialize a new object key set with the given elements", func() {
				s := NewObjectKeySet(k1, k2, k3, k4, k5, k6)
				Expect(s).To(Equal(ObjectKeySet{
					k1: struct{}{},
					k2: struct{}{},
					k3: struct{}{},
					k4: struct{}{},
					k5: struct{}{},
					k6: struct{}{},
				}))
			})
		})

		Describe("Insert", func() {
			It("should insert the given items", func() {
				s := NewObjectKeySet()
				s.Insert(k1, k2, k3)
				Expect(s).To(Equal(ObjectKeySet{
					k1: struct{}{},
					k2: struct{}{},
					k3: struct{}{},
				}))

				s.Insert(k4, k5, k6)
				Expect(s).To(Equal(ObjectKeySet{
					k1: struct{}{},
					k2: struct{}{},
					k3: struct{}{},
					k4: struct{}{},
					k5: struct{}{},
					k6: struct{}{},
				}))
			})
		})

		Describe("Delete", func() {
			It("should delete the given items from the set", func() {
				s := NewObjectKeySet(k1, k2, k3, k4, k5, k6)

				s.Delete(k2, k4, k6)

				Expect(s).To(Equal(ObjectKeySet{
					k1: struct{}{},
					k3: struct{}{},
					k5: struct{}{},
				}))
			})
		})

		Describe("Has", func() {
			It("should return whether the given item is present in the set", func() {
				s := NewObjectKeySet(k1, k2, k3)

				Expect(s.Has(k1)).To(BeTrue(), "set should have key: set %#v, key %s", s, k1)
				Expect(s.Has(k2)).To(BeTrue(), "set should have key: set %#v, key %s", s, k2)
				Expect(s.Has(k3)).To(BeTrue(), "set should have key: set %#v, key %s", s, k3)
				Expect(s.Has(k4)).To(BeFalse(), "set should not have key: set %#v, key %s", s, k4)
				Expect(s.Has(k5)).To(BeFalse(), "set should not have key: set %#v, key %s", s, k5)
				Expect(s.Has(k6)).To(BeFalse(), "set should not have key: set %#v, key %s", s, k6)
			})
		})

		Describe("Len", func() {
			It("should return the length of the set", func() {
				Expect(NewObjectKeySet(k1, k2, k3).Len()).To(Equal(3))
				Expect((ObjectKeySet)(nil).Len()).To(Equal(0))
				Expect(NewObjectKeySet().Len()).To(Equal(0))
			})
		})
	})
})
