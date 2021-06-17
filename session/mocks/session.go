// Code generated by MockGen. DO NOT EDIT.
// Source: session.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	protos "github.com/tutumagi/pitaya/protos"
	net "net"
	reflect "reflect"
)

// MockNetworkEntity is a mock of NetworkEntity interface
type MockNetworkEntity struct {
	ctrl     *gomock.Controller
	recorder *MockNetworkEntityMockRecorder
}

// MockNetworkEntityMockRecorder is the mock recorder for MockNetworkEntity
type MockNetworkEntityMockRecorder struct {
	mock *MockNetworkEntity
}

// NewMockNetworkEntity creates a new mock instance
func NewMockNetworkEntity(ctrl *gomock.Controller) *MockNetworkEntity {
	mock := &MockNetworkEntity{ctrl: ctrl}
	mock.recorder = &MockNetworkEntityMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockNetworkEntity) EXPECT() *MockNetworkEntityMockRecorder {
	return m.recorder
}

// Push mocks base method
func (m *MockNetworkEntity) Push(route string, v interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Push", route, v)
	ret0, _ := ret[0].(error)
	return ret0
}

// Push indicates an expected call of Push
func (mr *MockNetworkEntityMockRecorder) Push(route, v interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Push", reflect.TypeOf((*MockNetworkEntity)(nil).Push), route, v)
}

// ResponseMID mocks base method
func (m *MockNetworkEntity) ResponseMID(ctx context.Context, mid uint, v interface{}, isError ...bool) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, mid, v}
	for _, a := range isError {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ResponseMID", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// ResponseMID indicates an expected call of ResponseMID
func (mr *MockNetworkEntityMockRecorder) ResponseMID(ctx, mid, v interface{}, isError ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, mid, v}, isError...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ResponseMID", reflect.TypeOf((*MockNetworkEntity)(nil).ResponseMID), varargs...)
}

// Close mocks base method
func (m *MockNetworkEntity) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close
func (mr *MockNetworkEntityMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockNetworkEntity)(nil).Close))
}

// Kick mocks base method
func (m *MockNetworkEntity) Kick(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Kick", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Kick indicates an expected call of Kick
func (mr *MockNetworkEntityMockRecorder) Kick(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Kick", reflect.TypeOf((*MockNetworkEntity)(nil).Kick), ctx)
}

// RemoteAddr mocks base method
func (m *MockNetworkEntity) RemoteAddr() net.Addr {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoteAddr")
	ret0, _ := ret[0].(net.Addr)
	return ret0
}

// RemoteAddr indicates an expected call of RemoteAddr
func (mr *MockNetworkEntityMockRecorder) RemoteAddr() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoteAddr", reflect.TypeOf((*MockNetworkEntity)(nil).RemoteAddr))
}

// SendRequest mocks base method
func (m *MockNetworkEntity) SendRequest(ctx context.Context, entityID, entityType, serverID, route string, v interface{}) (*protos.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendRequest", ctx, entityID, entityType, serverID, route, v)
	ret0, _ := ret[0].(*protos.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SendRequest indicates an expected call of SendRequest
func (mr *MockNetworkEntityMockRecorder) SendRequest(ctx, entityID, entityType, serverID, route, v interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendRequest", reflect.TypeOf((*MockNetworkEntity)(nil).SendRequest), ctx, entityID, entityType, serverID, route, v)
}
