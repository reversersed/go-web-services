package jwt

import (
	"testing"

	"github.com/cristalhq/jwt/v3"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/internal/client/user"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/cache/freecache"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/validator"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
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
		Err:    jwt.Error("jwt: key is nil"),
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

			assert.Equal(t, testCase.Err, err)
			if response != nil {
				assert.Equal(t, response.Login, testCase.User.Login)
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
			assert.NoError(t, err)

			response, err := service.UpdateRefreshToken(token.RefreshToken.Value)
			assert.Equal(t, err, testCase.Err)

			if response != nil {
				assert.Equal(t, response.Login, testCase.User.Login)
			}

			_, err = service.UpdateRefreshToken(token.RefreshToken.Value)
			assert.Error(t, err)
		})
	}
}

func TestWrongRefreshToken(t *testing.T) {
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}
	val := validator.New()
	cache := freecache.NewCache(0)
	service := NewService(cache, logger, val, "secret")

	_, err := service.UpdateRefreshToken("")
	assert.Error(t, err)
}
