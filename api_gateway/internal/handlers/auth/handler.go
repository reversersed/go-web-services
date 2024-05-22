package auth

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
	model "github.com/reversersed/go-web-services/tree/main/api_gateway/internal/client/auth"
	mw "github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/jwt"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
)

const (
	url_auth   = "/api/v1/user/login"
	url_signup = "/api/v1/user/signup"
)

type UserService interface {
	AuthByLoginAndPassword(ctx context.Context, query *model.UserAuthQuery) (*model.User, error)
}

type Handler struct {
	Logger      *logging.Logger
	JwtService  jwt.JwtService
	UserService UserService
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, url_auth, mw.Middleware(h.Authenticate))
	router.HandlerFunc(http.MethodPut, url_auth, mw.Middleware(h.Authenticate))
	h.Logger.Info("auth service registered")
}

func (h *Handler) Authenticate(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	var token []byte
	var err error
	switch r.Method {
	case http.MethodPost:
		defer r.Body.Close()
		var query model.UserAuthQuery
		if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
			return err
		}
		if err = validator.New().Struct(query); err != nil {
			return mw.ValidationError(err.Error())
		}
		model, err := h.UserService.AuthByLoginAndPassword(r.Context(), &query)
		if err != nil {
			return err
		}
		token, err = h.JwtService.GenerateAccessToken(model)
		if err != nil {
			h.Logger.Warn(err)
			return err
		}
	case http.MethodPut:
		defer r.Body.Close()
		var query jwt.RefreshTokenQuery
		if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
			return err
		}
		if err = validator.New().Struct(query); err != nil {
			return mw.ValidationError(err.Error())
		}
		token, err = h.JwtService.UpdateRefreshToken(&query)
		if err != nil {
			h.Logger.Warn(err)
			return err
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write(token)
	return nil
}
