package jwt

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cristalhq/jwt/v3"
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
}{
	{"default user authorization", "userid", []string{"user"}, "", 200},
	{"default admin authorization", "userid", []string{"user", "admin"}, "admin", 200},
	{"admin authorization as user", "uid", []string{"user","admin"}, "", 200},
	{"user authorization as admin", "uid", []string{"user"}, "admin", 403},
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
func TestEmptyRequestMiddleware(t *testing.T) {
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}
	val := validator.New()
	cache := freecache.NewCache(0)
	service := NewService(cache, logger, val, "")
	handler := service.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://test", nil)
	handler.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusUnauthorized {
		t.Fatalf("excepted status code 404, but got %d", w.Result().StatusCode)
	}
}
func TestNilKeyMiddleware(t *testing.T) {
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}
	val := validator.New()
	cache := freecache.NewCache(0)
	service := NewService(cache, logger, val, "")
	handler := service.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	w := httptest.NewRecorder()
	signer, _ := jwt.NewSignerHS(jwt.HS256, []byte("secretCode"))
	builder := jwt.NewBuilder(signer)

	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "userId",
			Audience:  []string{"user"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		},
		Roles: []string{"user"},
		Login: "user",
		Email: "user@example.com",
	}
	token, _ := builder.Build(claims)

	r := httptest.NewRequest(http.MethodGet, "http://test", nil)
	r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.String()))
	handler.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusUnauthorized {
		t.Fatalf("excepted status code 404, but got %d", w.Result().StatusCode)
	}
}
func TestNilTokenMiddleware(t *testing.T) {
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}
	val := validator.New()
	cache := freecache.NewCache(0)
	service := NewService(cache, logger, val, "secretCode")
	handler := service.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://test", nil)
	r.Header.Add("Authorization", "Bearer ")
	handler.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusUnauthorized {
		t.Fatalf("excepted status code 404, but got %d", w.Result().StatusCode)
	}
}
func TestOldTokenMiddleware(t *testing.T) {
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}
	val := validator.New()
	cache := freecache.NewCache(0)
	service := NewService(cache, logger, val, "secretCode")
	handler := service.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	signer, _ := jwt.NewSignerHS(jwt.HS256, []byte("secretCode"))
	builder := jwt.NewBuilder(signer)

	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "userId",
			Audience:  []string{"user"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		},
		Roles: []string{"user"},
		Login: "user",
		Email: "user@example.com",
	}
	token, _ := builder.Build(claims)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://test", nil)
	r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.String()))
	handler.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusUnauthorized {
		t.Fatalf("excepted status code 404, but got %d", w.Result().StatusCode)
	}
}
