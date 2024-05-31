package user

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/reversersed/go-web-services/tree/main/api_user/internal/client"
	"github.com/reversersed/go-web-services/tree/main/api_user/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_user/pkg/logging"
)

const (
	url_auth              = "/users/auth"
	url_register          = "/users/register"
	url_email_confirmaton = "/users/email"
	url_user_find         = "/users"
	url_user_delete       = "/users/delete"
	url_user_changelogin  = "/users/changename"
)

type Service interface {
	SignInUser(ctx context.Context, query *client.AuthUserByLoginAndPassword) (*client.User, error)
	RegisterUser(ctx context.Context, query *client.RegisterUserQuery) (*client.User, error)
	SendEmailConfirmation(ctx context.Context, userId string) error
	ValidateEmailConfirmationCode(ctx context.Context, userId string, code string) error
	GetUserById(ctx context.Context, userId string) (*client.User, error)
	GetUserByLogin(ctx context.Context, login string) (*client.User, error)
	DeleteUser(ctx context.Context, userId, password string) error
	UpdateUserLogin(ctx context.Context, userId, newLogin string) (*client.User, error)
}
type Handler struct {
	Logger      *logging.Logger
	UserService Service
}

func (h *Handler) Register(route *httprouter.Router) {
	route.HandlerFunc(http.MethodPost, url_auth, errormiddleware.Middleware(h.AuthUser))
	route.HandlerFunc(http.MethodPost, url_register, errormiddleware.Middleware(h.RegUser))
	route.HandlerFunc(http.MethodGet, url_email_confirmaton, errormiddleware.Middleware(h.ConfirmEmail))
	route.HandlerFunc(http.MethodGet, url_user_find, errormiddleware.Middleware(h.FindUser))
	route.HandlerFunc(http.MethodDelete, url_user_delete, errormiddleware.Middleware(h.DeleteUser))
	route.HandlerFunc(http.MethodPatch, url_user_changelogin, errormiddleware.Middleware(h.ChangeUserLogin))
}
func (h *Handler) ChangeUserLogin(w http.ResponseWriter, r *http.Request) error {
	userId := r.Header.Get("User")
	if len(userId) <= 0 {
		return errormiddleware.UnauthorizedError([]string{"can't get user authorized id"}, "context id was empty")
	}
	var query client.ChangeUserLoginQuery
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	u, err := h.UserService.UpdateUserLogin(ctx, userId, query.Login)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	object, err := json.Marshal(u)
	if err != nil {
		return err
	}
	w.Write(object)
	return nil
}
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) error {
	userId := r.Header.Get("User")
	if len(userId) <= 0 {
		return errormiddleware.UnauthorizedError([]string{"can't get user authorized id"}, "context id was empty")
	}
	var query client.DeleteUserQuery
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err := h.UserService.DeleteUser(ctx, userId, query.Password)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
func (h *Handler) FindUser(w http.ResponseWriter, r *http.Request) error {
	user_principal := r.URL.Query().Get("id")

	if len(user_principal) > 0 {
		u, err := h.UserService.GetUserById(r.Context(), user_principal)
		if err != nil {
			return err
		}
		bytes, err := json.Marshal(u)
		if err != nil {
			return err
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/json")
		w.Write(bytes)
		return nil
	}
	user_principal = r.URL.Query().Get("login")

	if len(user_principal) > 0 {
		u, err := h.UserService.GetUserByLogin(r.Context(), user_principal)
		if err != nil {
			return err
		}
		bytes, err := json.Marshal(u)
		if err != nil {
			return err
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/json")
		w.Write(bytes)
		return nil
	}
	return errormiddleware.BadRequestError([]string{"query has to have one of parameters", "login: user login", "id: user id"}, "bad request provided")
}
func (h *Handler) ConfirmEmail(w http.ResponseWriter, r *http.Request) error {
	userId := r.Header.Get("User")
	if len(userId) <= 0 {
		return errormiddleware.UnauthorizedError([]string{"can't get user authorized id"}, "context id was empty")
	}
	if code := r.URL.Query().Get("code"); len(code) > 0 {
		err := h.UserService.ValidateEmailConfirmationCode(r.Context(), userId, code)
		if err != nil {
			return err
		}
		w.WriteHeader(http.StatusNoContent)
	} else {
		err := h.UserService.SendEmailConfirmation(r.Context(), userId)
		if err != nil {
			return err
		}
		w.WriteHeader(http.StatusOK)
	}
	return nil
}
func (h *Handler) AuthUser(w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()
	var query client.AuthUserByLoginAndPassword

	h.Logger.Info("decoding auth body...")
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		h.Logger.Warn("error occured while decoding request body: %w", err)
		return errormiddleware.BadRequestError([]string{"invalid json scheme"}, err.Error())
	}
	u, err := h.UserService.SignInUser(r.Context(), &query)
	if err != nil {
		return errormiddleware.NotFoundError([]string{"user with provided login and password not found"}, err.Error())
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
		return errormiddleware.BadRequestError([]string{"invalid json scheme"}, err.Error())
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
	h.Logger.Infof("user %s has been registered", query.Login)
	return nil
}
