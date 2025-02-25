// // SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// // SPDX-License-Identifier: Apache-2.0
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
	isgomock struct{}
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
func (m *MockEachListItemFunc) Call(obj client.Object) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Call", obj)
	ret0, _ := ret[0].(error)
	return ret0
}

// Call indicates an expected call of Call.
func (mr *MockEachListItemFuncMockRecorder) Call(obj any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Call", reflect.TypeOf((*MockEachListItemFunc)(nil).Call), obj)
}
