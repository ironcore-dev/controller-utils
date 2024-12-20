// // SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// // SPDX-License-Identifier: Apache-2.0
//

// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/ironcore-dev/controller-utils/mock/controller-runtime/client (interfaces: IndexerFunc)
//
// Generated by this command:
//
//	mockgen -copyright_file ../../../hack/boilerplate.go.txt -package client -destination funcs.go github.com/ironcore-dev/controller-utils/mock/controller-runtime/client IndexerFunc
//

// Package client is a generated GoMock package.
package client

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

// MockIndexerFunc is a mock of IndexerFunc interface.
type MockIndexerFunc struct {
	ctrl     *gomock.Controller
	recorder *MockIndexerFuncMockRecorder
	isgomock struct{}
}

// MockIndexerFuncMockRecorder is the mock recorder for MockIndexerFunc.
type MockIndexerFuncMockRecorder struct {
	mock *MockIndexerFunc
}

// NewMockIndexerFunc creates a new mock instance.
func NewMockIndexerFunc(ctrl *gomock.Controller) *MockIndexerFunc {
	mock := &MockIndexerFunc{ctrl: ctrl}
	mock.recorder = &MockIndexerFuncMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIndexerFunc) EXPECT() *MockIndexerFuncMockRecorder {
	return m.recorder
}

// Call mocks base method.
func (m *MockIndexerFunc) Call(object client.Object) []string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Call", object)
	ret0, _ := ret[0].([]string)
	return ret0
}

// Call indicates an expected call of Call.
func (mr *MockIndexerFuncMockRecorder) Call(object any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Call", reflect.TypeOf((*MockIndexerFunc)(nil).Call), object)
}
