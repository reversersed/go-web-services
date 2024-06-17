// Code generated by MockGen. DO NOT EDIT.
// Source: notification.go

// Package mock_receivers is a generated GoMock package.
package mock_receivers

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	client "github.com/reversersed/go-web-services/tree/main/api_notification/internal/client"
)

// Mocknotification_service is a mock of notification_service interface.
type Mocknotification_service struct {
	ctrl     *gomock.Controller
	recorder *Mocknotification_serviceMockRecorder
}

// Mocknotification_serviceMockRecorder is the mock recorder for Mocknotification_service.
type Mocknotification_serviceMockRecorder struct {
	mock *Mocknotification_service
}

// NewMocknotification_service creates a new mock instance.
func NewMocknotification_service(ctrl *gomock.Controller) *Mocknotification_service {
	mock := &Mocknotification_service{ctrl: ctrl}
	mock.recorder = &Mocknotification_serviceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Mocknotification_service) EXPECT() *Mocknotification_serviceMockRecorder {
	return m.recorder
}

// SendNotification mocks base method.
func (m *Mocknotification_service) SendNotification(ctx context.Context, query *client.SendNotificationMessage) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendNotification", ctx, query)
}

// SendNotification indicates an expected call of SendNotification.
func (mr *Mocknotification_serviceMockRecorder) SendNotification(ctx, query interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendNotification", reflect.TypeOf((*Mocknotification_service)(nil).SendNotification), ctx, query)
}
