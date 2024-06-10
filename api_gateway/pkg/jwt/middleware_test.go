package jwt

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/reversersed/go-web-services/tree/main/api_gateway/internal/client/user"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/cache/freecache"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/rest"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/validator"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

var testCases = []struct {
	Name         string
	Uid          string
	UserRole     []string
	RequiredRole string
	StatusCode   int
	Err          error
}{
	{"default user authorization", "userid", []string{"user"}, "", 200, nil},
}

func TestMiddleware(t *testing.T) {
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}
	val := validator.New()
	cache := freecache.NewCache(0)
	service := NewService(cache, logger, val, "secretCode")

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			handler := service.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if key := r.Context().Value(rest.UserIdKey); key != testCase.Uid {
					t.Fatalf("excepted user id %s but got %s", testCase.Uid, key)
				}
				w.WriteHeader(200)
			}), testCase.RequiredRole)

			u := &user.User{
				Id:    testCase.Uid,
				Roles: testCase.UserRole,
			}
			token, err := service.GenerateAccessToken(u)
			if err != nil {
				t.Fatalf("excepted token but got error %v", err)
			}

			req := httptest.NewRequest(http.MethodGet, "http://test", nil)
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.Token))
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, req)

			if response.Result().StatusCode != testCase.StatusCode {
				t.Fatalf("excepted status code %d but got %d", testCase.StatusCode, response.Result().StatusCode)
			}
		})
	}
}
func TestNilKeyMiddleware(t *testing.T) {

}
func TestNilTokenMiddleware(t *testing.T) {

}
func TestOldTokenMiddleware(t *testing.T) {

}
