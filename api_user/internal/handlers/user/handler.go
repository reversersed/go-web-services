package user

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/reversersed/go-web-services/tree/main/api_user/internal/client"
	"github.com/reversersed/go-web-services/tree/main/api_user/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_user/pkg/logging"
)

const (
	url_auth     = "/users/auth"
	url_register = "/users/register"
)

type Service interface {
	SignInUser(ctx context.Context, query *client.AuthUserByLoginAndPassword) (*client.User, error)
	RegisterUser(ctx context.Context, query *client.RegisterUserQuery) (*client.User, error)
}
type Handler struct {
	Logger      *logging.Logger
	UserService Service
}

func (h *Handler) Register(route *httprouter.Router) {
	route.HandlerFunc(http.MethodPost, url_auth, errormiddleware.Middleware(h.AuthUser))
	route.HandlerFunc(http.MethodPost, url_register, errormiddleware.Middleware(h.RegUser))
}

func (h *Handler) AuthUser(w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()
	var query client.AuthUserByLoginAndPassword

	h.Logger.Info("decoding auth body...")
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		h.Logger.Warn("error occured while decoding request body: %w", err)
		return errormiddleware.BadRequestError("invalid json scheme", []string{err.Error()})
	}
	u, err := h.UserService.SignInUser(r.Context(), &query)
	if err != nil {
		return errormiddleware.NotFoundError("user with provided login and password not found", []string{err.Error()})
	}

	object, err := json.Marshal(u)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(object)
	return nil
}
func (h *Handler) RegUser(w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()
	var query client.RegisterUserQuery
	h.Logger.Info("decoding auth body...")
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		h.Logger.Warn("error occured while decoding request body: %w", err)
		return errormiddleware.BadRequestError("invalid json scheme", []string{err.Error()})
	}

	u, err := h.UserService.RegisterUser(r.Context(), &query)
	if err != nil {
		return err
	}
	object, err := json.Marshal(u)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(object)
	return nil
}
