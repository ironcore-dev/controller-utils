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

// Package metautils provides utilities to work with objects on the meta layer.
package metautils

import (
	"fmt"
	"reflect"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
