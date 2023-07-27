// // Copyright 2021 OnMetal authors
// //
// // Licensed under the Apache License, Version 2.0 (the "License");
// // you may not use this file except in compliance with the License.
// // You may obtain a copy of the License at
// //
// //      http://www.apache.org/licenses/LICENSE-2.0
// //
// // Unless required by applicable law or agreed to in writing, software
// // distributed under the License is distributed on an "AS IS" BASIS,
// // WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// // See the License for the specific language governing permissions and
// // limitations under the License.
//

// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/onmetal/controller-utils/clientutils (interfaces: PatchProvider)

// Package clientutils is a generated GoMock package.
package clientutils

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

// MockPatchProvider is a mock of PatchProvider interface.
type MockPatchProvider struct {
	ctrl     *gomock.Controller
	recorder *MockPatchProviderMockRecorder
}

// MockPatchProviderMockRecorder is the mock recorder for MockPatchProvider.
type MockPatchProviderMockRecorder struct {
	mock *MockPatchProvider
}

// NewMockPatchProvider creates a new mock instance.
func NewMockPatchProvider(ctrl *gomock.Controller) *MockPatchProvider {
	mock := &MockPatchProvider{ctrl: ctrl}
	mock.recorder = &MockPatchProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPatchProvider) EXPECT() *MockPatchProviderMockRecorder {
	return m.recorder
}

// PatchFor mocks base method.
func (m *MockPatchProvider) PatchFor(arg0 client.Object) client.Patch {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PatchFor", arg0)
	ret0, _ := ret[0].(client.Patch)
	return ret0
}

// PatchFor indicates an expected call of PatchFor.
func (mr *MockPatchProviderMockRecorder) PatchFor(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PatchFor", reflect.TypeOf((*MockPatchProvider)(nil).PatchFor), arg0)
}
