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

package clientutils

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// SharedFieldIndexer allows registering and calling field index functions shared by different users.
type SharedFieldIndexer struct {
	indexer client.FieldIndexer
	*sharedFieldIndexerMap
}

// NewSharedFieldIndexer creates a new SharedFieldIndexer.
func NewSharedFieldIndexer(indexer client.FieldIndexer, scheme *runtime.Scheme) *SharedFieldIndexer {
	return &SharedFieldIndexer{
		indexer:               indexer,
		sharedFieldIndexerMap: newSharedFieldIndexerMap(scheme),
	}
}

// Register registers the client.IndexerFunc for the given client.Object and field.
func (s *SharedFieldIndexer) Register(obj client.Object, field string, extractValue client.IndexerFunc) error {
	updated, err := s.setIfNotPresent(obj, field, extractValue)
	if err != nil {
		return err
	}
	if !updated {
		return fmt.Errorf("indexer for type %T field %s already registered", obj, field)
	}
	return nil
}

// MustRegister registers the client.IndexerFunc for the given client.Object and field.
func (s *SharedFieldIndexer) MustRegister(obj client.Object, field string, extractValue client.IndexerFunc) {
	utilruntime.Must(s.Register(obj, field, extractValue))
}

// IndexField calls a registered client.IndexerFunc for the given client.Object and field.
// If the object / field is unknown or its GVK could not be determined, it errors.
func (s *SharedFieldIndexer) IndexField(ctx context.Context, obj client.Object, field string) error {
	entry, err := s.get(obj, field)
	if err != nil {
		return err
	}

	if entry == nil {
		return fmt.Errorf("unknown field %s for type %T", field, obj)
	}
	if entry.initialized {
		return nil
	}
	if err := s.indexer.IndexField(ctx, obj, field, entry.extractValue); err != nil {
		return err
	}
	entry.initialized = true
	return nil
}

type sharedFieldIndexerMap struct {
	scheme       *runtime.Scheme
	unstructured *specificSharedFieldIndexerMap
	metadata     *specificSharedFieldIndexerMap
	structured   *specificSharedFieldIndexerMap
}

func newSharedFieldIndexerMap(scheme *runtime.Scheme) *sharedFieldIndexerMap {
	return &sharedFieldIndexerMap{
		scheme:       scheme,
		unstructured: newSpecificSharedFieldIndexerMap(),
		metadata:     newSpecificSharedFieldIndexerMap(),
		structured:   newSpecificSharedFieldIndexerMap(),
	}
}

type mapEntry struct {
	initialized  bool
	extractValue client.IndexerFunc
}

type specificSharedFieldIndexerMap struct {
	gvkToNameToEntry map[schema.GroupVersionKind]map[string]*mapEntry
}

func newSpecificSharedFieldIndexerMap() *specificSharedFieldIndexerMap {
	return &specificSharedFieldIndexerMap{gvkToNameToEntry: make(map[schema.GroupVersionKind]map[string]*mapEntry)}
}

func (s *specificSharedFieldIndexerMap) get(gvk schema.GroupVersionKind, name string) *mapEntry {
	return s.gvkToNameToEntry[gvk][name]
}

func (s *specificSharedFieldIndexerMap) setIfNotPresent(gvk schema.GroupVersionKind, name string, extractValue client.IndexerFunc) (updated bool) {
	nameToEntry := s.gvkToNameToEntry[gvk]
	if nameToEntry == nil {
		nameToEntry = make(map[string]*mapEntry)
		s.gvkToNameToEntry[gvk] = nameToEntry
	}

	if _, ok := nameToEntry[name]; ok {
		return false
	}
	nameToEntry[name] = &mapEntry{extractValue: extractValue}
	return true
}

func (s *sharedFieldIndexerMap) mapFor(obj client.Object) (*specificSharedFieldIndexerMap, schema.GroupVersionKind, error) {
	gvk, err := apiutil.GVKForObject(obj, s.scheme)
	if err != nil {
		return nil, schema.GroupVersionKind{}, err
	}

	switch obj.(type) {
	case *unstructured.Unstructured:
		return s.unstructured, gvk, nil
	case *metav1.PartialObjectMetadata:
		return s.metadata, gvk, nil
	default:
		return s.structured, gvk, nil
	}
}

func (s *sharedFieldIndexerMap) get(obj client.Object, name string) (*mapEntry, error) {
	m, gvk, err := s.mapFor(obj)
	if err != nil {
		return nil, err
	}

	return m.get(gvk, name), nil
}

func (s *sharedFieldIndexerMap) setIfNotPresent(obj client.Object, name string, extractValue client.IndexerFunc) (updated bool, err error) {
	m, gvk, err := s.mapFor(obj)
	if err != nil {
		return false, err
	}

	return m.setIfNotPresent(gvk, name, extractValue), nil
}
