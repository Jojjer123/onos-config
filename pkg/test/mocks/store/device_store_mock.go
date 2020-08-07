// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/onosproject/onos-config/pkg/store/device (interfaces: Store)

// Package mock_device is a generated GoMock package.
package store

import (
	gomock "github.com/golang/mock/gomock"
	device "github.com/onosproject/onos-topo/api/device"
	reflect "reflect"
)

// MockDeviceStore is a mock of Store interface
type MockDeviceStore struct {
	ctrl     *gomock.Controller
	recorder *MockDeviceStoreMockRecorder
}

// MockDeviceStoreMockRecorder is the mock recorder for MockDeviceStore
type MockDeviceStoreMockRecorder struct {
	mock *MockDeviceStore
}

// NewMockDeviceStore creates a new mock instance
func NewMockDeviceStore(ctrl *gomock.Controller) *MockDeviceStore {
	mock := &MockDeviceStore{ctrl: ctrl}
	mock.recorder = &MockDeviceStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDeviceStore) EXPECT() *MockDeviceStoreMockRecorder {
	return m.recorder
}

// Get mocks base method
func (m *MockDeviceStore) Get(arg0 device.ID) (*device.Device, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0)
	ret0, _ := ret[0].(*device.Device)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get
func (mr *MockDeviceStoreMockRecorder) Get(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockDeviceStore)(nil).Get), arg0)
}

// List mocks base method
func (m *MockDeviceStore) List(arg0 chan<- *device.Device) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// List indicates an expected call of List
func (mr *MockDeviceStoreMockRecorder) List(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockDeviceStore)(nil).List), arg0)
}

// Update mocks base method
func (m *MockDeviceStore) Update(arg0 *device.Device) (*device.Device, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", arg0)
	ret0, _ := ret[0].(*device.Device)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Update indicates an expected call of Update
func (mr *MockDeviceStoreMockRecorder) Update(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockDeviceStore)(nil).Update), arg0)
}

// Watch mocks base method
func (m *MockDeviceStore) Watch(arg0 chan<- *device.ListResponse) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Watch", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Watch indicates an expected call of Watch
func (mr *MockDeviceStoreMockRecorder) Watch(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Watch", reflect.TypeOf((*MockDeviceStore)(nil).Watch), arg0)
}
