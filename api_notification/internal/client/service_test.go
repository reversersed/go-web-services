package client_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/reversersed/go-web-services/tree/main/api_notification/internal/client"
	mock "github.com/reversersed/go-web-services/tree/main/api_notification/internal/client/mocks"
	"github.com/reversersed/go-web-services/tree/main/api_notification/internal/config"
	cache_mock "github.com/reversersed/go-web-services/tree/main/api_notification/pkg/cache/mocks"
	"github.com/reversersed/go-web-services/tree/main/api_notification/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_notification/pkg/logging"
	"github.com/reversersed/go-web-services/tree/main/api_notification/pkg/validator"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestSendNotification(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log, hook := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}
	storage := mock.NewMockStorage(ctrl)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	cache := cache_mock.NewMockCache(ctrl)
	service := client.NewService(storage, logger, cache, validator.New(), &config.UrlConfig{Url_User_Service: server.URL})

	caseTable := []struct {
		Name           string
		Handler        http.HandlerFunc
		MockBehaviour  func(s *mock.MockStorage, c *cache_mock.MockCache)
		ExceptedOutput string
		Model          *client.SendNotificationMessage
	}{
		{
			Name:           "Nil body",
			Handler:        nil,
			MockBehaviour:  func(s *mock.MockStorage, c *cache_mock.MockCache) {},
			ExceptedOutput: "received nil query",
			Model:          nil,
		},
		{
			Name:           "Wrong query",
			Handler:        nil,
			MockBehaviour:  func(s *mock.MockStorage, c *cache_mock.MockCache) {},
			ExceptedOutput: "Error code: IE-0004, Error: userid: field is required, content: field is required, type: field is required, Dev message: received wrong notification query",
			Model:          &client.SendNotificationMessage{},
		},
		{
			Name:    "Storage error",
			Handler: nil,
			MockBehaviour: func(s *mock.MockStorage, c *cache_mock.MockCache) {
				c.EXPECT().Get([]byte("57bf425a34ce5ee85891b914")).Return(nil, errors.New("not found"))
				s.EXPECT().IsUserExists(gomock.Any(), "57bf425a34ce5ee85891b914").Return(false, errors.New("error storage"))
			},
			ExceptedOutput: "error storage",
			Model: &client.SendNotificationMessage{
				UserId:  "57bf425a34ce5ee85891b914",
				Content: "content notification",
				Type:    client.Info,
			},
		},
		{
			Name: "Error response for user",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write(errormiddleware.NotFoundError([]string{"not found"}, "not found").Marshall())
			},
			MockBehaviour: func(s *mock.MockStorage, c *cache_mock.MockCache) {
				c.EXPECT().Get([]byte("57bf425a34ce5ee85891b914")).Return(nil, errors.New("not found"))
				s.EXPECT().IsUserExists(gomock.Any(), "57bf425a34ce5ee85891b914").Return(false, nil)
			},
			ExceptedOutput: errormiddleware.NotFoundError([]string{"not found"}, "not found").Error(),
			Model: &client.SendNotificationMessage{
				UserId:  "57bf425a34ce5ee85891b914",
				Content: "content notification",
				Type:    client.Info,
			},
		},
		{
			Name: "Bad response",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			ExceptedOutput: "EOF",
			MockBehaviour: func(s *mock.MockStorage, c *cache_mock.MockCache) {
				c.EXPECT().Get([]byte("57bf425a34ce5ee85891b914")).Return(nil, errors.New("not found"))
				s.EXPECT().IsUserExists(gomock.Any(), "57bf425a34ce5ee85891b914").Return(false, nil)
			},
			Model: &client.SendNotificationMessage{
				UserId:  "57bf425a34ce5ee85891b914",
				Content: "content notification",
				Type:    client.Info,
			},
		},
		{
			Name: "User creation error",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				user := struct {
					Login string `json:"login"`
				}{"username"}
				u, _ := json.Marshal(&user)
				w.Write(u)
			},
			ExceptedOutput: "Error while creating user: wrong request",
			MockBehaviour: func(s *mock.MockStorage, c *cache_mock.MockCache) {
				c.EXPECT().Get([]byte("57bf425a34ce5ee85891b914")).Return(nil, errors.New("not found"))
				s.EXPECT().IsUserExists(gomock.Any(), "57bf425a34ce5ee85891b914").Return(false, nil)
				s.EXPECT().CreateUser(gomock.Any(), "57bf425a34ce5ee85891b914", "username").Return(errors.New("wrong request"))
			},
			Model: &client.SendNotificationMessage{
				UserId:  "57bf425a34ce5ee85891b914",
				Content: "content notification",
				Type:    client.Info,
			},
		},
		{
			Name: "Send notification error",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				user := struct {
					Login string `json:"login"`
				}{"username"}
				u, _ := json.Marshal(&user)
				w.Write(u)
			},
			ExceptedOutput: "Error sending notification: wrong request",
			MockBehaviour: func(s *mock.MockStorage, c *cache_mock.MockCache) {
				c.EXPECT().Get([]byte("57bf425a34ce5ee85891b914")).Return(nil, errors.New("not found"))
				s.EXPECT().IsUserExists(gomock.Any(), "57bf425a34ce5ee85891b914").Return(false, nil)
				s.EXPECT().CreateUser(gomock.Any(), "57bf425a34ce5ee85891b914", "username").Return(nil)
				s.EXPECT().SendNotification(gomock.Any(), gomock.Any(), "57bf425a34ce5ee85891b914").Return(errors.New("wrong request"))
			},
			Model: &client.SendNotificationMessage{
				UserId:  "57bf425a34ce5ee85891b914",
				Content: "content notification",
				Type:    client.Info,
			},
		},
		{
			Name:           "Successful user exist",
			Handler:        nil,
			ExceptedOutput: "Notification info sended to user 57bf425a34ce5ee85891b914 (Content: content notification)",
			MockBehaviour: func(s *mock.MockStorage, c *cache_mock.MockCache) {
				c.EXPECT().Get([]byte("57bf425a34ce5ee85891b914")).Return([]byte(""), nil)
				s.EXPECT().SendNotification(gomock.Any(), gomock.Any(), "57bf425a34ce5ee85891b914").Return(nil)
				c.EXPECT().Set([]byte("57bf425a34ce5ee85891b914"), []byte(""), int(time.Hour))
			},
			Model: &client.SendNotificationMessage{
				UserId:  "57bf425a34ce5ee85891b914",
				Content: "content notification",
				Type:    client.Info,
			},
		},
	}
	for _, tt := range caseTable {
		t.Run(tt.Name, func(t *testing.T) {
			server.Config.Handler = tt.Handler
			tt.MockBehaviour(storage, cache)
			service.SendNotification(context.Background(), tt.Model)
			assert.Equal(t, tt.ExceptedOutput, hook.LastEntry().Message)
		})
	}
}

func TestUserDeleted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log, hook := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}
	storage := mock.NewMockStorage(ctrl)
	cache := cache_mock.NewMockCache(ctrl)
	service := client.NewService(storage, logger, cache, validator.New(), &config.UrlConfig{Url_User_Service: ""})

	storage.EXPECT().DeleteUser(gomock.Any(), "userid").Return(nil)
	service.OnUserDeleted(context.Background(), "userid")
	assert.Nil(t, hook.LastEntry())

	storage.EXPECT().DeleteUser(gomock.Any(), "userid").Return(errors.New("error handled"))
	service.OnUserDeleted(context.Background(), "userid")
	if assert.NotNil(t, hook.LastEntry()) {
		assert.Equal(t, "error handled", hook.LastEntry().Message)
	}
}
func TestUserLoginChanged(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log, hook := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}
	storage := mock.NewMockStorage(ctrl)
	cache := cache_mock.NewMockCache(ctrl)
	service := client.NewService(storage, logger, cache, validator.New(), &config.UrlConfig{Url_User_Service: ""})

	storage.EXPECT().ChangeUserLogin(gomock.Any(), "57bf425a34ce5ee85891b914", "user").Return(nil)
	service.OnUserLoginChanged(context.Background(), &client.UserLoginChangedMessage{
		UserId:   "57bf425a34ce5ee85891b914",
		NewLogin: "user",
	})
	assert.Nil(t, hook.LastEntry())

	service.OnUserLoginChanged(context.Background(), &client.UserLoginChangedMessage{})
	if assert.NotNil(t, hook.LastEntry()) {
		assert.Equal(t, "received wrong user login changed query: Error code: IE-0004, Error: userid: field is required, newlogin: field is required, Dev message: ", hook.LastEntry().Message)
	}

	storage.EXPECT().ChangeUserLogin(gomock.Any(), "57bf425a34ce5ee85891b914", "user").Return(errors.New("error"))
	service.OnUserLoginChanged(context.Background(), &client.UserLoginChangedMessage{
		UserId:   "57bf425a34ce5ee85891b914",
		NewLogin: "user",
	})
	if assert.NotNil(t, hook.LastEntry()) {
		assert.Equal(t, "error", hook.LastEntry().Message)
	}
}
