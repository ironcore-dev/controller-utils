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

// Package unstructuredutils provides utilities working with the unstructured.Unstructured type.
package unstructuredutils

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReadFile reads unstructured objects from a file with the given name.
// For further reference, have a look at Read.
func ReadFile(filename string) ([]unstructured.Unstructured, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		utilruntime.HandleError(f.Close())
	}()

	return Read(f)
}

// Read treats io.Reader as an incoming YAML or JSON stream and reads all unstructured.Unstructured objects of it.
//
// The document has to be well-formed. For multi-doc YAMLs, '---' is used as separator.
// Empty sub-documents are filtered from the resulting list.
func Read(r io.Reader) ([]unstructured.Unstructured, error) {
	d := yaml.NewYAMLOrJSONDecoder(bufio.NewReader(r), 4096)
	var objs []unstructured.Unstructured
	for {
		ext := runtime.RawExtension{}
		if err := d.Decode(&ext); err != nil {
			if !errors.Is(io.EOF, err) {
				return nil, fmt.Errorf("error parsing: %w", err)
			}
			return objs, nil
		}

		ext.Raw = bytes.TrimSpace(ext.Raw)
		if len(ext.Raw) == 0 || bytes.Equal(ext.Raw, []byte("null")) {
			continue
		}

		obj := &unstructured.Unstructured{}
		if _, _, err := scheme.Codecs.UniversalDeserializer().Decode(ext.Raw, nil, obj); err != nil {
			return nil, fmt.Errorf("invalid object: %w", err)
		}

		objs = append(objs, *obj)
	}
}

// UnstructuredSliceToObjectSliceNoCopy transforms the given list of unstructured.Unstructured to a list of
// client.Object, performing no copy while doing so.
//
// When creating the list, the resulting client.Object objects are obtained from having a pointer to the original
// slice item.
func UnstructuredSliceToObjectSliceNoCopy(unstructureds []unstructured.Unstructured) []client.Object {
	if unstructureds == nil {
		return nil
	}
	res := make([]client.Object, 0, len(unstructureds))
	for i := range unstructureds {
		res = append(res, &unstructureds[i])
	}
	return res
}

// UnstructuredSliceToObjectSlice transforms the given list of unstructured.Unstructured to a list of
// client.Object, copying the unstructured.Unstructured and using the pointers of them for the resulting client.Object.
func UnstructuredSliceToObjectSlice(unstructureds []unstructured.Unstructured) []client.Object {
	if unstructureds == nil {
		return nil
	}
	res := make([]client.Object, 0, len(unstructureds))
	for _, u := range unstructureds {
		u := u
		res = append(res, &u)
	}
	return res
}
