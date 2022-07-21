// Copyright 2022 OnMetal authors
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

package set_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/onmetal/controller-utils/set"
)

var _ = Describe("Set", func() {
	Describe("New", func() {
		It("should initialize and set the set", func() {
			s := New[int](1, 2, 1)
			Expect(s).To(Equal(Set[int]{1: {}, 2: {}}))
		})
	})

	Describe("Insert", func() {
		It("should insert the items into the set", func() {
			s := New[int]()
			s.Insert(1, 2, 1)

			Expect(s).To(Equal(Set[int]{1: {}, 2: {}}))
		})
	})

	Describe("Delete", func() {
		It("should delete the items", func() {
			s := New[int](1, 2, 3, 4)
			s.Delete(1, 3)
			Expect(s).To(Equal(Set[int]{2: {}, 4: {}}))
		})
	})

	Describe("Has", func() {
		It("should report whether the value is present", func() {
			s := New[int](1, 2, 3, 4)

			Expect(s.Has(1)).To(BeTrue())
			Expect(s.Has(5)).To(BeFalse())
		})
	})

	Describe("HasAll", func() {
		It("should report whether all values are present", func() {
			s := New[int](1, 2, 3, 4)

			Expect(s.HasAll(1, 2, 3, 4)).To(BeTrue())
			Expect(s.HasAll(1, 3)).To(BeTrue())
			Expect(s.HasAll(5)).To(BeFalse())
			Expect(s.HasAll(1, 2, 5)).To(BeFalse())
		})
	})

	Describe("HasAny", func() {
		It("should report whether any value is present", func() {
			s := New[int](1, 2, 3, 4)

			Expect(s.HasAny(1, 2, 3, 4)).To(BeTrue())
			Expect(s.HasAny(1, 5, 10)).To(BeTrue())
			Expect(s.HasAny(5)).To(BeFalse())
			Expect(s.HasAny(0, 5)).To(BeFalse())
		})
	})

	Describe("Difference", func() {
		It("should return the difference of two sets", func() {
			s1 := New[int](1, 2, 3, 4)
			s2 := New[int](3, 4, 5, 6)

			Expect(s1.Difference(s2)).To(Equal(New[int](1, 2)))
			Expect(s2.Difference(s1)).To(Equal(New[int](5, 6)))
		})
	})

	Describe("Union", func() {
		It("should return the union of the two sets", func() {
			s1 := New[int](1, 2, 3, 4)
			s2 := New[int](3, 4, 5, 6)

			Expect(s1.Union(s2)).To(Equal(New[int](1, 2, 3, 4, 5, 6)))
		})
	})

	Describe("Intersection", func() {
		It("should return the intersection of the two sets", func() {
			s1 := New[int](1, 2, 3, 4)
			s2 := New[int](3, 4, 5, 6)

			Expect(s1.Intersection(s2)).To(Equal(New[int](3, 4)))
		})
	})

	Describe("IsSuperset", func() {
		It("should report whether a set is a superset of another", func() {
			s1 := New[int](1, 2, 3, 4)
			s2 := New[int](3, 4)

			Expect(s1.IsSuperset(s2)).To(BeTrue())
			Expect(s1.IsSuperset(s1)).To(BeTrue())
			Expect(s2.IsSuperset(s1)).To(BeFalse())
		})
	})

	Describe("Equal", func() {
		It("should report whether two slices are equal", func() {
			s1 := New[int](1, 2, 3, 4)
			s2 := New[int](3, 4, 5, 6)

			Expect(s1.Equal(s1)).To(BeTrue())
			Expect(s2.Equal(s2)).To(BeTrue())
			Expect(s1.Equal(s2)).To(BeFalse())
		})
	})

	Describe("PopAny", func() {
		It("should return any single element from the set", func() {
			s := New[int](1, 2, 3, 4)
			v, ok := s.PopAny()
			Expect(ok).To(BeTrue())
			Expect(v).To(BeElementOf(1, 2, 3, 4))

			empty := New[int]()
			v, ok = empty.PopAny()
			Expect(ok).To(BeFalse())
			Expect(v).To(BeZero())
		})
	})

	Describe("Len", func() {
		It("should report the length of the slice", func() {
			s := New[int](1, 2, 3, 4)
			Expect(s.Len()).To(Equal(4))

			empty := New[int]()
			Expect(empty.Len()).To(Equal(0))

			var nilSet Set[int]
			Expect(nilSet.Len()).To(Equal(0))
		})
	})
})
