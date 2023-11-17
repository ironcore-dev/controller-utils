// Copyright 2022 IronCore authors
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

package clientutils

import "sigs.k8s.io/controller-runtime/pkg/client"

// ObjectKeySet set is a set of client.ObjectKey.
type ObjectKeySet map[client.ObjectKey]struct{}

// Insert inserts the given items into the ObjectKeySet.
// The ObjectKeySet has to be non-nil for this operation.
func (s ObjectKeySet) Insert(items ...client.ObjectKey) {
	for _, item := range items {
		s[item] = struct{}{}
	}
}

// Has checks if the given item is in the set.
func (s ObjectKeySet) Has(item client.ObjectKey) bool {
	_, ok := s[item]
	return ok
}

// Delete removes the given items from the ObjectKeySet.
// The ObjectKeySet has to be non-nil for this operation.
func (s ObjectKeySet) Delete(items ...client.ObjectKey) {
	for _, item := range items {
		delete(s, item)
	}
}

// Len returns the length of the ObjectKeySet.
func (s ObjectKeySet) Len() int {
	return len(s)
}

// NewObjectKeySet creates a new ObjectKeySet and initializes it with the given items.
func NewObjectKeySet(items ...client.ObjectKey) ObjectKeySet {
	s := make(ObjectKeySet)
	s.Insert(items...)
	return s
}
