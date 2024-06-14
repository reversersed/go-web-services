package user

import (
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/julienschmidt/httprouter"
	mock_jwt "github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/jwt/mocks"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
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
	jwt := mock_jwt.NewMockJwtService(ctrl)
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
