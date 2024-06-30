// Code generated by MockGen. DO NOT EDIT.
// Source: handler.go

// Package mock_genre is a generated GoMock package.
package mock_genre

import (
	context "context"
	http "net/http"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockService is a mock of Service interface.
type MockService struct {
	ctrl     *gomock.Controller
	recorder *MockServiceMockRecorder
}

// MockServiceMockRecorder is the mock recorder for MockService.
type MockServiceMockRecorder struct {
	mock *MockService
}

// NewMockService creates a new mock instance.
func NewMockService(ctrl *gomock.Controller) *MockService {
	mock := &MockService{ctrl: ctrl}
	mock.recorder = &MockServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockService) EXPECT() *MockServiceMockRecorder {
	return m.recorder
}

// SendGetGeneric mocks base method.
func (m *MockService) SendGetGeneric(ctx context.Context, path string, params map[string][]string) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendGetGeneric", ctx, path, params)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SendGetGeneric indicates an expected call of SendGetGeneric.
func (mr *MockServiceMockRecorder) SendGetGeneric(ctx, path, params interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendGetGeneric", reflect.TypeOf((*MockService)(nil).SendGetGeneric), ctx, path, params)
}

// SendPostGeneric mocks base method.
func (m *MockService) SendPostGeneric(ctx context.Context, path string, body []byte) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendPostGeneric", ctx, path, body)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SendPostGeneric indicates an expected call of SendPostGeneric.
func (mr *MockServiceMockRecorder) SendPostGeneric(ctx, path, body interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendPostGeneric", reflect.TypeOf((*MockService)(nil).SendPostGeneric), ctx, path, body)
}

// MockJwtService is a mock of JwtService interface.
type MockJwtService struct {
	ctrl     *gomock.Controller
	recorder *MockJwtServiceMockRecorder
}

// MockJwtServiceMockRecorder is the mock recorder for MockJwtService.
type MockJwtServiceMockRecorder struct {
	mock *MockJwtService
}

// NewMockJwtService creates a new mock instance.
func NewMockJwtService(ctrl *gomock.Controller) *MockJwtService {
	mock := &MockJwtService{ctrl: ctrl}
	mock.recorder = &MockJwtServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockJwtService) EXPECT() *MockJwtServiceMockRecorder {
	return m.recorder
}

// Middleware mocks base method.
func (m *MockJwtService) Middleware(h http.HandlerFunc, roles ...string) http.HandlerFunc {
	m.ctrl.T.Helper()
	varargs := []interface{}{h}
	for _, a := range roles {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Middleware", varargs...)
	ret0, _ := ret[0].(http.HandlerFunc)
	return ret0
}

// Middleware indicates an expected call of Middleware.
func (mr *MockJwtServiceMockRecorder) Middleware(h interface{}, roles ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{h}, roles...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Middleware", reflect.TypeOf((*MockJwtService)(nil).Middleware), varargs...)
}
