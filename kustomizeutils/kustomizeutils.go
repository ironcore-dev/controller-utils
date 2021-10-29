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
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// BuildKustomization is a shorthand for creating an in-memory directory, creating 'kustomization.yaml' with the
// yaml-encoded contents of the given types.Kustomization in it and running a krusty.Kustomizer on the directory.
// This is useful for quickly using types from remote repositories via Kustomize.
func BuildKustomization(kustomization types.Kustomization) (resmap.ResMap, error) {
	fs := filesys.MakeEmptyDirInMemory()
	data, err := yaml.Marshal(kustomization)
	if err != nil {
		return nil, fmt.Errorf("could not marshal kustomization: %w", err)
	}

	if err := fs.WriteFile("kustomization.yaml", data); err != nil {
		return nil, fmt.Errorf("could not write kustomization: %w", err)
	}

	kustomizer := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
	return kustomizer.Run(fs, ".")
}

// BuildKustomizationIntoList is a shorthand for BuildKustomization + DecodeResMapIntoList.
func BuildKustomizationIntoList(decoder runtime.Decoder, kustomization types.Kustomization, into runtime.Object) error {
	resMap, err := BuildKustomization(kustomization)
	if err != nil {
		return fmt.Errorf("error building kustomization: %w", err)
	}

	if err := DecodeResMapIntoList(decoder, resMap, into); err != nil {
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
