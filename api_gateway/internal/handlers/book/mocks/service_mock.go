// Code generated by MockGen. DO NOT EDIT.
// Source: handler.go

// Package mock_book is a generated GoMock package.
package mock_book

import (
	context "context"
	io "io"
	http "net/http"
	url "net/url"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	book "github.com/reversersed/go-web-services/tree/main/api_gateway/internal/client/book"
)

// MockBookService is a mock of BookService interface.
type MockBookService struct {
	ctrl     *gomock.Controller
	recorder *MockBookServiceMockRecorder
}

// MockBookServiceMockRecorder is the mock recorder for MockBookService.
type MockBookServiceMockRecorder struct {
	mock *MockBookService
}

// NewMockBookService creates a new mock instance.
func NewMockBookService(ctrl *gomock.Controller) *MockBookService {
	mock := &MockBookService{ctrl: ctrl}
	mock.recorder = &MockBookServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBookService) EXPECT() *MockBookServiceMockRecorder {
	return m.recorder
}

// AddBook mocks base method.
func (m *MockBookService) AddBook(ctx context.Context, body io.Reader, contentType string) (*book.Book, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddBook", ctx, body, contentType)
	ret0, _ := ret[0].(*book.Book)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddBook indicates an expected call of AddBook.
func (mr *MockBookServiceMockRecorder) AddBook(ctx, body, contentType interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddBook", reflect.TypeOf((*MockBookService)(nil).AddBook), ctx, body, contentType)
}

// FindBooks mocks base method.
func (m *MockBookService) FindBooks(ctx context.Context, params url.Values) ([]*book.Book, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindBooks", ctx, params)
	ret0, _ := ret[0].([]*book.Book)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindBooks indicates an expected call of FindBooks.
func (mr *MockBookServiceMockRecorder) FindBooks(ctx, params interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindBooks", reflect.TypeOf((*MockBookService)(nil).FindBooks), ctx, params)
}

// GetBook mocks base method.
func (m *MockBookService) GetBook(ctx context.Context, id string) (*book.Book, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBook", ctx, id)
	ret0, _ := ret[0].(*book.Book)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBook indicates an expected call of GetBook.
func (mr *MockBookServiceMockRecorder) GetBook(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBook", reflect.TypeOf((*MockBookService)(nil).GetBook), ctx, id)
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
