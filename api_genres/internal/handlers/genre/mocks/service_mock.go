// Code generated by MockGen. DO NOT EDIT.
// Source: handler.go

// Package mock_genre is a generated GoMock package.
package mock_genre

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	client "github.com/reversersed/go-web-services/tree/main/api_genres/internal/client"
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

// AddGenre mocks base method.
func (m *MockService) AddGenre(ctx context.Context, genre *client.AddGenreQuery) (*client.Genre, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddGenre", ctx, genre)
	ret0, _ := ret[0].(*client.Genre)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddGenre indicates an expected call of AddGenre.
func (mr *MockServiceMockRecorder) AddGenre(ctx, genre interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddGenre", reflect.TypeOf((*MockService)(nil).AddGenre), ctx, genre)
}

// GetAllGenres mocks base method.
func (m *MockService) GetAllGenres(ctx context.Context) ([]*client.Genre, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllGenres", ctx)
	ret0, _ := ret[0].([]*client.Genre)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllGenres indicates an expected call of GetAllGenres.
func (mr *MockServiceMockRecorder) GetAllGenres(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllGenres", reflect.TypeOf((*MockService)(nil).GetAllGenres), ctx)
}

// GetGenre mocks base method.
func (m *MockService) GetGenre(ctx context.Context, id string) ([]*client.Genre, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGenre", ctx, id)
	ret0, _ := ret[0].([]*client.Genre)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetGenre indicates an expected call of GetGenre.
func (mr *MockServiceMockRecorder) GetGenre(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGenre", reflect.TypeOf((*MockService)(nil).GetGenre), ctx, id)
}
