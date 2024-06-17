package user

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

var userList = []*User{
	{
		Id:             "1",
		Login:          "user",
		Email:          "user@example.com",
		EmailConfirmed: true,
		Roles:          []string{"user"},
	},
	{
		Id:             "2",
		Login:          "admin",
		Email:          "admin@example.com",
		EmailConfirmed: true,
		Roles:          []string{"user", "admin"},
	},
}

// Test FindUser() method
var findUserCases = []struct {
	Name      string
	UserId    string
	UserLogin string
	Excepted  *User
	ErrorCode errormiddleware.Code
}{
	{"Find user by id", "1", "", userList[0], ""},
	{"Find user by login", "", "admin", userList[1], ""},
	{"Error finding user", "0", "", nil, errormiddleware.NotFoundErrorCode},
	{"Request without fields", "", "", nil, errormiddleware.BadRequestErrorCode},
	{"Empty body received", "-1", "", nil, errormiddleware.InternalErrorCode},
}

func TestFindUser(t *testing.T) {
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		login := r.URL.Query().Get("login")
		for _, user := range userList {
			if user.Id == id || user.Login == login {
				us, _ := json.Marshal(user)
				w.WriteHeader(http.StatusOK)
				w.Write(us)
				return
			}
		}
		if id == "-1" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write(errormiddleware.NotFoundError([]string{"user not found"}, "not found").Marshall())
	}))
	for _, findUserCase := range findUserCases {
		t.Run(findUserCase.Name, func(t *testing.T) {
			service := NewService(server.URL, "/users", logger)
			user, err := service.FindUser(context.Background(), findUserCase.UserId, findUserCase.UserLogin)
			if len(findUserCase.ErrorCode) > 0 {
				assert.Error(t, err)
				e, ok := err.(*errormiddleware.Error)
				if ok {
					assert.Equal(t, e.Code, findUserCase.ErrorCode)
				} else {
					assert.Equal(t, findUserCase.ErrorCode, errormiddleware.InternalErrorCode)
				}
			}
			assert.Equal(t, findUserCase.Excepted, user)
		})
	}
}

// Test UserEmailConfirmation() method
var userEmailCases = []struct {
	Name           string
	Code           string
	ExceptedStatus int
	ErrorCode      errormiddleware.Code
}{
	{"Sending user email", "", http.StatusOK, ""},
	{"User code confirmation", "1111", http.StatusNoContent, ""},
	{"Wrong code confirmation", "1234", http.StatusNotFound, errormiddleware.NotFoundErrorCode},
}

func TestUserEmailConfirmation(t *testing.T) {
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		switch code {
		case "":
			w.WriteHeader(http.StatusOK)
		case "1111":
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write(errormiddleware.NotFoundError([]string{"incorrect code"}, "not found").Marshall())
		}
	}))
	for _, userEmailCase := range userEmailCases {
		t.Run(userEmailCase.Name, func(t *testing.T) {
			service := NewService(server.URL, "/users", logger)
			result, err := service.UserEmailConfirmation(context.Background(), userEmailCase.Code)
			if len(userEmailCase.ErrorCode) > 0 {
				assert.Error(t, err)
				e, ok := err.(*errormiddleware.Error)
				assert.True(t, ok)
				assert.Equal(t, e.Code, userEmailCase.ErrorCode)
			} else {
				assert.Equal(t, userEmailCase.ExceptedStatus, result)
			}
		})
	}
}

// Test AuthByLoginAndPassword() method
var authUserCases = []struct {
	Name      string
	Query     *UserAuthQuery
	Excepted  *User
	ErrorCode errormiddleware.Code
}{
	{"User authentication", &UserAuthQuery{Login: "admin", Password: "admin"}, userList[0], ""},
	{"Nil query", nil, nil, errormiddleware.BadRequestErrorCode},
	{"Wrong password", &UserAuthQuery{Login: "admin", Password: "123"}, nil, errormiddleware.NotFoundErrorCode},
	{"Empty body received", &UserAuthQuery{Login: "user"}, nil, errormiddleware.InternalErrorCode},
}

func TestAuthByLoginAndPassword(t *testing.T) {
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var query UserAuthQuery
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(errormiddleware.BadRequestError([]string{""}, "").Marshall())
			return
		}
		if query.Login == "admin" && query.Password == "admin" {
			w.WriteHeader(http.StatusOK)
			user, _ := json.Marshal(userList[0])
			w.Write(user)
			return
		}
		if query.Login == "user" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write(errormiddleware.NotFoundError([]string{""}, "").Marshall())
	}))
	for _, authUserCase := range authUserCases {
		t.Run(authUserCase.Name, func(t *testing.T) {
			service := NewService(server.URL, "/users", logger)
			user, err := service.AuthByLoginAndPassword(context.Background(), authUserCase.Query)
			if len(authUserCase.ErrorCode) > 0 {
				assert.Error(t, err)
				e, ok := err.(*errormiddleware.Error)
				if ok {
					assert.Equal(t, e.Code, authUserCase.ErrorCode)
				} else {
					assert.Equal(t, authUserCase.ErrorCode, errormiddleware.InternalErrorCode)
				}
			} else if assert.NotNil(t, user) {
				assert.Equal(t, authUserCase.Excepted.Id, user.Id)
			}
		})
	}
}

// Test RegisterUser() method
var registerUserCases = []struct {
	Name      string
	Query     *UserRegisterQuery
	Excepted  *User
	ErrorCode errormiddleware.Code
}{
	{"User registration", &UserRegisterQuery{Login: "admin", Password: "admin"}, userList[1], ""},
	{"Nil query", nil, nil, errormiddleware.BadRequestErrorCode},
	{"Wrong password", &UserRegisterQuery{Login: "admin", Password: "123"}, nil, errormiddleware.NotFoundErrorCode},
	{"Empty body received", &UserRegisterQuery{Login: "user"}, nil, errormiddleware.InternalErrorCode},
}

func TestRegisterUser(t *testing.T) {
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var query UserRegisterQuery
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(errormiddleware.BadRequestError([]string{""}, "").Marshall())
			return
		}
		if query.Login == "admin" && query.Password == "admin" {
			w.WriteHeader(http.StatusOK)
			user, _ := json.Marshal(userList[1])
			w.Write(user)
			return
		}
		if query.Login == "user" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write(errormiddleware.NotFoundError([]string{""}, "").Marshall())
	}))
	for _, registerUserCase := range registerUserCases {
		t.Run(registerUserCase.Name, func(t *testing.T) {
			service := NewService(server.URL, "/users", logger)
			user, err := service.RegisterUser(context.Background(), registerUserCase.Query)
			if len(registerUserCase.ErrorCode) > 0 {
				assert.Error(t, err)
				e, ok := err.(*errormiddleware.Error)
				if ok {
					assert.Equal(t, e.Code, registerUserCase.ErrorCode)
				} else {
					assert.Equal(t, registerUserCase.ErrorCode, errormiddleware.InternalErrorCode)
				}
			} else if assert.NotNil(t, user) {
				assert.Equal(t, registerUserCase.Excepted.Id, user.Id)
			}
		})
	}
}

// Test DeleteUser() method
var deleteUserCases = []struct {
	Name      string
	Query     *DeleteUserQuery
	Excepted  *User
	ErrorCode errormiddleware.Code
}{
	{"User deleting", &DeleteUserQuery{Password: "admin"}, userList[1], ""},
	{"Nil query", nil, nil, errormiddleware.BadRequestErrorCode},
	{"Wrong password", &DeleteUserQuery{Password: "123"}, nil, errormiddleware.NotFoundErrorCode},
}

func TestDeleteUser(t *testing.T) {
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var query DeleteUserQuery
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(errormiddleware.BadRequestError([]string{""}, "").Marshall())
			return
		}
		if query.Password == "admin" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write(errormiddleware.NotFoundError([]string{""}, "").Marshall())
	}))
	for _, deleteUserCase := range deleteUserCases {
		t.Run(deleteUserCase.Name, func(t *testing.T) {
			service := NewService(server.URL, "/users", logger)
			err := service.DeleteUser(context.Background(), deleteUserCase.Query)
			if len(deleteUserCase.ErrorCode) > 0 {
				assert.Error(t, err)
				e, ok := err.(*errormiddleware.Error)
				if ok {
					assert.Equal(t, e.Code, deleteUserCase.ErrorCode)
				} else {
					assert.Equal(t, deleteUserCase.ErrorCode, errormiddleware.InternalErrorCode)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test UpdateUserLogin() method
var updateUserLoginCases = []struct {
	Name          string
	Query         *UpdateUserLoginQuery
	ExceptedLogin string
	ErrorCode     errormiddleware.Code
}{
	{"User registration", &UpdateUserLoginQuery{NewLogin: "admin1"}, "admin1", ""},
	{"Nil query", nil, "", errormiddleware.BadRequestErrorCode},
	{"Not found user", &UpdateUserLoginQuery{NewLogin: ""}, "", errormiddleware.NotFoundErrorCode},
	{"Empty body received", &UpdateUserLoginQuery{NewLogin: "user"}, "", errormiddleware.InternalErrorCode},
}

func TestUpdateUserLogin(t *testing.T) {
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var query UpdateUserLoginQuery
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(errormiddleware.BadRequestError([]string{""}, "").Marshall())
			return
		}
		if query.NewLogin == "admin1" {
			w.WriteHeader(http.StatusOK)
			usr := *userList[1]
			usr.Login = query.NewLogin
			user, _ := json.Marshal(usr)
			w.Write(user)
			return
		}
		if query.NewLogin == "user" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write(errormiddleware.NotFoundError([]string{""}, "").Marshall())
	}))
	for _, updateUserLoginCase := range updateUserLoginCases {
		t.Run(updateUserLoginCase.Name, func(t *testing.T) {
			service := NewService(server.URL, "/users", logger)
			user, err := service.UpdateUserLogin(context.Background(), updateUserLoginCase.Query)
			if len(updateUserLoginCase.ErrorCode) > 0 {
				assert.Error(t, err)
				e, ok := err.(*errormiddleware.Error)
				if ok {
					assert.Equal(t, e.Code, updateUserLoginCase.ErrorCode)
				} else {
					assert.Equal(t, updateUserLoginCase.ErrorCode, errormiddleware.InternalErrorCode)
				}
			} else if assert.NotNil(t, user) {
				assert.Equal(t, updateUserLoginCase.ExceptedLogin, user.Login)
			}
		})
	}
}
