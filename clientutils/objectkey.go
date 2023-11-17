// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

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
