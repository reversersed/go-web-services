package jwt

import (
	"encoding/json"
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
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var testCases = []struct {
	Name         string
	Uid          string
	UserRole     []string
	RequiredRole string
	StatusCode   int
}{
	{"default user authorization", "userid", []string{"user"}, "", http.StatusOK},
	{"default admin authorization", "userid", []string{"user", "admin"}, "admin", http.StatusOK},
	{"admin authorization as user", "uid", []string{"user", "admin"}, "", http.StatusOK},
	{"user authorization as admin", "uid", []string{"user"}, "admin", http.StatusForbidden},
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
				assert.Equal(t, testCase.Uid, r.Context().Value(rest.UserIdKey))
				w.WriteHeader(http.StatusOK)
			}), testCase.RequiredRole)

			u := &user.User{
				Id:    testCase.Uid,
				Roles: testCase.UserRole,
			}
			token, err := service.GenerateAccessToken(u)
			assert.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "http://test", nil)
			req.AddCookie(token.Token)
			req.AddCookie(token.RefreshToken)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, req)

			assert.Equal(t, testCase.StatusCode, response.Result().StatusCode)
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
	assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
}
func TestNilKeyMiddleware(t *testing.T) {
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}
	val := validator.New()
	cache := freecache.NewCache(0)
	service := NewService(cache, logger, val, "")
	handler := service.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
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
	r.AddCookie(&http.Cookie{Name: TokenCookieName, Value: token.String()})
	handler.ServeHTTP(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
}
func TestNilTokenMiddleware(t *testing.T) {
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}
	val := validator.New()
	cache := freecache.NewCache(0)
	service := NewService(cache, logger, val, "secretCode")
	handler := service.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://test", nil)
	r.AddCookie(&http.Cookie{Name: TokenCookieName, Value: ""})
	handler.ServeHTTP(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
}
func TestOldTokenMiddleware(t *testing.T) {
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}
	val := validator.New()
	cache := freecache.NewCache(0)
	service := NewService(cache, logger, val, "secretCode")
	handler := service.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	signer, _ := jwt.NewSignerHS(jwt.HS256, []byte("secretCode"))
	builder := jwt.NewBuilder(signer)

	u := &user.User{
		Id:    "userId",
		Login: "user",
		Roles: []string{"user"},
		Email: "user@example.com",
	}
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        u.Id,
			Audience:  u.Roles,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		},
		Roles: u.Roles,
		Login: u.Login,
		Email: u.Email,
	}
	token, _ := builder.Build(claims)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://test", nil)

	refreshToken := primitive.NewObjectID().Hex()
	r.AddCookie(&http.Cookie{Name: TokenCookieName, Value: token.String()})

	handler.ServeHTTP(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)

	r.AddCookie(&http.Cookie{Name: RefreshCookieName, Value: refreshToken})
	userBytes, _ := json.Marshal(u)
	cache.Set([]byte(refreshToken), userBytes, int((7*24*time.Hour)/time.Second))

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
}
