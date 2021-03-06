// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/manager/manager.go

// Package mock is a generated GoMock package.
package mock

import (
	gomock "github.com/golang/mock/gomock"
	v1 "k8s.io/api/core/v1"
	reflect "reflect"
)

// MockManager is a mock of Manager interface
type MockManager struct {
	ctrl     *gomock.Controller
	recorder *MockManagerMockRecorder
}

// MockManagerMockRecorder is the mock recorder for MockManager
type MockManagerMockRecorder struct {
	mock *MockManager
}

// NewMockManager creates a new mock instance
func NewMockManager(ctrl *gomock.Controller) *MockManager {
	mock := &MockManager{ctrl: ctrl}
	mock.recorder = &MockManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockManager) EXPECT() *MockManagerMockRecorder {
	return m.recorder
}

// Endpoints mocks base method
func (m *MockManager) Endpoints(namespace, name string) (*v1.Endpoints, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Endpoints", namespace, name)
	ret0, _ := ret[0].(*v1.Endpoints)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Endpoints indicates an expected call of Endpoints
func (mr *MockManagerMockRecorder) Endpoints(namespace, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Endpoints", reflect.TypeOf((*MockManager)(nil).Endpoints), namespace, name)
}

// PodsWithLabels mocks base method
func (m *MockManager) PodsWithLabels(namespace, labels string) (*v1.PodList, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PodsWithLabels", namespace, labels)
	ret0, _ := ret[0].(*v1.PodList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PodsWithLabels indicates an expected call of PodsWithLabels
func (mr *MockManagerMockRecorder) PodsWithLabels(namespace, labels interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PodsWithLabels", reflect.TypeOf((*MockManager)(nil).PodsWithLabels), namespace, labels)
}

// EventChan mocks base method
func (m *MockManager) EventChan() <-chan struct{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EventChan")
	ret0, _ := ret[0].(<-chan struct{})
	return ret0
}

// EventChan indicates an expected call of EventChan
func (mr *MockManagerMockRecorder) EventChan() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EventChan", reflect.TypeOf((*MockManager)(nil).EventChan))
}

// ErrorChan mocks base method
func (m *MockManager) ErrorChan() <-chan error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ErrorChan")
	ret0, _ := ret[0].(<-chan error)
	return ret0
}

// ErrorChan indicates an expected call of ErrorChan
func (mr *MockManagerMockRecorder) ErrorChan() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ErrorChan", reflect.TypeOf((*MockManager)(nil).ErrorChan))
}
