package auth

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	model "github.com/reversersed/go-web-services/tree/main/api_gateway/internal/client/auth"
	mw "github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
)

const (
	url_auth   = "/api/v1/user/login"
	url_signup = "/api/v1/user/signup"
)

type UserService interface {
	AuthByLoginAndPassword(ctx context.Context, login, password string) (*model.User, error)
}

type Handler struct {
	Logger *logging.Logger
	//jwt
	UserService UserService
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, url_auth, mw.Middleware(h.Authenticate))
	h.Logger.Info("auth service registered")
}

func (h *Handler) Authenticate(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	token, _ := json.Marshal("хуй")
	switch r.Method {
	case http.MethodPost:
		defer r.Body.Close()

		//put to update refresh token
	}

	w.WriteHeader(http.StatusOK)
	w.Write(token)

	return nil
}
