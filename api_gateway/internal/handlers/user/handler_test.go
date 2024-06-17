package user

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
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
	"github.com/stretchr/testify/assert"
)

var h *Handler

func TestMain(m *testing.M) {
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}
	h = &Handler{Logger: logger, Validator: validator.New()}

	os.Exit(m.Run())
}
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
	jwt := mock.NewMockJwtService(ctrl)
	h.JwtService = jwt
	jwt.EXPECT().Middleware(gomock.Any()).AnyTimes()

	router := httprouter.New()
	h.Register(router)
	for _, registerCase := range registerCases {
		t.Run(registerCase.Name, func(t *testing.T) {
			handler, _, _ := router.Lookup(registerCase.Method, registerCase.Path)
			assert.NotNil(t, handler, "handler %s (%s) with method %s not found", registerCase.Name, registerCase.Path, registerCase.Method)
		})
	}
}

func TestHandlers(t *testing.T) {
	type handlerOptions struct {
		Name           string
		MockBehaviour  func(s *mock.MockUserService, j *mock.MockJwtService)
		InputJson      func() *[]byte
		ExceptedStatus int
		ExceptedError  error
		ExceptedBody   string
	}
	var testTable = []struct {
		HandlerName string
		Handler     func(w http.ResponseWriter, r *http.Request) error
		Method      string
		Options     []handlerOptions
	}{
		//UpdateUserLogin
		{
			HandlerName: "UpdateUserLogin",
			Handler:     h.UpdateUserLogin,
			Method:      http.MethodPatch,
			Options: []handlerOptions{
				//Successful
				{
					Name: "success",
					MockBehaviour: func(s *mock.MockUserService, j *mock.MockJwtService) {
						usr := &model.User{
							Login: "user",
							Roles: []string{"user", "admin"},
							Email: "user@example.com",
						}
						gomock.InOrder(
							s.EXPECT().UpdateUserLogin(gomock.Any(), gomock.Any()).Return(
								usr, nil,
							),
							j.EXPECT().GenerateAccessToken(usr).Return(&model.JwtResponse{
								Login:        usr.Login,
								Roles:        usr.Roles,
								Token:        "EXAMPLE TOKEN",
								RefreshToken: "TOKEN",
							}, nil),
						)
					},
					InputJson: func() *[]byte {
						byt, _ := json.Marshal(&model.UpdateUserLoginQuery{NewLogin: "user"})
						return &byt
					},
					ExceptedStatus: http.StatusOK,
					ExceptedBody:   "{\"login\":\"user\",\"roles\":[\"user\",\"admin\"],\"token\":\"EXAMPLE TOKEN\",\"refreshtoken\":\"TOKEN\"}",
				},
				//Validation error
				{
					Name: "validation",
					InputJson: func() *[]byte {
						byt, _ := json.Marshal(&model.UpdateUserLoginQuery{})
						return &byt
					},
					ExceptedStatus: http.StatusNotImplemented,
					ExceptedError:  errormiddleware.NewError([]string{"newlogin: field is required"}, errormiddleware.ValidationErrorCode, "wrong request"),
					ExceptedBody:   "{\"messages\":[\"newlogin: field is required\"],\"dev_message\":\"wrong request\",\"code\":\"IE-0004\"}",
				},
				//Nil body
				{
					Name:           "nil body",
					ExceptedStatus: http.StatusInternalServerError,
					ExceptedError:  errors.New("EOF"),
					ExceptedBody:   "{\"messages\":[\"EOF\"],\"dev_message\":\"Something wrong happened while service executing\",\"code\":\"IE-0001\"}",
				},
				//User service error return
				{
					Name: "user service returned error",
					MockBehaviour: func(s *mock.MockUserService, j *mock.MockJwtService) {
						s.EXPECT().UpdateUserLogin(gomock.Any(), gomock.Any()).Return(nil, errors.New("error updating login"))
					},
					InputJson: func() *[]byte {
						byt, _ := json.Marshal(&model.UpdateUserLoginQuery{NewLogin: "user"})
						return &byt
					},
					ExceptedStatus: http.StatusInternalServerError, //middleware sets 500 when error is not custom
					ExceptedError:  errors.New("error updating login"),
					ExceptedBody:   "{\"messages\":[\"error updating login\"],\"dev_message\":\"Something wrong happened while service executing\",\"code\":\"IE-0001\"}",
				},
				//Jwt service error return
				{
					Name: "jwt service returned error",
					MockBehaviour: func(s *mock.MockUserService, j *mock.MockJwtService) {
						usr := &model.User{
							Login: "user",
							Roles: []string{"user"},
							Email: "user@example.com",
						}

						gomock.InOrder(
							s.EXPECT().UpdateUserLogin(gomock.Any(), gomock.Any()).Return(
								usr, nil,
							),
							j.EXPECT().GenerateAccessToken(usr).Return(nil, errors.New("jwt service error")),
						)
					},
					InputJson: func() *[]byte {
						byt, _ := json.Marshal(&model.UpdateUserLoginQuery{NewLogin: "user"})
						return &byt
					},
					ExceptedStatus: http.StatusInternalServerError, //middleware sets 500 when error is not custom
					ExceptedError:  errors.New("jwt service error"),
					ExceptedBody:   "{\"messages\":[\"jwt service error\"],\"dev_message\":\"Something wrong happened while service executing\",\"code\":\"IE-0001\"}",
				},
			},
		},
		//DeleteUser
		{
			HandlerName: "DeleteUser",
			Handler:     h.DeleteUser,
			Method:      http.MethodDelete,
			Options: []handlerOptions{
				//Successful
				{
					Name: "success",
					MockBehaviour: func(s *mock.MockUserService, j *mock.MockJwtService) {
						s.EXPECT().DeleteUser(gomock.Any(), gomock.Any()).Return(nil)
					},
					ExceptedStatus: http.StatusNoContent,
					InputJson: func() *[]byte {
						byte, _ := json.Marshal(&model.DeleteUserQuery{
							Password: "password",
						})
						return &byte
					},
				},
				//Nil body
				{
					Name:           "nil body",
					ExceptedStatus: http.StatusInternalServerError,
					ExceptedError:  errors.New("EOF"),
					ExceptedBody:   `{"messages":["EOF"],"dev_message":"Something wrong happened while service executing","code":"IE-0001"}`,
				},
				//Validation error returned
				{
					Name: "validation error",
					InputJson: func() *[]byte {
						byte, _ := json.Marshal(&model.DeleteUserQuery{})
						return &byte
					},
					ExceptedStatus: http.StatusNotImplemented,
					ExceptedError:  errormiddleware.NewError([]string{"password: field is required"}, errormiddleware.ValidationErrorCode, "wrong request"),
					ExceptedBody:   `{"messages":["password: field is required"],"dev_message":"wrong request","code":"IE-0004"}`,
				},
				//Service error returned
				{
					Name: "service error returned",
					MockBehaviour: func(s *mock.MockUserService, j *mock.MockJwtService) {
						s.EXPECT().DeleteUser(gomock.Any(), gomock.Any()).Return(errors.New("wrong password"))
					},
					InputJson: func() *[]byte {
						byte, _ := json.Marshal(&model.DeleteUserQuery{
							Password: "password",
						})
						return &byte
					},
					ExceptedStatus: http.StatusInternalServerError,
					ExceptedError:  errors.New("wrong password"),
					ExceptedBody:   `{"messages":["wrong password"],"dev_message":"Something wrong happened while service executing","code":"IE-0001"}`,
				},
			},
		},
		//EmailConfirmation
		{
			HandlerName: "EmailConfirmation",
			Handler:     h.EmailConfirmation,
			Method:      http.MethodPost,
			Options: []handlerOptions{
				//Successful 200 code
				{
					Name: "success 200 code",
					MockBehaviour: func(s *mock.MockUserService, j *mock.MockJwtService) {
						s.EXPECT().UserEmailConfirmation(gomock.Any(), gomock.Any()).Return(http.StatusOK, nil)
					},
					ExceptedStatus: http.StatusOK,
				},
				//Successful 204 code
				{
					Name: "success 204 code",
					MockBehaviour: func(s *mock.MockUserService, j *mock.MockJwtService) {
						s.EXPECT().UserEmailConfirmation(gomock.Any(), gomock.Any()).Return(http.StatusNoContent, nil)
					},
					ExceptedStatus: http.StatusNoContent,
				},
				//Service error returned
				{
					Name: "service error returned",
					MockBehaviour: func(s *mock.MockUserService, j *mock.MockJwtService) {
						s.EXPECT().UserEmailConfirmation(gomock.Any(), gomock.Any()).Return(0, errors.New("wrong email"))
					},
					ExceptedStatus: http.StatusInternalServerError,
					ExceptedError:  errors.New("wrong email"),
					ExceptedBody:   `{"messages":["wrong email"],"dev_message":"Something wrong happened while service executing","code":"IE-0001"}`,
				},
				//Service wrong status code
				{
					Name: "service invalid status code",
					MockBehaviour: func(s *mock.MockUserService, j *mock.MockJwtService) {
						s.EXPECT().UserEmailConfirmation(gomock.Any(), gomock.Any()).Return(http.StatusInternalServerError, nil)
					},
					ExceptedStatus: http.StatusInternalServerError,
					ExceptedError:  errors.New("service responded with invalid status code: 500"),
					ExceptedBody:   `{"messages":["service responded with invalid status code: 500"],"dev_message":"Something wrong happened while service executing","code":"IE-0001"}`,
				},
			},
		},
		//UpdateToken
		{
			HandlerName: "UpdateToken",
			Handler:     h.UpdateToken,
			Method:      http.MethodPost,
			Options: []handlerOptions{
				//Successful
				{
					Name: "success",
					MockBehaviour: func(s *mock.MockUserService, j *mock.MockJwtService) {
						j.EXPECT().UpdateRefreshToken(gomock.Any()).Return(&model.JwtResponse{
							Login:        "user",
							Roles:        []string{"user"},
							Token:        "token",
							RefreshToken: "refresh token",
						}, nil)
					},
					InputJson: func() *[]byte {
						byte, _ := json.Marshal(&model.RefreshTokenQuery{
							RefreshToken: "token",
						})
						return &byte
					},
					ExceptedStatus: http.StatusOK,
					ExceptedBody:   "{\"login\":\"user\",\"roles\":[\"user\"],\"token\":\"token\",\"refreshtoken\":\"refresh token\"}",
				},
				//Nil body
				{
					Name:           "nil body",
					ExceptedStatus: http.StatusInternalServerError,
					ExceptedError:  errors.New("EOF"),
					ExceptedBody:   `{"messages":["EOF"],"dev_message":"Something wrong happened while service executing","code":"IE-0001"}`,
				},
				//Validation error
				{
					Name: "validation error",
					InputJson: func() *[]byte {
						byte, _ := json.Marshal(&model.RefreshTokenQuery{})
						return &byte
					},
					ExceptedStatus: http.StatusNotImplemented,
					ExceptedError:  errormiddleware.NewError([]string{"refreshtoken: field is required"}, errormiddleware.ValidationErrorCode, "wrong token format"),
					ExceptedBody:   `{"messages":["refreshtoken: field is required"],"dev_message":"wrong token format","code":"IE-0004"}`,
				},
				//Service error
				{
					Name: "service error",
					InputJson: func() *[]byte {
						byte, _ := json.Marshal(&model.RefreshTokenQuery{RefreshToken: "123"})
						return &byte
					},
					MockBehaviour: func(s *mock.MockUserService, j *mock.MockJwtService) {
						j.EXPECT().UpdateRefreshToken(gomock.Any()).Return(nil, errors.New("wrong token"))
					},
					ExceptedStatus: http.StatusInternalServerError,
					ExceptedError:  errors.New("wrong token"),
					ExceptedBody:   `{"messages":["wrong token"],"dev_message":"Something wrong happened while service executing","code":"IE-0001"}`,
				},
			},
		},
		//FindUser
		{
			HandlerName: "FindUser",
			Handler:     h.FindUser,
			Method:      http.MethodGet,
			Options: []handlerOptions{
				//Successful
				{
					Name: "success",
					MockBehaviour: func(s *mock.MockUserService, j *mock.MockJwtService) {
						s.EXPECT().FindUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(&model.User{
							Login: "user",
							Roles: []string{"user"},
							Email: "user@example.com",
						}, nil)
					},
					ExceptedStatus: http.StatusOK,
					ExceptedBody:   "{\"id\":\"\",\"login\":\"user\",\"roles\":[\"user\"],\"email\":\"user@example.com\",\"emailconfirmed\":false}",
				},
				//Service error
				{
					Name: "service error",
					MockBehaviour: func(s *mock.MockUserService, j *mock.MockJwtService) {
						s.EXPECT().FindUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("wrong id"))
					},
					ExceptedStatus: http.StatusInternalServerError,
					ExceptedError:  errors.New("wrong id"),
					ExceptedBody:   `{"messages":["wrong id"],"dev_message":"Something wrong happened while service executing","code":"IE-0001"}`,
				},
			},
		},
		//Authentication
		{
			HandlerName: "Authenticate",
			Handler:     h.Authenticate,
			Method:      http.MethodPost,
			Options: []handlerOptions{
				//Successful
				{
					Name: "success",
					MockBehaviour: func(s *mock.MockUserService, j *mock.MockJwtService) {
						usr := &model.User{
							Login: "user",
							Roles: []string{"user"},
							Email: "user@example.com",
						}
						gomock.InOrder(
							s.EXPECT().AuthByLoginAndPassword(gomock.Any(), gomock.Any()).Return(usr, nil),
							j.EXPECT().GenerateAccessToken(usr).Return(&model.JwtResponse{
								Login:        usr.Login,
								Roles:        usr.Roles,
								Token:        "token",
								RefreshToken: "refreshtoken",
							}, nil),
						)
					},
					InputJson: func() *[]byte {
						byte, _ := json.Marshal(&model.UserAuthQuery{
							Login:    "user",
							Password: "usr",
						})
						return &byte
					},
					ExceptedStatus: http.StatusOK,
					ExceptedBody:   "{\"login\":\"user\",\"roles\":[\"user\"],\"token\":\"token\",\"refreshtoken\":\"refreshtoken\"}",
				},
				//Nil body
				{
					Name:           "nil body",
					ExceptedStatus: http.StatusInternalServerError,
					ExceptedError:  errors.New("EOF"),
					ExceptedBody:   `{"messages":["EOF"],"dev_message":"Something wrong happened while service executing","code":"IE-0001"}`,
				},
				//Validation error
				{
					Name: "validation",
					InputJson: func() *[]byte {
						byte, _ := json.Marshal(&model.UserAuthQuery{})
						return &byte
					},
					ExceptedStatus: http.StatusNotImplemented,
					ExceptedError:  errormiddleware.NewError([]string{"login: field is required", "password: field is required"}, errormiddleware.ValidationErrorCode, "wrong query format"),
					ExceptedBody:   `{"messages":["login: field is required","password: field is required"],"dev_message":"wrong query format","code":"IE-0004"}`,
				},
				//User service error
				{
					Name: "user service error",
					MockBehaviour: func(s *mock.MockUserService, j *mock.MockJwtService) {
						s.EXPECT().AuthByLoginAndPassword(gomock.Any(), gomock.Any()).Return(nil, errors.New("wrong password"))
					},
					InputJson: func() *[]byte {
						byte, _ := json.Marshal(&model.UserAuthQuery{Login: "user", Password: "password"})
						return &byte
					},
					ExceptedStatus: http.StatusInternalServerError,
					ExceptedError:  errors.New("wrong password"),
					ExceptedBody:   `{"messages":["wrong password"],"dev_message":"Something wrong happened while service executing","code":"IE-0001"}`,
				},
				//Jwt service error
				{
					Name: "jwt service error",
					MockBehaviour: func(s *mock.MockUserService, j *mock.MockJwtService) {
						s.EXPECT().AuthByLoginAndPassword(gomock.Any(), gomock.Any()).Return(nil, nil)
						j.EXPECT().GenerateAccessToken(nil).Return(nil, errors.New("wrong model"))
					},
					InputJson: func() *[]byte {
						byte, _ := json.Marshal(&model.UserAuthQuery{Login: "user", Password: "password"})
						return &byte
					},
					ExceptedStatus: http.StatusInternalServerError,
					ExceptedError:  errors.New("wrong model"),
					ExceptedBody:   `{"messages":["wrong model"],"dev_message":"Something wrong happened while service executing","code":"IE-0001"}`,
				},
			},
		},
		//UserRegister
		{
			HandlerName: "UserRegister",
			Handler:     h.UserRegister,
			Method:      http.MethodPost,
			Options: []handlerOptions{
				//Successful
				{
					Name: "success",
					MockBehaviour: func(s *mock.MockUserService, j *mock.MockJwtService) {
						usr := &model.User{
							Login: "user",
							Roles: []string{"user"},
							Email: "user@example.com",
						}
						gomock.InOrder(
							s.EXPECT().RegisterUser(gomock.Any(), gomock.Any()).Return(usr, nil),
							j.EXPECT().GenerateAccessToken(usr).Return(&model.JwtResponse{
								Login:        usr.Login,
								Roles:        usr.Roles,
								Token:        "token",
								RefreshToken: "refresh token",
							}, nil),
						)
					},
					InputJson: func() *[]byte {
						byte, _ := json.Marshal(&model.UserRegisterQuery{
							Login:    "user",
							Email:    "user@example.com",
							Password: "Password1!",
						})
						return &byte
					},
					ExceptedStatus: http.StatusOK,
					ExceptedBody:   "{\"login\":\"user\",\"roles\":[\"user\"],\"token\":\"token\",\"refreshtoken\":\"refresh token\"}",
				},
				//Nil body
				{
					Name:           "nil body",
					ExceptedStatus: http.StatusInternalServerError,
					ExceptedError:  errors.New("EOF"),
					ExceptedBody:   `{"messages":["EOF"],"dev_message":"Something wrong happened while service executing","code":"IE-0001"}`,
				},
				//Validation error
				{
					Name: "validation error",
					InputJson: func() *[]byte {
						byte, _ := json.Marshal(&model.UserRegisterQuery{})
						return &byte
					},
					ExceptedStatus: http.StatusNotImplemented,
					ExceptedError:  errormiddleware.NewError([]string{"login: field is required", "email: field is required", "password: field is required"}, errormiddleware.ValidationErrorCode, "wrong query format"),
					ExceptedBody:   `{"messages":["login: field is required","email: field is required","password: field is required"],"dev_message":"wrong query format","code":"IE-0004"}`,
				},
				//User service error
				{
					Name: "user service error",
					MockBehaviour: func(s *mock.MockUserService, j *mock.MockJwtService) {
						s.EXPECT().RegisterUser(gomock.Any(), gomock.Any()).Return(nil, errors.New("login already taken"))
					},
					InputJson: func() *[]byte {
						byte, _ := json.Marshal(&model.UserRegisterQuery{
							Login:    "user",
							Email:    "user@example.com",
							Password: "Password1!",
						})
						return &byte
					},
					ExceptedStatus: http.StatusInternalServerError,
					ExceptedError:  errors.New("login already taken"),
					ExceptedBody:   `{"messages":["login already taken"],"dev_message":"Something wrong happened while service executing","code":"IE-0001"}`,
				},
				//Jwt service error
				{
					Name: "jwt service error",
					MockBehaviour: func(s *mock.MockUserService, j *mock.MockJwtService) {
						s.EXPECT().RegisterUser(gomock.Any(), gomock.Any()).Return(&model.User{}, nil)
						j.EXPECT().GenerateAccessToken(gomock.Any()).Return(nil, errors.New("wrong model"))
					},
					InputJson: func() *[]byte {
						byte, _ := json.Marshal(&model.UserRegisterQuery{
							Login:    "user",
							Email:    "user@example.com",
							Password: "Password1!",
						})
						return &byte
					},
					ExceptedStatus: http.StatusInternalServerError,
					ExceptedError:  errors.New("wrong model"),
					ExceptedBody:   `{"messages":["wrong model"],"dev_message":"Something wrong happened while service executing","code":"IE-0001"}`,
				},
			},
		},
	}
	for _, tt := range testTable {
		for _, testCase := range tt.Options {
			t.Run(fmt.Sprintf("%s %s", tt.HandlerName, testCase.Name), func(t *testing.T) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()
				jwt := mock.NewMockJwtService(ctrl)
				user := mock.NewMockUserService(ctrl)
				if testCase.MockBehaviour != nil {
					testCase.MockBehaviour(user, jwt)
				}
				h.JwtService = jwt
				h.UserService = user

				w := httptest.NewRecorder()
				var r *http.Request
				if testCase.InputJson != nil && testCase.InputJson() != nil {
					r = httptest.NewRequest(tt.Method, "http://test", bytes.NewBuffer(*testCase.InputJson()))
				} else {
					r = httptest.NewRequest(tt.Method, "http://test", nil)
				}
				err := errormiddleware.Middleware(tt.Handler)(w, r)
				assert.Equal(t, testCase.ExceptedStatus, w.Result().StatusCode)
				assert.Equal(t, testCase.ExceptedError, err)

				body := w.Body.String()
				if assert.Len(t, body, len(testCase.ExceptedBody)) {
					assert.Equal(t, testCase.ExceptedBody, body)
				}
			})
		}
	}
}
