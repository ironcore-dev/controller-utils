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

package clientutils

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// ObjectRef references an object regardless of its version.
type ObjectRef struct {
	GroupKind schema.GroupKind
	Key       client.ObjectKey
}

// ObjectRefFromObject creates a new ObjectRef from the given client.Object.
func ObjectRefFromObject(scheme *runtime.Scheme, obj client.Object) (ObjectRef, error) {
	gvk, err := apiutil.GVKForObject(obj, scheme)
	if err != nil {
		return ObjectRef{}, err
	}

	return ObjectRef{Key: client.ObjectKeyFromObject(obj), GroupKind: gvk.GroupKind()}, nil
}

// ObjectRefsFromObjects creates a list of ObjectRef from a list of client.Object.
func ObjectRefsFromObjects(scheme *runtime.Scheme, objs []client.Object) ([]ObjectRef, error) {
	if objs == nil {
		return nil, nil
	}
	refs := make([]ObjectRef, 0, len(objs))
	for _, obj := range objs {
		ref, err := ObjectRefFromObject(scheme, obj)
		if err != nil {
			return nil, err
		}

		refs = append(refs, ref)
	}
	return refs, nil
}

// ObjectRefFromGetRequest creates a new ObjectRef from the given GetRequest.
func ObjectRefFromGetRequest(scheme *runtime.Scheme, req GetRequest) (ObjectRef, error) {
	gvk, err := apiutil.GVKForObject(req.Object, scheme)
	if err != nil {
		return ObjectRef{}, err
	}

	return ObjectRef{Key: req.Key, GroupKind: gvk.GroupKind()}, nil
}

// ObjectRefsFromGetRequests creates a list of ObjectRef from the given list of GetRequest.
func ObjectRefsFromGetRequests(scheme *runtime.Scheme, reqs []GetRequest) ([]ObjectRef, error) {
	if reqs == nil {
		return nil, nil
	}
	res := make([]ObjectRef, 0, len(reqs))
	for _, req := range reqs {
		ref, err := ObjectRefFromGetRequest(scheme, req)
		if err != nil {
			return nil, err
		}

		res = append(res, ref)
	}
	return res, nil
}

// ObjectRefSet is a set of ObjectRef references.
type ObjectRefSet map[ObjectRef]struct{}

// Insert inserts the given items into the set.
func (s ObjectRefSet) Insert(items ...ObjectRef) {
	for _, item := range items {
		s[item] = struct{}{}
	}
}

// Has checks if the given item is present in the set.
func (s ObjectRefSet) Has(item ObjectRef) bool {
	_, ok := s[item]
	return ok
}

// Delete deletes the given items from the set, if present.
func (s ObjectRefSet) Delete(items ...ObjectRef) {
	for _, item := range items {
		delete(s, item)
	}
}

// Len returns the length of the set.
func (s ObjectRefSet) Len() int {
	return len(s)
}

// NewObjectRefSet creates a new ObjectRefSet with the given set.
func NewObjectRefSet(items ...ObjectRef) ObjectRefSet {
	s := make(ObjectRefSet)
	s.Insert(items...)
	return s
}

// ObjectRefSetReferencesObject is a utility function to determine whether an ObjectRefSet contains a client.Object.
func ObjectRefSetReferencesObject(scheme *runtime.Scheme, s ObjectRefSet, obj client.Object) (bool, error) {
	ref, err := ObjectRefFromObject(scheme, obj)
	if err != nil {
		return false, err
	}

	return s.Has(ref), nil
}

// ObjectRefSetReferencesGetRequest is a utility function to determine whether an ObjectRefSet contains a GetRequest.
func ObjectRefSetReferencesGetRequest(scheme *runtime.Scheme, s ObjectRefSet, req GetRequest) (bool, error) {
	ref, err := ObjectRefFromGetRequest(scheme, req)
	if err != nil {
		return false, err
	}

	return s.Has(ref), nil
}

// ObjectRefSetFromObjects creates a new ObjectRefSet from the given list of client.Object.
func ObjectRefSetFromObjects(scheme *runtime.Scheme, objs []client.Object) (ObjectRefSet, error) {
	s := NewObjectRefSet()
	for _, obj := range objs {
		ref, err := ObjectRefFromObject(scheme, obj)
		if err != nil {
			return nil, err
		}

		s.Insert(ref)
	}
	return s, nil
}

// ObjectRefSetFromGetRequestSet creates a new ObjectRefSet from the given GetRequestSet.
func ObjectRefSetFromGetRequestSet(scheme *runtime.Scheme, s2 *GetRequestSet) (ObjectRefSet, error) {
	s := NewObjectRefSet()
	var err error
	s2.Iterate(func(request GetRequest) (cont bool) {
		var ref ObjectRef
		ref, err = ObjectRefFromGetRequest(scheme, request)
		if err != nil {
			return false
		}

		s.Insert(ref)
		return true
	})
	if err != nil {
		return nil, err
	}

	return s, nil
}
