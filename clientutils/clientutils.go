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

// Package clientutils provides utilities for working with the client package of
// controller-runtime.
package clientutils

import (
	"context"
	"fmt"
	"reflect"

	"github.com/onmetal/controller-utils/metautils"
	"github.com/onmetal/controller-utils/unstructuredutils"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type clientMeta interface {
	Scheme() *runtime.Scheme
	RESTMapper() meta.RESTMapper
}

type nonReaderClient interface {
	client.Writer
	client.StatusClient
	clientMeta
}

type readerClient struct {
	client.Reader
	nonReaderClient
}

// ReaderClient returns a client.Client that uses the given client.Reader for all read operations and the
// client.Client for the remaining operations.
func ReaderClient(r client.Reader, c client.Client) client.Client {
	return readerClient{r, c}
}

// IgnoreAlreadyExists returns nil if the given error matches apierrors.IsAlreadyExists.
func IgnoreAlreadyExists(err error) error {
	if apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

// CreateMultipleFromFile creates multiple objects by reading the given file as unstructured objects and then creating
// the read objects using the given client and options.
func CreateMultipleFromFile(ctx context.Context, c client.Client, filename string, opts ...client.CreateOption) ([]unstructured.Unstructured, error) {
	objs, err := unstructuredutils.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	if err := CreateMultiple(ctx, c, unstructuredutils.UnstructuredSliceToObjectSliceNoCopy(objs), opts...); err != nil {
		return nil, err
	}

	return objs, nil
}

// CreateMultiple creates multiple objects using the given client and options.
func CreateMultiple(ctx context.Context, c client.Client, objs []client.Object, opts ...client.CreateOption) error {
	for _, obj := range objs {
		if err := c.Create(ctx, obj, opts...); err != nil {
			return fmt.Errorf("error creating object %s: %w",
				client.ObjectKeyFromObject(obj), err)
		}
	}
	return nil
}

// GetRequest is a request to get an object with the given key and object (that is later used to write the result into).
type GetRequest struct {
	Key    client.ObjectKey
	Object client.Object
}

// GetRequestFromObject converts the given client.Object to a GetRequest. Namespace and name should be present on
// the object.
func GetRequestFromObject(obj client.Object) GetRequest {
	return GetRequest{
		Key:    client.ObjectKeyFromObject(obj),
		Object: obj,
	}
}

// GetRequestsFromObjects converts each client.Object into a GetRequest using GetRequestFromObject.
func GetRequestsFromObjects(objs []client.Object) []GetRequest {
	if objs == nil {
		return nil
	}
	res := make([]GetRequest, 0, len(objs))
	for _, obj := range objs {
		res = append(res, GetRequestFromObject(obj))
	}
	return res
}

// ObjectsFromGetRequests retrieves all client.Object objects from the given slice of GetRequest.
func ObjectsFromGetRequests(reqs []GetRequest) []client.Object {
	if reqs == nil {
		return nil
	}
	objs := make([]client.Object, 0, len(reqs))
	for _, req := range reqs {
		objs = append(objs, req.Object)
	}
	return objs
}

type getRequestTypedKey struct {
	typ       reflect.Type
	objectKey client.ObjectKey
}

type getRequestUnstructuredKey struct {
	gvk       schema.GroupVersionKind
	objectKey client.ObjectKey
}

// GetRequestSet is a set of GetRequest.
//
// Internally, the objects are differentiated by either being typed or unstructured.
// For unstructured objects, the group version kind they supply alongside their client.ObjectKey is used as identity.
// For typed objects, their element type (all typed objects *have* to be pointers to structs) alongside their
// client.ObjectKey is used as identity.
// If a typed object is *not* a pointer to a struct, a panic will happen.
type GetRequestSet struct {
	typed        map[getRequestTypedKey]client.Object
	unstructured map[getRequestUnstructuredKey]client.Object
}

func (s *GetRequestSet) unstructuredKey(req GetRequest) getRequestUnstructuredKey {
	u := req.Object.(*unstructured.Unstructured)
	return getRequestUnstructuredKey{
		gvk:       u.GroupVersionKind(),
		objectKey: req.Key,
	}
}

func (s *GetRequestSet) typedKey(req GetRequest) getRequestTypedKey {
	t := reflect.TypeOf(req.Object)
	// Taken from runtime.Scheme.AddKnownTypes.
	// In this case it's fine to panic as we distinguish between typed and unstructured
	// objects beforehand.
	if t.Kind() != reflect.Ptr {
		panic("All types must be pointers to structs")
	}
	t = t.Elem()
	if t.Kind() != reflect.Struct {
		panic("All types must be pointers to struct")
	}
	return getRequestTypedKey{
		typ:       t,
		objectKey: req.Key,
	}
}

// Insert inserts the given items into the set.
func (s *GetRequestSet) Insert(items ...GetRequest) {
	for _, item := range items {
		switch item.Object.(type) {
		case *unstructured.Unstructured:
			s.unstructured[s.unstructuredKey(item)] = item.Object
		default:
			s.typed[s.typedKey(item)] = item.Object
		}
	}
}

// Len returns the length of the set.
func (s *GetRequestSet) Len() int {
	return len(s.typed) + len(s.unstructured)
}

// Has checks if the given item is present in the set.
func (s *GetRequestSet) Has(item GetRequest) bool {
	var ok bool
	switch item.Object.(type) {
	case *unstructured.Unstructured:
		_, ok = s.unstructured[s.unstructuredKey(item)]
	default:
		_, ok = s.typed[s.typedKey(item)]
	}
	return ok
}

// Delete deletes the given items from the set, if they were present.
func (s *GetRequestSet) Delete(items ...GetRequest) {
	for _, item := range items {
		switch item.Object.(type) {
		case *unstructured.Unstructured:
			delete(s.unstructured, s.unstructuredKey(item))
		default:
			delete(s.typed, s.typedKey(item))
		}
	}
}

// Iterate iterates through the get requests of this set using the given function.
// If the function returns true (i.e. stop), the iteration is canceled.
func (s *GetRequestSet) Iterate(f func(GetRequest) (cont bool)) {
	for k, v := range s.typed {
		if cont := f(GetRequest{Key: k.objectKey, Object: v}); !cont {
			return
		}
	}
	for k, v := range s.unstructured {
		if cont := f(GetRequest{Key: k.objectKey, Object: v}); !cont {
			return
		}
	}
}

// List returns all GetRequests of this set.
func (s *GetRequestSet) List() []GetRequest {
	res := make([]GetRequest, 0, s.Len())
	s.Iterate(func(request GetRequest) (cont bool) {
		res = append(res, request)
		return true
	})
	return res
}

// NewGetRequestSet creates a new set of GetRequest.
//
// Internally, the objects are differentiated by either being typed or unstructured.
// For unstructured objects, the group version kind they supply alongside their client.ObjectKey is used as identity.
// For typed objects, their element type (all typed objects *have* to be pointers to structs) alongside their
// client.ObjectKey is used as identity.
// If a typed object is *not* a pointer to a struct, a panic will happen.
func NewGetRequestSet(items ...GetRequest) *GetRequestSet {
	s := &GetRequestSet{
		typed:        make(map[getRequestTypedKey]client.Object),
		unstructured: make(map[getRequestUnstructuredKey]client.Object),
	}
	s.Insert(items...)
	return s
}

// GetMultipleFromFile creates multiple objects by reading the given file as unstructured objects and then creating
// the read objects using the given client and options.
func GetMultipleFromFile(ctx context.Context, c client.Client, filename string) ([]unstructured.Unstructured, error) {
	objs, err := unstructuredutils.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	reqs := make([]GetRequest, 0, len(objs))
	for i := range objs {
		obj := &objs[i]
		reqs = append(reqs, GetRequest{
			Key:    client.ObjectKeyFromObject(obj),
			Object: obj,
		})
	}

	if err := GetMultiple(ctx, c, reqs); err != nil {
		return nil, err
	}

	return objs, nil
}

// GetMultiple gets multiple objects using the given client. The results are written back into the given GetRequest.
func GetMultiple(ctx context.Context, c client.Client, reqs []GetRequest) error {
	for _, req := range reqs {
		if err := c.Get(ctx, req.Key, req.Object); err != nil {
			return fmt.Errorf("error getting object %s: %w", req.Key, err)
		}
	}
	return nil
}

// apply is a PatchProvider always providing client.Apply.
type apply struct{}

// PatchFor implements PatchProvider.
func (a apply) PatchFor(obj client.Object) client.Patch {
	return client.Apply
}

// ApplyAll provides client.Apply for any given object.
var ApplyAll = apply{}

// PatchProvider retrieves a patch for any given object.
type PatchProvider interface {
	PatchFor(obj client.Object) client.Patch
}

// PatchRequest is the request to patch an object with a patch.
type PatchRequest struct {
	Object client.Object
	Patch  client.Patch
}

// PatchRequestFromObjectAndProvider is a shorthand to create a PatchRequest using a client.Object and PatchProvider.
func PatchRequestFromObjectAndProvider(obj client.Object, provider PatchProvider) PatchRequest {
	return PatchRequest{
		Object: obj,
		Patch:  provider.PatchFor(obj),
	}
}

// PatchRequestsFromObjectsAndProvider converts all client.Object objects to PatchRequest using
// PatchRequestFromObjectAndProvider.
func PatchRequestsFromObjectsAndProvider(objs []client.Object, provider PatchProvider) []PatchRequest {
	if objs == nil {
		return nil
	}
	res := make([]PatchRequest, 0, len(objs))
	for _, obj := range objs {
		res = append(res, PatchRequestFromObjectAndProvider(obj, provider))
	}
	return res
}

// ObjectsFromPatchRequests extracts all client.Object objects from the given slice of PatchRequest.
func ObjectsFromPatchRequests(reqs []PatchRequest) []client.Object {
	if reqs == nil {
		return nil
	}
	objs := make([]client.Object, 0, len(reqs))
	for _, req := range reqs {
		objs = append(objs, req.Object)
	}
	return objs
}

// PatchMultiple executes multiple PatchRequest with the given client.PatchOption.
func PatchMultiple(ctx context.Context, c client.Client, reqs []PatchRequest, opts ...client.PatchOption) error {
	for _, req := range reqs {
		if err := c.Patch(ctx, req.Object, req.Patch, opts...); err != nil {
			return fmt.Errorf("error patching object %s: %w",
				client.ObjectKeyFromObject(req.Object),
				err,
			)
		}
	}
	return nil
}

// PatchMultipleFromFile patches all objects from the given filename using the patchFor function.
// The returned unstructured.Unstructured objects contain the result of applying them.
func PatchMultipleFromFile(
	ctx context.Context,
	c client.Client,
	filename string,
	patchProvider PatchProvider,
	opts ...client.PatchOption,
) ([]unstructured.Unstructured, error) {
	objs, err := unstructuredutils.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	reqs := make([]PatchRequest, 0, len(objs))
	for i := range objs {
		obj := &objs[i]
		reqs = append(reqs, PatchRequest{obj, patchProvider.PatchFor(obj)})
	}

	if err := PatchMultiple(ctx, c, reqs, opts...); err != nil {
		return nil, err
	}

	return objs, nil
}

// DeleteMultipleFromFile deletes all client.Object objects from the given file with the given
// client.DeleteOption options.
func DeleteMultipleFromFile(ctx context.Context, c client.Client, filename string, opts ...client.DeleteOption) error {
	us, err := unstructuredutils.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	objs := unstructuredutils.UnstructuredSliceToObjectSliceNoCopy(us)
	return DeleteMultiple(ctx, c, objs, opts...)
}

// DeleteMultiple deletes multiple given client.Object objects using the given client.DeleteOption options.
func DeleteMultiple(ctx context.Context, c client.Client, objs []client.Object, opts ...client.DeleteOption) error {
	for _, obj := range objs {
		if err := c.Delete(ctx, obj, opts...); err != nil {
			return fmt.Errorf("error deleting object %s: %w",
				client.ObjectKeyFromObject(obj),
				err,
			)
		}
	}
	return nil
}

// ListAndFilter is a shorthand for doing a client.Client.List followed by filtering the list's elements
// with the given function.
func ListAndFilter(ctx context.Context, c client.Client, list client.ObjectList, filterFunc func(object client.Object) (bool, error), opts ...client.ListOption) error {
	if err := c.List(ctx, list, opts...); err != nil {
		return err
	}

	items, err := metautils.ExtractList(list)
	if err != nil {
		return fmt.Errorf("error extracting list: %w", err)
	}

	var filtered []client.Object
	for _, item := range items {
		ok, err := filterFunc(item)
		if err != nil {
			return err
		}
		if ok {
			item := item
			filtered = append(filtered, item)
		}
	}

	if err := metautils.SetList(list, filtered); err != nil {
		return fmt.Errorf("error setting list: %w", err)
	}

	return nil
}

// ListAndFilterControlledBy is a shorthand for doing a client.List followed by filtering the list's elements
// using metautils.IsControlledBy.
func ListAndFilterControlledBy(ctx context.Context, c client.Client, owner client.Object, list client.ObjectList, opts ...client.ListOption) error {
	scheme := c.Scheme()
	return ListAndFilter(ctx, c, list, func(object client.Object) (bool, error) {
		return metautils.IsControlledBy(scheme, owner, object)
	}, opts...)
}

func setObject(dst, src client.Object) error {
	dstV, err := conversion.EnforcePtr(dst)
	if err != nil {
		return err
	}

	srcV, err := conversion.EnforcePtr(src)
	if err != nil {
		return err
	}

	if !srcV.Type().AssignableTo(dstV.Type()) {
		return fmt.Errorf("cannot assign %T to %T", src, dst)
	}

	dstV.Set(srcV.Convert(dstV.Type()))
	return nil
}

// IsOlderThan returns a function that determines whether an object is older than another.
func IsOlderThan(obj client.Object) func(other client.Object) (bool, error) {
	return func(other client.Object) (bool, error) {
		return obj.GetCreationTimestamp().Time.After(other.GetCreationTimestamp().Time), nil
	}
}

// CreateOrUseAndPatch traverses through a slice of objects and tries to find a matching object using matchFunc.
// If it does, the matching object is set to the object, optionally patched and returned.
// If multiple objects match, the winning object is the oldest.
// If no object matches, initFunc is called and the new object is created.
// mutateFunc is optional, if none is specified no mutation will happen.
func CreateOrUseAndPatch(
	ctx context.Context,
	c client.Client,
	objects []client.Object,
	obj client.Object,
	matchFunc func() (bool, error),
	lessFunc func(other client.Object) (bool, error),
	mutateFunc func() error,
) (controllerutil.OperationResult, []client.Object, error) {
	var (
		base  = obj.DeepCopyObject().(client.Object)
		best  client.Object
		other []client.Object
	)
	for _, object := range objects {
		object := object
		if err := setObject(obj, object); err != nil {
			return controllerutil.OperationResultNone, nil, err
		}

		match, err := matchFunc()
		if err != nil {
			return controllerutil.OperationResultNone, nil, err
		}

		if match {
			if best == nil {
				best = object
				continue
			}

			less, err := lessFunc(best)
			if err != nil {
				return controllerutil.OperationResultNone, nil, err
			}
			if !less {
				other = append(other, best)
				best = object
				continue
			}
		}
		other = append(other, object)
	}
	if best != nil {
		if err := setObject(obj, best); err != nil {
			return controllerutil.OperationResultNone, nil, err
		}
		baseObj := obj.DeepCopyObject().(client.Object)
		if mutateFunc != nil {
			if err := mutateFunc(); err != nil {
				return controllerutil.OperationResultNone, nil, err
			}
		}
		if equality.Semantic.DeepEqual(baseObj, obj) {
			return controllerutil.OperationResultNone, other, nil
		}

		if err := c.Patch(ctx, obj, client.MergeFrom(baseObj)); err != nil {
			return controllerutil.OperationResultNone, nil, err
		}
		return controllerutil.OperationResultUpdated, other, nil
	}

	if err := setObject(obj, base); err != nil {
		return controllerutil.OperationResultNone, nil, err
	}
	if mutateFunc != nil {
		if err := mutateFunc(); err != nil {
			return controllerutil.OperationResultNone, nil, err
		}
	}
	if err := c.Create(ctx, obj); err != nil {
		return controllerutil.OperationResultNone, nil, err
	}
	return controllerutil.OperationResultCreated, other, nil
}

// DeleteIfExists deletes the given object, if it exists. It returns any non apierrors.IsNotFound error
// and whether the object actually existed or not.
func DeleteIfExists(ctx context.Context, c client.Client, obj client.Object, opts ...client.DeleteOption) (existed bool, err error) {
	if err := c.Delete(ctx, obj, opts...); err != nil {
		if !apierrors.IsNotFound(err) {
			return false, err
		}
		return false, nil
	}
	return true, nil
}

// DeleteMultipleIfExist deletes the given objects, if they exist. It returns any non apierrors.IsNotFound error
// and any object that existed before issuing the delete request.
func DeleteMultipleIfExist(ctx context.Context, c client.Client, objs []client.Object, opts ...client.DeleteOption) (existed []client.Object, err error) {
	for i, obj := range objs {
		ok, err := DeleteIfExists(ctx, c, obj, opts...)
		if err != nil {
			return existed, fmt.Errorf("[object %d]: error deleting %v: %w", i, obj, err)
		}
		if ok {
			obj := obj
			existed = append(existed, obj)
		}
	}
	return existed, nil
}

// PatchAddFinalizer issues a patch to add the given finalizer to the given object.
// The client.Patch method will be called regardless whether the finalizer was already present or not.
func PatchAddFinalizer(ctx context.Context, c client.Client, obj client.Object, finalizer string) error {
	baseObj := obj.DeepCopyObject().(client.Object)
	controllerutil.AddFinalizer(obj, finalizer)
	return c.Patch(ctx, obj, client.MergeFrom(baseObj))
}

// PatchRemoveFinalizer issues a patch to remove the given finalizer from the given object.
// The client.Patch method will be called regardless whether the finalizer was already gone or not.
func PatchRemoveFinalizer(ctx context.Context, c client.Client, obj client.Object, finalizer string) error {
	baseObj := obj.DeepCopyObject().(client.Object)
	controllerutil.RemoveFinalizer(obj, finalizer)
	return c.Patch(ctx, obj, client.MergeFrom(baseObj))
}

// PatchEnsureFinalizer checks if the given object has the given finalizer and, if not, issues a patch request
// to add it. The modified result reports whether the object had to be modified.
func PatchEnsureFinalizer(ctx context.Context, c client.Client, obj client.Object, finalizer string) (modified bool, err error) {
	if controllerutil.ContainsFinalizer(obj, finalizer) {
		return false, nil
	}

	if err := PatchAddFinalizer(ctx, c, obj, finalizer); err != nil {
		return false, err
	}
	return true, nil
}

// PatchEnsureNoFinalizer checks if the given object has the given finalizer and, if yes, issues a patch request
// to remove it. The modified result reports whether the object had to be modified.
func PatchEnsureNoFinalizer(ctx context.Context, c client.Client, obj client.Object, finalizer string) (modified bool, err error) {
	if !controllerutil.ContainsFinalizer(obj, finalizer) {
		return false, nil
	}

	if err := PatchRemoveFinalizer(ctx, c, obj, finalizer); err != nil {
		return false, err
	}
	return true, nil
}
