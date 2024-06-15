package user

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/julienschmidt/httprouter"
	model "github.com/reversersed/go-web-services/tree/main/api_gateway/internal/client/user"
	mock "github.com/reversersed/go-web-services/tree/main/api_gateway/internal/handlers/user/mocks"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/validator"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestRegister(t *testing.T) {

	var registerCases = []struct {
		Name   string
		Path   string
		Method string
	}{
		{"Authorization", url_auth, http.MethodPost},
		{"Token refresh", url_refresh, http.MethodPost},
		{"Registration", url_register, http.MethodPost},
		{"Email confirmation", url_confirm_email, http.MethodPost},
		{"Find user", url_find_user, http.MethodGet},
		{"Delete user account", url_delete_user, http.MethodDelete},
		{"Update user login", url_update_user_login, http.MethodPatch},
	}

	ctrl := gomock.NewController(t)
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}
	jwt := mock.NewMockJwtService(ctrl)
	h := &Handler{JwtService: jwt, Logger: logger}
	jwt.EXPECT().Middleware(gomock.Any()).AnyTimes()

	router := httprouter.New()
	h.Register(router)
	for _, registerCase := range registerCases {
		t.Run(registerCase.Name, func(t *testing.T) {
			if handler, _, _ := router.Lookup(registerCase.Method, registerCase.Path); handler == nil {
				t.Errorf("handler %s (%s) with method %s not found", registerCase.Name, registerCase.Path, registerCase.Method)
			}
		})
	}
}

func TestUpdateUserLogin(t *testing.T) {
	var testTable = []struct {
		Name           string
		MockBehaviour  func(s *mock.MockUserService, j *mock.MockJwtService)
		InputJson      *model.UpdateUserLoginQuery
		ExceptedStatus int
		ExceptedError  error
		ExceptedBody   string
	}{
		{
			Name: "Successful response",
			MockBehaviour: func(s *mock.MockUserService, j *mock.MockJwtService) {
				usr := &model.User{
					Login: "user",
					Roles: []string{"user", "admin"},
					Email: "user@example.com",
				}
				s.EXPECT().UpdateUserLogin(gomock.Any(), gomock.Any()).Return(
					usr, nil,
				)
				j.EXPECT().GenerateAccessToken(usr).Return(&model.JwtResponse{
					Login:        usr.Login,
					Roles:        usr.Roles,
					Token:        "EXAMPLE TOKEN",
					RefreshToken: "TOKEN",
				}, nil)
			},
			InputJson:      &model.UpdateUserLoginQuery{NewLogin: "user"},
			ExceptedStatus: http.StatusOK,
			ExceptedError:  nil,
			ExceptedBody:   "{\"login\":\"user\",\"roles\":[\"user\",\"admin\"],\"token\":\"EXAMPLE TOKEN\",\"refreshtoken\":\"TOKEN\"}",
		},
		{
			Name:           "Validation error",
			MockBehaviour:  func(s *mock.MockUserService, j *mock.MockJwtService) {},
			InputJson:      &model.UpdateUserLoginQuery{},
			ExceptedStatus: http.StatusNotImplemented,
			ExceptedError:  errors.New("newlogin: field is required"),
			ExceptedBody:   "",
		},
		{
			Name:           "Nil body",
			MockBehaviour:  func(s *mock.MockUserService, j *mock.MockJwtService) {},
			InputJson:      nil,
			ExceptedStatus: http.StatusInternalServerError,
			ExceptedError:  errors.New("EOF"),
			ExceptedBody:   "",
		},
		{
			Name: "User service error",
			MockBehaviour: func(s *mock.MockUserService, j *mock.MockJwtService) {
				s.EXPECT().UpdateUserLogin(gomock.Any(), gomock.Any()).Return(nil, errors.New("error updating login"))
			},
			InputJson:      &model.UpdateUserLoginQuery{NewLogin: "user"},
			ExceptedStatus: http.StatusInternalServerError, //middleware sets 500 when error is not custom
			ExceptedError:  errors.New("error updating login"),
			ExceptedBody:   "",
		},
		{
			Name: "Jwt service error",
			MockBehaviour: func(s *mock.MockUserService, j *mock.MockJwtService) {
				usr := &model.User{
					Login: "user",
					Roles: []string{"user"},
					Email: "user@example.com",
				}
				s.EXPECT().UpdateUserLogin(gomock.Any(), gomock.Any()).Return(
					usr, nil,
				)
				j.EXPECT().GenerateAccessToken(usr).Return(nil, errors.New("jwt service error"))
			},
			InputJson:      &model.UpdateUserLoginQuery{NewLogin: "user"},
			ExceptedStatus: http.StatusInternalServerError, //middleware sets 500 when error is not custom
			ExceptedError:  errors.New("jwt service error"),
			ExceptedBody:   "",
		},
	}

	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}
	for _, tt := range testTable {
		t.Run(tt.Name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			jwt := mock.NewMockJwtService(ctrl)
			user := mock.NewMockUserService(ctrl)
			tt.MockBehaviour(user, jwt)
			h := &Handler{JwtService: jwt, UserService: user, Logger: logger, Validator: validator.New()}

			req, _ := json.Marshal(tt.InputJson)
			w := httptest.NewRecorder()
			var r *http.Request
			if tt.InputJson != nil {
				r = httptest.NewRequest(http.MethodPatch, "http://test", bytes.NewBuffer(req))
			} else {
				r = httptest.NewRequest(http.MethodPatch, "http://test", nil)
			}
			handler := errormiddleware.Middleware(h.UpdateUserLogin)
			err := handler(w, r)
			if tt.ExceptedStatus != w.Result().StatusCode {
				t.Fatalf("excepeted status code %d but got %d", tt.ExceptedStatus, w.Result().StatusCode)
			}
			if tt.ExceptedError != nil && err == nil {
				t.Fatalf("excepeted error but got nil")
			} else if tt.ExceptedError != nil && err != nil {
				if tt.ExceptedError.Error() != err.Error() {
					t.Fatalf("excepted error %v but got %v", tt.ExceptedError, err)
				}
				return
			}
			if tt.ExceptedError == nil && err != nil {
				t.Fatalf("excepted error nil but got %v", err)
			}

			body := w.Body.String()
			if len(tt.ExceptedBody) == 0 && len(body) != 0 {
				t.Fatalf("excepted body to be nil but got %s", body)
			}
			if tt.ExceptedBody != body {
				t.Fatalf("excepted body %s but got %s", tt.ExceptedBody, body)
			}
		})
	}
}
