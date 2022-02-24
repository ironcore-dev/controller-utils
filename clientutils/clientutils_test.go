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

package clientutils_test

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/golang/mock/gomock"
	. "github.com/onmetal/controller-utils/clientutils"
	mockclient "github.com/onmetal/controller-utils/mock/controller-runtime/client"
	mockclientutils "github.com/onmetal/controller-utils/mock/controller-utils/clientutils"
	"github.com/onmetal/controller-utils/testdata"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ = Describe("Clientutils", func() {
	const (
		objectsPath = "../testdata/bases/objects.yaml"
	)

	var (
		ctx  context.Context
		ctrl *gomock.Controller

		c *mockclient.MockClient

		cmGR schema.GroupResource

		namespace string

		cm    *corev1.ConfigMap
		cmKey client.ObjectKey

		uPod *unstructured.Unstructured

		secret    *corev1.Secret
		secretKey client.ObjectKey

		patchProvider *mockclientutils.MockPatchProvider
	)
	setup := func() {
		ctx = context.Background()
		ctrl = gomock.NewController(GinkgoT())

		c = mockclient.NewMockClient(ctrl)

		cmGR = schema.GroupResource{Group: corev1.GroupName, Resource: "configmaps"}

		namespace = corev1.NamespaceDefault

		cm = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      "my-cm",
			},
		}
		cmKey = client.ObjectKeyFromObject(cm)

		uPod = &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]interface{}{
					"namespace": namespace,
					"name":      "my-pod",
				},
			},
		}

		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      "my-secret",
			},
		}
		secretKey = client.ObjectKeyFromObject(secret)

		patchProvider = mockclientutils.NewMockPatchProvider(ctrl)
	}
	setup()
	BeforeEach(setup)

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("IgnoreAlreadyExists", func() {
		It("should ignore already exists errors", func() {
			alreadyExistsErr := IgnoreAlreadyExists(apierrors.NewAlreadyExists(cmGR, ""))
			Expect(IgnoreAlreadyExists(alreadyExistsErr)).To(BeNil())
		})

		It("should not ignore other errors or nil", func() {
			err := fmt.Errorf("some error")
			Expect(IgnoreAlreadyExists(err)).To(BeIdenticalTo(err))
			Expect(IgnoreAlreadyExists(nil)).To(BeNil())
		})
	})

	Describe("CreateMultipleFromFile", func() {
		It("should error if the file does not exist", func() {
			_, err := CreateMultipleFromFile(ctx, c, "should-not-exist")
			Expect(err).To(HaveOccurred())
		})

		It("should abort and return any error from creating", func() {
			someErr := fmt.Errorf("some error")
			c.EXPECT().Create(ctx, testdata.UnstructuredSecret()).Return(someErr)

			_, err := CreateMultipleFromFile(ctx, c, objectsPath)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, someErr)).To(BeTrue())
		})

		It("should create the given objects from the file", func() {
			gomock.InOrder(
				c.EXPECT().Create(ctx, testdata.UnstructuredSecret()),
				c.EXPECT().Create(ctx, testdata.UnstructuredConfigMap()),
			)

			objs, err := CreateMultipleFromFile(ctx, c, objectsPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(objs).To(Equal(testdata.UnstructuredObjects()))
		})
	})

	Describe("CreateMultiple", func() {
		It("should abort and return any error from creating", func() {
			someErr := fmt.Errorf("some error")
			c.EXPECT().Create(ctx, cm).Return(someErr)

			err := CreateMultiple(ctx, c, []client.Object{cm, secret})
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, someErr)).To(BeTrue())
		})

		It("should create multiple objects", func() {
			gomock.InOrder(
				c.EXPECT().Create(ctx, cm),
				c.EXPECT().Create(ctx, secret),
			)
			Expect(CreateMultiple(ctx, c, []client.Object{cm, secret})).To(Succeed())
		})
	})

	Describe("GetRequestFromObject", func() {
		It("should create a get request from the given object", func() {
			Expect(GetRequestFromObject(cm)).To(Equal(GetRequest{
				Key:    cmKey,
				Object: cm,
			}))
		})
	})

	Describe("GetRequestsFromObjects", func() {
		It("should return nil if the requests are nil", func() {
			Expect(GetRequestsFromObjects(nil)).To(BeNil())
		})

		It("should create get requests from the given objects", func() {
			Expect(GetRequestsFromObjects([]client.Object{cm, secret})).To(Equal([]GetRequest{
				{
					Key:    cmKey,
					Object: cm,
				},
				{
					Key:    secretKey,
					Object: secret,
				},
			}))
		})
	})

	Describe("ObjectsFromGetRequests", func() {
		It("should extract the objects from the get requests", func() {
			Expect(ObjectsFromGetRequests([]GetRequest{
				{
					Key:    cmKey,
					Object: cm,
				},
				{
					Key:    secretKey,
					Object: secret,
				},
			})).To(Equal([]client.Object{cm, secret}))
		})
	})

	Context("GetRequestSet", func() {
		Describe("NewGetRequestSet", func() {
			It("should return a new get request set with the given items", func() {
				s := NewGetRequestSet(GetRequestFromObject(cm), GetRequestFromObject(uPod))

				Expect(s.Has(GetRequestFromObject(cm))).To(BeTrue())
				Expect(s.Has(GetRequestFromObject(uPod))).To(BeTrue())
				Expect(s.Len()).To(Equal(2))
			})
		})

		Describe("Insert", func() {
			It("should insert the given items into the set, deduping if necessary", func() {
				s := NewGetRequestSet()
				s.Insert(GetRequestFromObject(cm))
				s.Insert(GetRequestFromObject(cm))
				s.Insert(GetRequestFromObject(uPod))
				s.Insert(GetRequestFromObject(uPod))

				Expect(s.Has(GetRequestFromObject(cm))).To(BeTrue())
				Expect(s.Has(GetRequestFromObject(uPod))).To(BeTrue())
				Expect(s.Len()).To(Equal(2))
			})

			It("should panic if the object is typed but not a pointer to a struct", func() {
				type BadConfigMap struct {
					*corev1.ConfigMap
				}
				s := NewGetRequestSet()
				Expect(func() {
					s.Insert(GetRequestFromObject(BadConfigMap{cm}))
				}).To(Panic())
			})

			It("should panic if the object is typed and a pointer but not to a struct", func() {
				type BadObject struct {
					client.Object
				}
				s := NewGetRequestSet()
				Expect(func() {
					s.Insert(GetRequestFromObject(&BadObject{client.Object(nil)}))
				}).To(Panic())
			})
		})

		Describe("Has", func() {
			It("should determine whether the given item is present in the set", func() {
				s := NewGetRequestSet(GetRequestFromObject(cm))
				Expect(s.Has(GetRequestFromObject(cm))).To(BeTrue())
				Expect(s.Has(GetRequestFromObject(uPod))).To(BeFalse())
			})
		})

		Describe("Delete", func() {
			It("should delete the item so it's not present anymore", func() {
				s := NewGetRequestSet(GetRequestFromObject(cm))
				Expect(s.Has(GetRequestFromObject(cm))).To(BeTrue())
				Expect(s.Has(GetRequestFromObject(uPod))).To(BeFalse())

				s.Delete(GetRequestFromObject(cm))
				s.Delete(GetRequestFromObject(uPod))

				Expect(s.Has(GetRequestFromObject(cm))).To(BeFalse())
				Expect(s.Has(GetRequestFromObject(uPod))).To(BeFalse())
			})
		})

		Describe("Iterate", func() {
			It("should iterate through the entries in the set, stopping if requested (typed)", func() {
				s := NewGetRequestSet(GetRequestFromObject(cm), GetRequestFromObject(secret))
				var items []GetRequest
				s.Iterate(func(request GetRequest) (cont bool) {
					items = append(items, request)
					return false
				})

				Expect(items).To(SatisfyAny(
					Equal([]GetRequest{GetRequestFromObject(cm)}),
					Equal([]GetRequest{GetRequestFromObject(secret)}),
				))
			})

			It("should iterate through the entries in the set, stopping if requested (unstructured)", func() {
				s := NewGetRequestSet(GetRequestFromObject(testdata.UnstructuredSecret()), GetRequestFromObject(testdata.UnstructuredConfigMap()))
				var items []GetRequest
				s.Iterate(func(request GetRequest) (cont bool) {
					items = append(items, request)
					return false
				})

				Expect(items).To(SatisfyAny(
					Equal([]GetRequest{GetRequestFromObject(testdata.UnstructuredSecret())}),
					Equal([]GetRequest{GetRequestFromObject(testdata.UnstructuredConfigMap())}),
				))
			})

			It("should iterate through all elements if no stop is requested", func() {
				s := NewGetRequestSet(GetRequestFromObject(cm), GetRequestFromObject(secret))
				var items []GetRequest
				s.Iterate(func(request GetRequest) (cont bool) {
					items = append(items, request)
					return true
				})

				Expect(items).To(ConsistOf(GetRequestFromObject(cm), GetRequestFromObject(secret)))
			})
		})

		Describe("List", func() {
			It("should contain all entries as a list", func() {
				s := NewGetRequestSet(GetRequestFromObject(cm), GetRequestFromObject(uPod))
				Expect(s.List()).To(ConsistOf(GetRequestFromObject(cm), GetRequestFromObject(uPod)))
			})
		})
	})

	Describe("GetMultiple", func() {
		It("should abort and return any error from getting", func() {
			someErr := fmt.Errorf("some error")
			c.EXPECT().Get(ctx, cmKey, cm).Return(someErr)

			err := GetMultiple(ctx, c, []GetRequest{GetRequestFromObject(cm), GetRequestFromObject(secret)})
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, someErr)).To(BeTrue())
		})

		It("should get multiple referenced objects", func() {
			gomock.InOrder(
				c.EXPECT().Get(ctx, cmKey, cm),
				c.EXPECT().Get(ctx, secretKey, secret),
			)

			Expect(GetMultiple(ctx, c, []GetRequest{
				{
					Key:    cmKey,
					Object: cm,
				},
				{
					Key:    secretKey,
					Object: secret,
				},
			})).To(Succeed())
		})
	})

	Describe("GetMultipleFromFile", func() {
		It("should error if the file does not exist", func() {
			_, err := GetMultipleFromFile(ctx, c, "should-not-exist")
			Expect(err).To(HaveOccurred())
		})

		It("should abort and return any error from getting", func() {
			someErr := fmt.Errorf("some error")
			c.EXPECT().Get(ctx, client.ObjectKeyFromObject(testdata.UnstructuredSecret()), testdata.UnstructuredSecret()).Return(someErr)

			_, err := GetMultipleFromFile(ctx, c, objectsPath)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, someErr)).To(BeTrue())
		})

		It("should get multiple referenced objects from file", func() {
			gomock.InOrder(
				c.EXPECT().Get(ctx, testdata.SecretKey(), testdata.UnstructuredSecret()),
				c.EXPECT().Get(ctx, testdata.ConfigMapKey(), testdata.UnstructuredConfigMap()),
			)

			objs, err := GetMultipleFromFile(ctx, c, objectsPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(objs).To(Equal(testdata.UnstructuredObjects()))
		})
	})

	Describe("ApplyAll", func() {
		It("should return client.Apply for any object", func() {
			Expect(ApplyAll.PatchFor(cm)).To(Equal(client.Apply))
			Expect(ApplyAll.PatchFor(secret)).To(Equal(client.Apply))
			Expect(ApplyAll.PatchFor(uPod)).To(Equal(client.Apply))
		})
	})

	Describe("PatchRequestFromObjectAndProvider", func() {
		It("should create a patch request from the given object and provider", func() {
			patchProvider.EXPECT().PatchFor(cm).Return(client.Apply)
			Expect(PatchRequestFromObjectAndProvider(cm, patchProvider)).To(Equal(PatchRequest{
				Object: cm,
				Patch:  client.Apply,
			}))
		})
	})

	Describe("PatchRequestsFromObjectsAndProvider", func() {
		It("should return nil if the objects are nil", func() {
			Expect(PatchRequestsFromObjectsAndProvider(nil, patchProvider)).To(BeNil())
		})

		It("should create patch requests from the given objects and provider", func() {
			gomock.InOrder(
				patchProvider.EXPECT().PatchFor(cm).Return(client.Apply),
				patchProvider.EXPECT().PatchFor(secret).Return(client.Apply),
			)

			Expect(PatchRequestsFromObjectsAndProvider([]client.Object{cm, secret}, patchProvider)).To(Equal(
				[]PatchRequest{
					{
						Object: cm,
						Patch:  client.Apply,
					},
					{
						Object: secret,
						Patch:  client.Apply,
					},
				},
			))
		})
	})

	Describe("ObjectsFromPatchRequests", func() {
		It("should return nil if the requests are nil", func() {
			Expect(ObjectsFromPatchRequests(nil)).To(BeNil())
		})

		It("should retrieve all objects from the patch requests", func() {
			reqs := []PatchRequest{
				{
					Object: cm,
					Patch:  client.Apply,
				},
				{
					Object: secret,
					Patch:  client.Apply,
				},
			}
			Expect(ObjectsFromPatchRequests(reqs)).To(Equal([]client.Object{cm, secret}))
		})
	})

	Describe("PatchMultiple", func() {
		It("should abort and return any error from patching", func() {
			reqs := []PatchRequest{
				{
					Object: cm,
					Patch:  client.Apply,
				},
				{
					Object: secret,
					Patch:  client.Apply,
				},
			}
			someErr := fmt.Errorf("some error")
			c.EXPECT().Patch(ctx, cm, client.Apply).Return(someErr)

			err := PatchMultiple(ctx, c, reqs)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, someErr)).To(BeTrue())
		})

		It("should patch multiple objects", func() {
			reqs := []PatchRequest{
				{
					Object: cm,
					Patch:  client.Apply,
				},
				{
					Object: secret,
					Patch:  client.Apply,
				},
			}
			gomock.InOrder(
				c.EXPECT().Patch(ctx, cm, client.Apply),
				c.EXPECT().Patch(ctx, secret, client.Apply),
			)
			Expect(PatchMultiple(ctx, c, reqs)).To(Succeed())
		})
	})

	Describe("PatchMultipleFromFile", func() {
		It("should error if the file does not exist", func() {
			_, err := PatchMultipleFromFile(ctx, c, "should-not-exist", patchProvider)
			Expect(err).To(HaveOccurred())
		})

		It("should abort and return any error from patching", func() {
			someErr := fmt.Errorf("some error")
			gomock.InOrder(
				patchProvider.EXPECT().PatchFor(testdata.UnstructuredSecret()).Return(client.Apply),
				patchProvider.EXPECT().PatchFor(testdata.UnstructuredConfigMap()).Return(client.Apply),

				c.EXPECT().Patch(ctx, testdata.UnstructuredSecret(), client.Apply).Return(someErr),
			)

			_, err := PatchMultipleFromFile(ctx, c, objectsPath, patchProvider)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, someErr)).To(BeTrue())
		})

		It("should patch multiple objects from file", func() {
			gomock.InOrder(
				patchProvider.EXPECT().PatchFor(testdata.UnstructuredSecret()).Return(client.Apply),
				patchProvider.EXPECT().PatchFor(testdata.UnstructuredConfigMap()).Return(client.Apply),

				c.EXPECT().Patch(ctx, testdata.UnstructuredSecret(), client.Apply),
				c.EXPECT().Patch(ctx, testdata.UnstructuredConfigMap(), client.Apply),
			)

			objs, err := PatchMultipleFromFile(ctx, c, objectsPath, patchProvider)
			Expect(err).NotTo(HaveOccurred())
			Expect(objs).To(Equal([]unstructured.Unstructured{*testdata.UnstructuredSecret(), *testdata.UnstructuredConfigMap()}))
		})
	})

	Describe("DeleteMultiple", func() {
		It("should abort and return any error from deleting", func() {
			someErr := fmt.Errorf("some error")
			c.EXPECT().Delete(ctx, cm).Return(someErr)

			err := DeleteMultiple(ctx, c, []client.Object{cm, secret})
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, someErr)).To(BeTrue())
		})

		It("should delete multiple objects", func() {
			gomock.InOrder(
				c.EXPECT().Delete(ctx, cm),
				c.EXPECT().Delete(ctx, secret),
			)
			Expect(DeleteMultiple(ctx, c, []client.Object{cm, secret})).To(Succeed())
		})
	})

	Describe("DeleteMultipleFromFile", func() {
		It("should error if the file does not exist", func() {
			Expect(DeleteMultipleFromFile(ctx, c, "should-not-exist")).To(HaveOccurred())
		})

		It("should abort and return any error from deleting", func() {
			someErr := fmt.Errorf("some error")
			c.EXPECT().Delete(ctx, testdata.UnstructuredSecret()).Return(someErr)

			err := DeleteMultipleFromFile(ctx, c, objectsPath)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, someErr)).To(BeTrue())
		})

		It("should patch multiple objects from file", func() {
			gomock.InOrder(
				c.EXPECT().Delete(ctx, testdata.UnstructuredSecret()),
				c.EXPECT().Delete(ctx, testdata.UnstructuredConfigMap()),
			)

			Expect(DeleteMultipleFromFile(ctx, c, objectsPath)).To(Succeed())
		})
	})

	Describe("ListAndFilter", func() {
		It("should list and filter the objects", func() {
			cm1 := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: corev1.NamespaceDefault,
					Name:      "foo",
				},
			}
			cm2 := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: corev1.NamespaceDefault,
					Name:      "bar",
				},
			}
			cm3 := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: corev1.NamespaceDefault,
					Name:      "baz",
				},
			}

			list := &corev1.ConfigMapList{}
			c.EXPECT().List(ctx, list).SetArg(1, corev1.ConfigMapList{
				Items: []corev1.ConfigMap{cm1, cm2, cm3},
			})

			Expect(ListAndFilter(ctx, c, list, func(obj client.Object) (bool, error) {
				return strings.HasPrefix(obj.(*corev1.ConfigMap).Name, "b"), nil
			})).To(Succeed())
			Expect(list.Items).To(Equal([]corev1.ConfigMap{cm2, cm3}))
		})
	})

	Describe("ListAndFilterControlledBy", func() {
		It("should list the objects controlled by an owner", func() {
			owner := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "owner",
					UID:       types.UID("owner-uuid"),
				},
			}
			cm1 := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "n1",
				},
			}
			Expect(controllerutil.SetControllerReference(&owner, &cm1, scheme.Scheme)).To(Succeed())
			cm2 := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "n2",
				},
			}
			cm3 := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "n3",
				},
			}
			Expect(controllerutil.SetControllerReference(&owner, &cm3, scheme.Scheme)).To(Succeed())

			list := &corev1.ConfigMapList{}
			gomock.InOrder(
				c.EXPECT().Scheme().Return(scheme.Scheme),
				c.EXPECT().List(ctx, list).SetArg(1, corev1.ConfigMapList{
					Items: []corev1.ConfigMap{cm1, cm2, cm3},
				}),
			)

			Expect(ListAndFilterControlledBy(ctx, c, &owner, list)).To(Succeed())
			Expect(list.Items).To(Equal([]corev1.ConfigMap{cm1, cm3}))
		})
	})

	Describe("CreateOrUse", func() {
		var (
			cm1, cm2, cm3 corev1.ConfigMap
		)
		BeforeEach(func() {
			cm1 = corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "n1",
				},
			}
			cm2 = corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "n2",
				},
			}
			cm3 = corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "n3",
				},
			}
		})

		It("should use an object if it matches", func() {
			cm := &corev1.ConfigMap{}
			res, other, err := CreateOrUse(ctx, c, []client.Object{&cm1, &cm2, &cm3}, cm, func() (bool, error) {
				return cm.Name == "n3", nil
			}, func() error {
				Fail("init should not be called")
				return nil
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(other).To(Equal([]client.Object{&cm1, &cm2}))
			Expect(res).To(Equal(CreateOrUseOperationResultUsed))
			Expect(cm).To(Equal(&cm3))
		})

		It("should create a new object if it does not match", func() {
			cm := &corev1.ConfigMap{}
			c.EXPECT().Create(ctx, cm)
			res, other, err := CreateOrUse(ctx, c, []client.Object{&cm1, &cm2, &cm3}, cm, func() (bool, error) {
				return false, nil
			}, func() error {
				cm.Name = "n4"
				return nil
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(other).To(Equal([]client.Object{&cm1, &cm2, &cm3}))
			Expect(res).To(Equal(CreateOrUseOperationResultCreated))
			Expect(cm).To(Equal(&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "n4",
				},
			}))
		})
	})

	Describe("CreateOrUseWithList", func() {
		var (
			cm1, cm2, cm3 corev1.ConfigMap
		)
		BeforeEach(func() {
			cm1 = corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "n1",
				},
			}
			cm2 = corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "n2",
				},
			}
			cm3 = corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "n3",
				},
			}
		})

		It("should use an object if it matches", func() {
			cm := &corev1.ConfigMap{}
			list := &corev1.ConfigMapList{Items: []corev1.ConfigMap{cm1, cm2, cm3}}
			res, err := CreateOrUseWithList(ctx, c, list, cm, func() (bool, error) {
				return cm.Name == "n3", nil
			}, func() error {
				Fail("init should not be called")
				return nil
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(list.Items).To(Equal([]corev1.ConfigMap{cm1, cm2}))
			Expect(res).To(Equal(CreateOrUseOperationResultUsed))
			Expect(cm).To(Equal(&cm3))
		})

		It("should create a new object if it does not match", func() {
			cm := &corev1.ConfigMap{}
			list := &corev1.ConfigMapList{Items: []corev1.ConfigMap{cm1, cm2, cm3}}
			c.EXPECT().Create(ctx, cm)
			res, err := CreateOrUseWithList(ctx, c, list, cm, func() (bool, error) {
				return false, nil
			}, func() error {
				cm.Name = "n4"
				return nil
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(list.Items).To(Equal([]corev1.ConfigMap{cm1, cm2, cm3}))
			Expect(res).To(Equal(CreateOrUseOperationResultCreated))
			Expect(cm).To(Equal(&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "n4",
				},
			}))
		})
	})

	Describe("CreateOrUseWithObjectSlicePointer", func() {
		var (
			cm1, cm2, cm3 corev1.ConfigMap
			slice         []corev1.ConfigMap
		)
		BeforeEach(func() {
			cm1 = corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "n1",
				},
			}
			cm2 = corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "n2",
				},
			}
			cm3 = corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "n3",
				},
			}
			slice = []corev1.ConfigMap{cm1, cm2, cm3}
		})

		It("should use an object if it matches", func() {
			cm := &corev1.ConfigMap{}
			res, err := CreateOrUseWithObjectSlicePointer(ctx, c, &slice, cm, func() (bool, error) {
				return cm.Name == "n3", nil
			}, func() error {
				Fail("init should not be called")
				return nil
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(slice).To(Equal([]corev1.ConfigMap{cm1, cm2}))
			Expect(res).To(Equal(CreateOrUseOperationResultUsed))
			Expect(cm).To(Equal(&cm3))
		})

		It("should create a new object if it does not match", func() {
			cm := &corev1.ConfigMap{}
			c.EXPECT().Create(ctx, cm)
			res, err := CreateOrUseWithObjectSlicePointer(ctx, c, &slice, cm, func() (bool, error) {
				return false, nil
			}, func() error {
				cm.Name = "n4"
				return nil
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(slice).To(Equal([]corev1.ConfigMap{cm1, cm2, cm3}))
			Expect(res).To(Equal(CreateOrUseOperationResultCreated))
			Expect(cm).To(Equal(&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "n4",
				},
			}))
		})
	})
})
