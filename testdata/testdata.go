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

// Package testdata contains test data for controller-utils.
package testdata

import (
	_ "embed"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	// ObjectsYAML is a yaml file containing multiple objects and empty documents.
	//go:embed bases/objects.yaml
	ObjectsYAML []byte

	// ConfigMapYAML is a yaml file containing a config map.
	//go:embed bases/cm.yaml
	ConfigMapYAML string
)

func Secret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: corev1.NamespaceDefault,
			Name:      "my-secret",
		},
		StringData: map[string]string{
			"foo": "bar",
		},
	}
}

func SecretKey() client.ObjectKey {
	return client.ObjectKeyFromObject(Secret())
}

func UnstructuredSecret() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Secret",
			"metadata": map[string]interface{}{
				"namespace": corev1.NamespaceDefault,
				"name":      "my-secret",
			},
			"stringData": map[string]interface{}{
				"foo": "bar",
			},
		},
	}
}

func ConfigMap() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "kube-system",
			Name:      "my-configmap",
		},
		Data: map[string]string{
			"baz": "qux",
		},
	}
}

func ConfigMapKey() client.ObjectKey {
	return client.ObjectKeyFromObject(ConfigMap())
}

func UnstructuredConfigMap() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"namespace": "kube-system",
				"name":      "my-configmap",
			},
			"data": map[string]interface{}{
				"baz": "qux",
			},
		},
	}
}

func UnstructuredMyConfigMap() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "my-config",
			},
			"data": map[string]interface{}{
				"foo": "bar",
			},
		},
	}
}

func Objects() []client.Object {
	return []client.Object{Secret(), ConfigMap()}
}

func UnstructuredObjects() []unstructured.Unstructured {
	return []unstructured.Unstructured{*UnstructuredSecret(), *UnstructuredConfigMap()}
}
