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
	"context"

	"github.com/golang/mock/gomock"
	mockclient "github.com/onmetal/controller-utils/mock/controller-runtime/client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("FieldIndexer", func() {
	var (
		ctx  context.Context
		ctrl *gomock.Controller
	)
	BeforeEach(func() {
		ctx = context.Background()
		ctrl = gomock.NewController(GinkgoT())
	})

	Context("SharedFieldIndexer", func() {
		var (
			fieldIndexer *mockclient.MockFieldIndexer
		)
		BeforeEach(func() {
			fieldIndexer = mockclient.NewMockFieldIndexer(ctrl)
		})

		It("should register an indexer func and call it", func() {
			f := mockclient.NewMockIndexerFunc(ctrl)
			gomock.InOrder(
				fieldIndexer.EXPECT().IndexField(ctx, &corev1.Pod{}, ".spec", gomock.Any()).Do(
					func(ctx context.Context, obj client.Object, field string, f client.IndexerFunc) error {
						f(obj)
						return nil
					}),
				f.EXPECT().Call(&corev1.Pod{}).Times(1),
			)

			idx := NewSharedFieldIndexer(fieldIndexer, scheme.Scheme)

			Expect(idx.Register(&corev1.Pod{}, ".spec", f.Call)).To(Succeed())

			Expect(idx.IndexField(ctx, &corev1.Pod{}, ".spec")).To(Succeed())
		})

		It("should error if a field is indexed twice", func() {
			f := mockclient.NewMockIndexerFunc(ctrl)
			idx := NewSharedFieldIndexer(fieldIndexer, scheme.Scheme)

			Expect(idx.Register(&corev1.Pod{}, ".spec", f.Call)).To(Succeed())
			Expect(idx.Register(&corev1.Pod{}, ".spec", f.Call)).To(MatchError("indexer for type *v1.Pod field .spec already registered"))
		})

		It("should call the index function only once", func() {
			f := mockclient.NewMockIndexerFunc(ctrl)
			gomock.InOrder(
				fieldIndexer.EXPECT().IndexField(ctx, &corev1.Pod{}, ".spec", gomock.Any()).Do(
					func(ctx context.Context, obj client.Object, field string, f client.IndexerFunc) error {
						f(obj)
						return nil
					}),
				f.EXPECT().Call(&corev1.Pod{}).Times(1),
			)

			idx := NewSharedFieldIndexer(fieldIndexer, scheme.Scheme)

			Expect(idx.Register(&corev1.Pod{}, ".spec", f.Call)).To(Succeed())

			Expect(idx.IndexField(ctx, &corev1.Pod{}, ".spec")).To(Succeed())
			Expect(idx.IndexField(ctx, &corev1.Pod{}, ".spec")).To(Succeed())
			Expect(idx.IndexField(ctx, &corev1.Pod{}, ".spec")).To(Succeed())
		})

		It("should work with unstructured objects", func() {
			pod := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
				},
			}
			f := mockclient.NewMockIndexerFunc(ctrl)
			gomock.InOrder(
				fieldIndexer.EXPECT().IndexField(ctx, pod, ".spec", gomock.Any()).Do(
					func(ctx context.Context, obj client.Object, field string, f client.IndexerFunc) error {
						f(obj)
						return nil
					}),
				f.EXPECT().Call(pod).Times(1),
			)

			idx := NewSharedFieldIndexer(fieldIndexer, scheme.Scheme)

			Expect(idx.Register(pod, ".spec", f.Call)).To(Succeed())

			Expect(idx.IndexField(ctx, pod, ".spec")).To(Succeed())
		})

		It("should work with partial object metadata", func() {
			pod := &metav1.PartialObjectMetadata{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Pod",
				},
			}
			f := mockclient.NewMockIndexerFunc(ctrl)
			gomock.InOrder(
				fieldIndexer.EXPECT().IndexField(ctx, pod, ".spec", gomock.Any()).Do(
					func(ctx context.Context, obj client.Object, field string, f client.IndexerFunc) error {
						f(obj)
						return nil
					}),
				f.EXPECT().Call(pod).Times(1),
			)

			idx := NewSharedFieldIndexer(fieldIndexer, scheme.Scheme)

			Expect(idx.Register(pod, ".spec", f.Call)).To(Succeed())

			Expect(idx.IndexField(ctx, pod, ".spec")).To(Succeed())
		})

		It("should error if the gvk could not be obtained", func() {
			f := mockclient.NewMockIndexerFunc(ctrl)
			idx := NewSharedFieldIndexer(fieldIndexer, scheme.Scheme)

			type CustomObject struct{ corev1.Pod }

			Expect(idx.Register(&CustomObject{}, ".spec", f.Call)).To(HaveOccurred())
		})

		It("should error if the indexed function is unknown", func() {
			idx := NewSharedFieldIndexer(fieldIndexer, scheme.Scheme)
			Expect(idx.IndexField(ctx, &corev1.Pod{}, "unknown")).To(HaveOccurred())
		})
	})
})
