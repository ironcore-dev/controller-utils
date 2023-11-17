// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

// Package metautils provides utilities to work with objects on the meta layer.
package metautils

import (
	"fmt"
	"reflect"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// ConvertAndSetList converts the given runtime.Objects into the item type of the list and sets
// the list items to be the converted items.
func ConvertAndSetList(scheme *runtime.Scheme, list runtime.Object, objs []runtime.Object) error {
	elemType, err := ListElementType(list)
	if err != nil {
		return err
	}

	var converted []runtime.Object
	for _, obj := range objs {
		into := reflect.New(elemType).Interface()
		if err := scheme.Convert(obj, into, nil); err != nil {
			return err
		}

		converted = append(converted, into.(runtime.Object))
	}
	return meta.SetList(list, converted)
}

// GVKForList determines the schema.GroupVersionKind for the given list.
// Effectively, this strips a 'List' suffix from the kind, if it exists.
func GVKForList(scheme *runtime.Scheme, list runtime.Object) (schema.GroupVersionKind, error) {
	gvk, err := apiutil.GVKForObject(list, scheme)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}

	gvk.Kind = strings.TrimSuffix(gvk.Kind, "List")
	return gvk, nil
}

// ListElementType returns the element type of the list.
// For instance, for an appsv1.DeploymentList, the element type is appsv1.Deployment.
func ListElementType(list runtime.Object) (reflect.Type, error) {
	itemsPtr, err := meta.GetItemsPtr(list)
	if err != nil {
		return nil, err
	}

	v := reflect.ValueOf(itemsPtr)
	return v.Type().Elem().Elem(), nil
}

// IsControlledBy checks if controlled is controlled by owner.
// An object is considered to be controlled if there is a controller (via metav1.GetControllerOf) whose
// GVK, name and UID match with the controller object.
func IsControlledBy(scheme *runtime.Scheme, owner, controlled client.Object) (bool, error) {
	controller := metav1.GetControllerOf(controlled)
	if controller == nil {
		return false, nil
	}

	gvk, err := apiutil.GVKForObject(owner, scheme)
	if err != nil {
		return false, fmt.Errorf("error getting object kinds of owner: %w", err)
	}

	gv, err := schema.ParseGroupVersion(controller.APIVersion)
	if err != nil {
		return false, fmt.Errorf("could not parse controller api version: %w", err)
	}

	return gvk.GroupVersion() == gv &&
		controller.Kind == gvk.Kind &&
		controller.Name == owner.GetName() &&
		controller.UID == owner.GetUID(), nil
}

// FilterControlledBy filters multiple objects by using IsControlledBy on each item.
func FilterControlledBy(scheme *runtime.Scheme, owner client.Object, objects []client.Object) ([]client.Object, error) {
	var filtered []client.Object
	for _, object := range objects {
		ok, err := IsControlledBy(scheme, owner, object)
		if err != nil {
			return nil, err
		}
		if ok {
			object := object
			filtered = append(filtered, object)
		}
	}
	return filtered, nil
}

// ExtractList extracts the items of a list into a slice of client.Object.
func ExtractList(obj client.ObjectList) ([]client.Object, error) {
	itemsPtr, err := meta.GetItemsPtr(obj)
	if err != nil {
		return nil, err
	}
	items, err := conversion.EnforcePtr(itemsPtr)
	if err != nil {
		return nil, err
	}
	objects, err := extractObjectSlice(items)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", obj, err)
	}
	return objects, nil
}

// MustExtractList extracts the items of a list into a slice of client.Object.
// It panics if it cannot extract the list.
func MustExtractList(obj client.ObjectList) []client.Object {
	res, err := ExtractList(obj)
	utilruntime.Must(err)
	return res
}

func enforceSlice(obj interface{}) (reflect.Value, error) {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Slice {
		if v.Kind() == reflect.Invalid {
			return reflect.Value{}, fmt.Errorf("expected slice, but got invalid kind")
		}
		return reflect.Value{}, fmt.Errorf("expected slice, but got %v type", v.Type())
	}
	return v, nil
}

func enforceSlicePtr(obj interface{}) (reflect.Value, error) {
	items, err := conversion.EnforcePtr(obj)
	if err != nil {
		return reflect.Value{}, err
	}

	if _, err := enforceSlice(items.Interface()); err != nil {
		return reflect.Value{}, err
	}
	return items, nil
}

// ExtractObjectSlice extracts client.Object from a given slice.
func ExtractObjectSlice(slice interface{}) ([]client.Object, error) {
	items, err := enforceSlice(slice)
	if err != nil {
		return nil, err
	}
	return extractObjectSlice(items)
}

// MustExtractObjectSlice extracts client.Object from a given slice.
// It panics if it cannot extract the objects from the slice.
func MustExtractObjectSlice(slice interface{}) []client.Object {
	res, err := ExtractObjectSlice(slice)
	utilruntime.Must(err)
	return res
}

// ExtractObjectSlicePointer extracts client.Object from a given slice pointer.
func ExtractObjectSlicePointer(slicePtr interface{}) ([]client.Object, error) {
	items, err := enforceSlicePtr(slicePtr)
	if err != nil {
		return nil, err
	}
	return extractObjectSlice(items)
}

// MustExtractObjectSlicePointer extracts client.Object from a given slice pointer.
// It panics if it cannot extract the objects from the slice pointer.
func MustExtractObjectSlicePointer(slicePtr interface{}) []client.Object {
	res, err := ExtractObjectSlicePointer(slicePtr)
	utilruntime.Must(err)
	return res
}

func extractObjectSlice(items reflect.Value) ([]client.Object, error) {
	list := make([]client.Object, items.Len())
	for i := range list {
		raw := items.Index(i)
		switch item := raw.Interface().(type) {
		case runtime.RawExtension:
			switch {
			case item.Object != nil:
				var found bool
				if list[i], found = item.Object.(client.Object); !found {
					return nil, fmt.Errorf("item[%v]: expected object, got %#v", i, item.Object)
				}
			case item.Raw != nil:
				return nil, fmt.Errorf("item[%v]: expected object, got runtime.RawExtension.Raw", i)
			default:
				list[i] = nil
			}
		case client.Object:
			list[i] = item
		default:
			var found bool
			if list[i], found = raw.Addr().Interface().(client.Object); !found {
				return nil, fmt.Errorf("item[%v]: Expected object, got %#v(%s)", i, raw.Interface(), raw.Kind())
			}
		}
	}
	return list, nil
}

// SetObjectSlice sets a slice pointer's values to the given objects.
func SetObjectSlice(slicePtr interface{}, objects []client.Object) error {
	items, err := enforceSlicePtr(slicePtr)
	if err != nil {
		return err
	}

	return setObjectSlice(items, objects)
}

// MustSetObjectSlice sets a slice pointer's values to the given objects.
// It panics if it cannot set the slice pointer values to the given objects.
func MustSetObjectSlice(slicePtr interface{}, objects []client.Object) {
	utilruntime.Must(SetObjectSlice(slicePtr, objects))
}

func setObjectSlice(items reflect.Value, objects []client.Object) error {
	if items.Type() == objectSliceType {
		items.Set(reflect.ValueOf(objects))
		return nil
	}

	slice := reflect.MakeSlice(items.Type(), len(objects), len(objects))
	for i := range objects {
		dest := slice.Index(i)
		if dest.Type() == reflect.TypeOf(runtime.RawExtension{}) {
			dest = dest.FieldByName("Object")
		}

		// check to see if you're directly assignable
		if reflect.TypeOf(objects[i]).AssignableTo(dest.Type()) {
			dest.Set(reflect.ValueOf(objects[i]))
			continue
		}

		src, err := conversion.EnforcePtr(objects[i])
		if err != nil {
			return err
		}
		if src.Type().AssignableTo(dest.Type()) {
			dest.Set(src)
		} else if src.Type().ConvertibleTo(dest.Type()) {
			dest.Set(src.Convert(dest.Type()))
		} else {
			return fmt.Errorf("item[%d]: can't assign or convert %v into %v", i, src.Type(), dest.Type())
		}
	}
	items.Set(slice)
	return nil
}

// objectSliceType is the type of a slice of Objects
var objectSliceType = reflect.TypeOf([]client.Object{})

// SetList sets the items in a client.ObjectList to the given objects.
func SetList(list client.ObjectList, objects []client.Object) error {
	itemsPtr, err := meta.GetItemsPtr(list)
	if err != nil {
		return err
	}
	items, err := conversion.EnforcePtr(itemsPtr)
	if err != nil {
		return err
	}
	return setObjectSlice(items, objects)
}

// MustSetList sets the items in a client.ObjectList to the given objects.
// It panics if it cannot set the items in the client.ObjectList.
func MustSetList(list client.ObjectList, objects []client.Object) {
	utilruntime.Must(SetList(list, objects))
}

// NewListForGVK creates a new client.ObjectList for the given singular schema.GroupVersionKind.
func NewListForGVK(scheme *runtime.Scheme, gvk schema.GroupVersionKind) (client.ObjectList, error) {
	// This is considered to be good-enough (used across controller-runtime).
	gvk = gvk.GroupVersion().WithKind(gvk.Kind + "List")
	obj, err := scheme.New(gvk)
	if err != nil {
		return nil, fmt.Errorf("error creating list for %s: %w", gvk, err)
	}

	list, ok := obj.(client.ObjectList)
	if !ok {
		return nil, fmt.Errorf("object %T does not implement client.ObjectList", obj)
	}

	return list, nil
}

// NewListForObject creates a new client.ObjectList for the given singular client.Object.
//
// This method disambiguates depending on the type of the given object:
// * If the given object is *unstructured.Unstructured, an *unstructured.UnstructuredList will be returned.
// * If the given object is *metav1.PartialObjectMetadata, a *metav1.PartialObjectMetadataList will be returned.
// * For all other cases, a new object with the corresponding kind will be created using scheme.New().
func NewListForObject(scheme *runtime.Scheme, obj client.Object) (schema.GroupVersionKind, client.ObjectList, error) {
	switch obj.(type) {
	case *unstructured.Unstructured:
		gvk := obj.GetObjectKind().GroupVersionKind()
		list := &unstructured.UnstructuredList{}
		list.SetGroupVersionKind(gvk)
		return gvk, list, nil
	case *metav1.PartialObjectMetadata:
		gvk := obj.GetObjectKind().GroupVersionKind()
		list := &metav1.PartialObjectMetadataList{}
		list.SetGroupVersionKind(gvk)
		return gvk, list, nil
	default:
		gvk, err := apiutil.GVKForObject(obj, scheme)
		if err != nil {
			return schema.GroupVersionKind{}, nil, fmt.Errorf("error getting gvk for %T: %w", obj, err)
		}
		list, err := NewListForGVK(scheme, gvk)
		if err != nil {
			return schema.GroupVersionKind{}, nil, fmt.Errorf("error creating list for %s: %w", gvk, err)
		}
		return gvk, list, nil
	}
}

// EachListItem traverses over all items of the client.ObjectList.
func EachListItem(list client.ObjectList, f func(obj client.Object) error) error {
	return meta.EachListItem(list, func(rObj runtime.Object) error {
		obj, ok := rObj.(client.Object)
		if !ok {
			return fmt.Errorf("object %T does not implement client.Object", rObj)
		}

		return f(obj)
	})
}

// FilterList filters the list with the given function, mutating it in-place with the filtered objects.
func FilterList(list client.ObjectList, f func(obj client.Object) bool) error {
	var filtered []client.Object
	if err := EachListItem(list, func(obj client.Object) error {
		if f(obj) {
			filtered = append(filtered, obj)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("error filtering list: %w", err)
	}

	return SetList(list, filtered)
}

type ObjectLabels interface {
	GetLabels() map[string]string
	SetLabels(labels map[string]string)
}

type ObjectAnnotations interface {
	GetAnnotations() map[string]string
	SetAnnotations(annotations map[string]string)
}

// HasLabel checks if the object has a label with the given key.
func HasLabel(obj ObjectLabels, key string) bool {
	_, ok := obj.GetLabels()[key]
	return ok
}

// SetLabel sets the given label on the object.
func SetLabel(obj ObjectLabels, key, value string) {
	labels := obj.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}

	labels[key] = value
	obj.SetLabels(labels)
}

// SetLabels sets the given labels on the object.
func SetLabels(obj ObjectLabels, set map[string]string) {
	labels := obj.GetLabels()
	if labels == nil {
		labels = set
	} else {
		for k, v := range set {
			labels[k] = v
		}
	}
	obj.SetLabels(labels)
}

// DeleteLabel deletes the label with the given key from the object.
func DeleteLabel(obj ObjectLabels, key string) {
	labels := obj.GetLabels()
	delete(labels, key)
	obj.SetLabels(labels)
}

// DeleteLabels deletes the labels with the given keys from the object.
func DeleteLabels(obj ObjectLabels, keys []string) {
	labels := obj.GetLabels()
	if len(labels) == 0 {
		return
	}
	for _, key := range keys {
		delete(labels, key)
	}
	obj.SetLabels(labels)
}

// HasAnnotation checks if the object has an annotation with the given key.
func HasAnnotation(obj ObjectAnnotations, key string) bool {
	_, ok := obj.GetAnnotations()[key]
	return ok
}

// SetAnnotation sets the given annotation on the object.
func SetAnnotation(obj ObjectAnnotations, key, value string) {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations[key] = value
	obj.SetAnnotations(annotations)
}

// SetAnnotations sets the given annotations on the object.
func SetAnnotations(obj ObjectAnnotations, set map[string]string) {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = set
	} else {
		for k, v := range set {
			annotations[k] = v
		}
	}
	obj.SetAnnotations(annotations)
}

// DeleteAnnotation deletes the annotation with the given key from the object.
func DeleteAnnotation(obj ObjectAnnotations, key string) {
	annotations := obj.GetAnnotations()
	delete(annotations, key)
	obj.SetAnnotations(annotations)
}

// DeleteAnnotations deletes the annotations with the given keys from the object.
func DeleteAnnotations(obj ObjectAnnotations, keys []string) {
	annotations := obj.GetAnnotations()
	if len(annotations) == 0 {
		return
	}
	for _, key := range keys {
		delete(annotations, key)
	}
	obj.SetAnnotations(annotations)
}
