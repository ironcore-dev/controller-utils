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

package set

import (
	"sort"

	"golang.org/x/exp/constraints"
)

// Empty is the presence marker in a Set.
type Empty struct{}

// Set is a set of unique values.
type Set[E comparable] map[E]Empty

// New creates a Set from a list of values.
func New[E comparable](items ...E) Set[E] {
	ss := Set[E]{}
	ss.Insert(items...)
	return ss
}

// Insert adds items to the set.
func (s Set[E]) Insert(items ...E) Set[E] {
	for _, item := range items {
		s[item] = Empty{}
	}
	return s
}

// Delete removes all items from the set.
func (s Set[E]) Delete(items ...E) Set[E] {
	for _, item := range items {
		delete(s, item)
	}
	return s
}

// Has returns true if and only if item is contained in the set.
func (s Set[E]) Has(item E) bool {
	_, contained := s[item]
	return contained
}

// HasAll returns true if and only if all items are contained in the set.
func (s Set[E]) HasAll(items ...E) bool {
	for _, item := range items {
		if !s.Has(item) {
			return false
		}
	}
	return true
}

// HasAny returns true if any items are contained in the set.
func (s Set[E]) HasAny(items ...E) bool {
	for _, item := range items {
		if s.Has(item) {
			return true
		}
	}
	return false
}

// Difference returns a set of objects that are not in other
// For example:
// s = {a1, a2, a3}
// other = {a1, a2, a4, a5}
// s.Difference(other) = {a3}
// other.Difference(s) = {a4, a5}
func (s Set[E]) Difference(other Set[E]) Set[E] {
	result := New[E]()
	for key := range s {
		if !other.Has(key) {
			result.Insert(key)
		}
	}
	return result
}

// Union returns a new set which includes items in either s or other.
// For example:
// s = {a1, a2}
// other = {a3, a4}
// s.Union(other) = {a1, a2, a3, a4}
// other.Union(s) = {a1, a2, a3, a4}
func (s Set[E]) Union(other Set[E]) Set[E] {
	result := New[E]()
	for key := range s {
		result.Insert(key)
	}
	for key := range other {
		result.Insert(key)
	}
	return result
}

// Intersection returns a new set which includes the item in BOTH s and other
// For example:
// s = {a1, a2}
// other = {a2, a3}
// s.Intersection(other) = {a2}
func (s Set[E]) Intersection(other Set[E]) Set[E] {
	var walk, compare Set[E]
	result := New[E]()
	if s.Len() < other.Len() {
		walk = s
		compare = other
	} else {
		walk = other
		compare = s
	}
	for key := range walk {
		if compare.Has(key) {
			result.Insert(key)
		}
	}
	return result
}

// IsSuperset returns true if and only if s is a superset of other.
func (s Set[E]) IsSuperset(other Set[E]) bool {
	for item := range other {
		if !s.Has(item) {
			return false
		}
	}
	return true
}

// Equal returns true if and only if s is equal (as a set) to other.
// Two sets are equal if their membership is identical.
// (In practice, this means same elements, order doesn't matter)
func (s Set[E]) Equal(other Set[E]) bool {
	return len(s) == len(other) && s.IsSuperset(other)
}

// Slice returns a slice of the items in random order.
func (s Set[E]) Slice() []E {
	res := make([]E, 0, len(s))
	for key := range s {
		res = append(res, key)
	}
	return res
}

// PopAny returns a single element from the set.
func (s Set[E]) PopAny() (E, bool) {
	for key := range s {
		s.Delete(key)
		return key, true
	}
	var zeroValue E
	return zeroValue, false
}

// Len returns the size of the set.
func (s Set[E]) Len() int {
	return len(s)
}

// SortedSlice takes a Set with constraints.Ordered items and returns a sorted slice of the items.
func SortedSlice[E constraints.Ordered](set Set[E]) []E {
	res := make([]E, 0, len(set))
	for item := range set {
		res = append(res, item)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i] < res[j]
	})
	return res
}
