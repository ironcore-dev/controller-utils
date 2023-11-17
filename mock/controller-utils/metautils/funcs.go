// // Copyright 2021 IronCore authors
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
// Source: github.com/ironcore-dev/controller-utils/mock/controller-utils/metautils (interfaces: EachListItemFunc)
//
// Generated by this command:
//
//	mockgen -copyright_file ../../../hack/boilerplate.go.txt -package metautils -destination=funcs.go github.com/ironcore-dev/controller-utils/mock/controller-utils/metautils EachListItemFunc
//
// Package metautils is a generated GoMock package.
package metautils

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

// MockEachListItemFunc is a mock of EachListItemFunc interface.
type MockEachListItemFunc struct {
	ctrl     *gomock.Controller
	recorder *MockEachListItemFuncMockRecorder
}

// MockEachListItemFuncMockRecorder is the mock recorder for MockEachListItemFunc.
type MockEachListItemFuncMockRecorder struct {
	mock *MockEachListItemFunc
}

// NewMockEachListItemFunc creates a new mock instance.
func NewMockEachListItemFunc(ctrl *gomock.Controller) *MockEachListItemFunc {
	mock := &MockEachListItemFunc{ctrl: ctrl}
	mock.recorder = &MockEachListItemFuncMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEachListItemFunc) EXPECT() *MockEachListItemFuncMockRecorder {
	return m.recorder
}

// Call mocks base method.
func (m *MockEachListItemFunc) Call(arg0 client.Object) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Call", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Call indicates an expected call of Call.
func (mr *MockEachListItemFuncMockRecorder) Call(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Call", reflect.TypeOf((*MockEachListItemFunc)(nil).Call), arg0)
}
