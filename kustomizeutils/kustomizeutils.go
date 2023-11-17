// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package kustomizeutils

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// RunKustomize is a shorthand for running kustomize in a target directory.
func RunKustomize(dir string) (resmap.ResMap, error) {
	kustomizer := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
	return kustomizer.Run(filesys.MakeFsOnDisk(), dir)
}

// RunKustomizeIntoList is a shorthand for running kustomize and parsing the result into the given list.
func RunKustomizeIntoList(dir string, decoder runtime.Decoder, into runtime.Object) error {
	res, err := RunKustomize(dir)
	if err != nil {
		return fmt.Errorf("error running kustomize: %w", err)
	}

	if err := DecodeResMapIntoList(decoder, res, into); err != nil {
		return fmt.Errorf("error decoding resmap into list: %w", err)
	}
	return nil
}

// DecodeResource decodes a resource.Resource into a given runtime.Object, if given.
// Shorthand for resource.Resource.MarshalJSON + runtime.Codec.Decode.
func DecodeResource(decoder runtime.Decoder, res *resource.Resource, defaults *schema.GroupVersionKind, into runtime.Object) (runtime.Object, *schema.GroupVersionKind, error) {
	data, err := res.MarshalJSON()
	if err != nil {
		return nil, nil, fmt.Errorf("could not marshal resource: %w", err)
	}

	return decoder.Decode(data, defaults, into)
}

// DecodeResMapObjects decodes the resmap.ResMap objects into a slice of runtime.Object.
func DecodeResMapObjects(deccoder runtime.Decoder, resMap resmap.ResMap) ([]runtime.Object, error) {
	resources := resMap.Resources()
	res := make([]runtime.Object, 0, len(resources))
	for _, rsc := range resMap.Resources() {
		obj, _, err := DecodeResource(deccoder, rsc, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("error decoding object: %w", err)
		}

		res = append(res, obj)
	}
	return res, nil
}

// DecodeResMapIntoList decodes a resmap.ResMap into a list represented by runtime.Object.
func DecodeResMapIntoList(decoder runtime.Decoder, resMap resmap.ResMap, into runtime.Object) error {
	objs, err := DecodeResMapObjects(decoder, resMap)
	if err != nil {
		return fmt.Errorf("error decoding objects: %w", err)
	}

	if err := meta.SetList(into, objs); err != nil {
		return fmt.Errorf("error setting list: %w", err)
	}
	return nil
}

// DecodeResMapUnstructureds decodes a resmap.ResMap into a slice of unstructured.Unstructured.
func DecodeResMapUnstructureds(resMap resmap.ResMap) ([]unstructured.Unstructured, error) {
	res := make([]unstructured.Unstructured, 0, resMap.Size())
	for _, rsc := range resMap.Resources() {
		data, err := rsc.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("error marshaling resource to json: %w", err)
		}

		obj := &unstructured.Unstructured{}
		if _, _, err := unstructured.UnstructuredJSONScheme.Decode(data, nil, obj); err != nil {
			return nil, fmt.Errorf("error decoding unstructured: %w", err)
		}
		res = append(res, *obj)
	}
	return res, nil
}
