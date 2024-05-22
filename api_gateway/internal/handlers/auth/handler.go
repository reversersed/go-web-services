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
	url_auth    = "/api/v1/users/login"
	url_refresh = "/api/v1/users/refresh"
	url_signup  = "/api/v1/users/signup"
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
	router.HandlerFunc(http.MethodPost, url_refresh, mw.Middleware(h.UpdateToken))
	h.Logger.Info("auth service registered")
}

// @Summary Generate new token
// @Description Generate new token by provided refresh token
// @Description Refresh token stored in cache and expires in 7 days. If system was restarted, all tokens are cleared and sessions are deleted
// @Tags users
// @Produce json
// @Param query body jwt.RefreshTokenQuery true "Request query with user's refresh token"
// @Success 200 {object} jwt.JwtResponse
// @Failure 404 {object} errormiddleware.Error
// @Failure 500 {object} errormiddleware.Error
// @Failure 501 {object} errormiddleware.Error
// @Router /users/refresh [post]
func (h *Handler) UpdateToken(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	defer r.Body.Close()
	var query jwt.RefreshTokenQuery
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		return err
	}
	if err := validator.New().Struct(query); err != nil {
		return mw.ValidationError(err.Error())
	}
	token, err := h.JwtService.UpdateRefreshToken(&query)
	if err != nil {
		h.Logger.Warn(err)
		return err
	}

	data, _ := json.Marshal(token)

	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return nil
}

// @Summary Authenticates user
// @Description Finds user by login and password
// @Description Returns a token and refresh token. Token expires in 1 hour, refresh token expires in 7 days and stores in cache (removing after system restart)
// @Produce json
// @Tags users
// @Param query body model.UserAuthQuery true "User credentials"
// @Success 200 {object} jwt.JwtResponse
// @Failure 404 {object} errormiddleware.Error
// @Failure 500 {object} errormiddleware.Error
// @Failure 501 {object} errormiddleware.Error
// @Router /users/login [post]
func (h *Handler) Authenticate(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	defer r.Body.Close()
	var query model.UserAuthQuery
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		return err
	}
	if err := validator.New().Struct(query); err != nil {
		return mw.ValidationError(err.Error())
	}
	model, err := h.UserService.AuthByLoginAndPassword(r.Context(), &query)
	if err != nil {
		return err
	}
	token, err := h.JwtService.GenerateAccessToken(model)
	if err != nil {
		h.Logger.Warn(err)
		return err
	}
	data, _ := json.Marshal(token)

	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return nil
}
