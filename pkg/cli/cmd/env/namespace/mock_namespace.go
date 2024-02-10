// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/radius-project/radius/pkg/cli/cmd/env/namespace (interfaces: Interface)

// Package namespace is a generated GoMock package.
package namespace

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	workspaces "github.com/radius-project/radius/pkg/cli/workspaces"
)

// MockInterface is a mock of Interface interface.
type MockInterface struct {
	ctrl     *gomock.Controller
	recorder *MockInterfaceMockRecorder
}

// MockInterfaceMockRecorder is the mock recorder for MockInterface.
type MockInterfaceMockRecorder struct {
	mock *MockInterface
}

// NewMockInterface creates a new mock instance.
func NewMockInterface(ctrl *gomock.Controller) *MockInterface {
	mock := &MockInterface{ctrl: ctrl}
	mock.recorder = &MockInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockInterface) EXPECT() *MockInterfaceMockRecorder {
	return m.recorder
}

// ValidateNamespace mocks base method.
func (m *MockInterface) ValidateNamespace(arg0 context.Context, arg1 string, arg2 workspaces.Workspace) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateNamespace", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// ValidateNamespace indicates an expected call of ValidateNamespace.
func (mr *MockInterfaceMockRecorder) ValidateNamespace(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateNamespace", reflect.TypeOf((*MockInterface)(nil).ValidateNamespace), arg0, arg1, arg2)
}
