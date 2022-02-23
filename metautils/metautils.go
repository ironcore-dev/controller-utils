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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
