package jwt

import (
	"errors"
	"testing"

	"github.com/reversersed/go-web-services/tree/main/api_gateway/internal/client/user"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/cache/freecache"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/validator"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

var generateTokenCases = []struct {
	Name   string
	User   user.User
	Secret string
	Err    error
}{
	{
		Name: "User token generating",
		User: user.User{
			Id:    "userTestedId",
			Login: "User",
			Roles: []string{"user"},
			Email: "email@example.com",
		},
		Secret: "secretKey",
	},
	{
		Name:   "nil secret key",
		User:   user.User{},
		Secret: "",
		Err:    errors.New("jwt: key is nil"),
	},
}

func TestGenerateToken(t *testing.T) {
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}
	val := validator.New()
	cache := freecache.NewCache(0)

	for _, testCase := range generateTokenCases {
		t.Run(testCase.Name, func(t *testing.T) {
			service := NewService(cache, logger, val, testCase.Secret)
			response, err := service.GenerateAccessToken(&testCase.User)
			if testCase.Err == nil && err != testCase.Err {
				t.Fatalf("excepted error nil but got %v", err)
			} else if testCase.Err != nil && err == nil {
				t.Fatalf("excepted error %v but got nil", testCase.Err)
			} else if testCase.Err != nil && err != nil && testCase.Err.Error() != err.Error() {
				t.Fatalf("excepted error %v but got %v", testCase.Err, err)
			}

			if response != nil && response.Login != testCase.User.Login {
				t.Fatalf("excepted login %s but got %s", testCase.User.Login, response.Login)
			}
		})
	}
}

var updateTokenCases = []struct {
	Name   string
	User   user.User
	Secret string
	Err    error
}{
	{
		Name: "User token updating",
		User: user.User{
			Id:    "userTestedId",
			Login: "User",
			Roles: []string{"user"},
			Email: "email@example.com",
		},
		Secret: "secretKey",
	},
}

func TestUpdateToken(t *testing.T) {
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}
	val := validator.New()
	cache := freecache.NewCache(0)

	for _, testCase := range updateTokenCases {
		t.Run(testCase.Name, func(t *testing.T) {
			service := NewService(cache, logger, val, testCase.Secret)

			token, err := service.GenerateAccessToken(&testCase.User)
			if err != nil {
				t.Fatalf("excepted token but got error %v", err)
			}
			response, err := service.UpdateRefreshToken(&RefreshTokenQuery{RefreshToken: token.RefreshToken})
			if testCase.Err == nil && err != testCase.Err {
				t.Fatalf("excepted error nil but got %v", err)
			} else if testCase.Err != nil && err == nil {
				t.Fatalf("excepted error %v but got nil", testCase.Err)
			} else if testCase.Err != nil && err != nil && testCase.Err.Error() != err.Error() {
				t.Fatalf("excepted error %v but got %v", testCase.Err, err)
			}

			if response != nil && response.Login != testCase.User.Login {
				t.Fatalf("excepted login %s but got %s", testCase.User.Login, response.Login)
			}

			_, err = service.UpdateRefreshToken(&RefreshTokenQuery{RefreshToken: token.RefreshToken})
			if err == nil {
				t.Fatalf("excepted error but got nil")
			}
		})
	}
}

func TestNilRefreshToken(t *testing.T) {
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}
	val := validator.New()
	cache := freecache.NewCache(0)
	service := NewService(cache, logger, val, "secret")

	_, err := service.UpdateRefreshToken(nil)
	if err == nil {
		t.Fatalf("excepted error but got nil")
	}
}
func TestWrongRefreshToken(t *testing.T) {
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}
	val := validator.New()
	cache := freecache.NewCache(0)
	service := NewService(cache, logger, val, "secret")

	_, err := service.UpdateRefreshToken(&RefreshTokenQuery{})
	if err == nil {
		t.Fatalf("excepted error but got nil")
	}
}
