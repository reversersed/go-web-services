package user

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
	model "github.com/reversersed/go-web-services/tree/main/api_gateway/internal/client/user"
	mw "github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/jwt"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
	valid "github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/validator"
)

const (
	url_auth          = "/api/v1/users/login"
	url_refresh       = "/api/v1/users/refresh"
	url_register      = "/api/v1/users/register"
	url_confirm_email = "/api/v1/users/email"
	url_find_user     = "/api/v1/users"
	url_delete_user   = "/api/v1/users/delete"
)

type UserService interface {
	AuthByLoginAndPassword(ctx context.Context, query *model.UserAuthQuery) (*model.User, error)
	RegisterUser(ctx context.Context, query *model.UserRegisterQuery) (*model.User, error)
	UserEmailConfirmation(ctx context.Context, code string) (int, error)
	FindUser(ctx context.Context, userid string, login string) (*model.User, error)
	DeleteUser(ctx context.Context, query *model.DeleteUserQuery) error
}

type Handler struct {
	Logger      *logging.Logger
	JwtService  jwt.JwtService
	UserService UserService
	Validator   *valid.Validator
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, url_auth, mw.Middleware(h.Authenticate))
	router.HandlerFunc(http.MethodPost, url_refresh, mw.Middleware(h.UpdateToken))
	router.HandlerFunc(http.MethodPost, url_register, mw.Middleware(h.UserRegister))
	router.HandlerFunc(http.MethodGet, url_confirm_email, jwt.Middleware(mw.Middleware(h.EmailConfirmation)))
	router.HandlerFunc(http.MethodGet, url_find_user, mw.Middleware(h.FindUser))
	router.HandlerFunc(http.MethodDelete, url_delete_user, jwt.Middleware(mw.Middleware(h.DeleteUser)))
	h.Logger.Info("auth service registered")
}

// @Summary Deletes user's account
// @Description Only user can delete his own account. To delete user he needs to confirm his password
// @Tags users
// @Produce json
// @Param Password body model.DeleteUserQuery true "User password"
// @Success 204 "Successful response. User was deleted, need to remove his session"
// @Failure 400 {object} errormiddleware.Error "Return's if user typed incorrect password"
// @Failure 401 {object} errormiddleware.Error "Return's if service can't authorize user"
// @Failure 404 {object} errormiddleware.Error "Return's if user is not authorized"
// @Failure 500 {object} errormiddleware.Error "Returns when there's some internal error that needs to be fixed or smtp server is not responding"
// @Failure 501 {object} errormiddleware.Error "Returns if query was incorrect"
// @Security ApiKeyAuth
// @Router /users/delete [delete]
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) error {
	var query model.DeleteUserQuery
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		return err
	}
	if err := h.Validator.Struct(query); err != nil {
		return mw.ValidationError(err.(validator.ValidationErrors), "wrong request")
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	h.Logger.Infof("received user deletion request")
	err := h.UserService.DeleteUser(ctx, &query)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// @Summary Confirm user's email
// @Description If code field is empty: send or resend confirmation message to user's email
// @Description Message can be resended every 1 minutes
// @Description If code field is not empty: validate code and approve email, code is expired within 10 minutes
// @Tags users
// @Produce json
// @Param code query string false "Confirmation code"
// @Success 200 "Successful response. Confirmation code was sent"
// @Success 204 "Successful response. Email was confirmed"
// @Failure 400 {object} errormiddleware.Error "Return's if user's email already confirmed"
// @Failure 401 {object} errormiddleware.Error "Return's if service can't authorize user"
// @Failure 403 {object} errormiddleware.Error "Return's if email can't be resend now (cooldown still active)"
// @Failure 404 {object} errormiddleware.Error "Return's if user is authorized, but service can't identity him"
// @Failure 500 {object} errormiddleware.Error "Returns when there's some internal error that needs to be fixed or smtp server is not responding"
// @Failure 501 {object} errormiddleware.Error "Returns when provided confirmation code is incorrect or code is expired"
// @Security ApiKeyAuth
// @Router /users/email [get]
func (h *Handler) EmailConfirmation(w http.ResponseWriter, r *http.Request) error {
	responseCode, err := h.UserService.UserEmailConfirmation(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		return err
	}
	if responseCode != 200 && responseCode != 204 {
		h.Logger.Errorf("user service returned invalid status code (%d) for email confirmation request", responseCode)
		return fmt.Errorf("service responded with invalid status code: %d", responseCode)
	}
	w.WriteHeader(responseCode)
	return nil
}

// @Summary Generate new token
// @Description Generate new token by provided refresh token
// @Description Refresh token stored in cache and expires in 7 days. If system was restarted, all tokens are cleared and sessions are deleted
// @Tags users
// @Produce json
// @Param query body jwt.RefreshTokenQuery true "Request query with user's refresh token"
// @Success 200 {object} jwt.JwtResponse "Successful response. Returns the same data as in authorization"
// @Failure 404 {object} errormiddleware.Error "Returns when service can't find user by provided credentials (user not found)"
// @Failure 500 {object} errormiddleware.Error "Returns when there's some internal error that needs to be fixed"
// @Failure 501 {object} errormiddleware.Error "Returns when provided data was not validated"
// @Router /users/refresh [post]
func (h *Handler) UpdateToken(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	defer r.Body.Close()
	var query jwt.RefreshTokenQuery
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		return err
	}
	if err := h.Validator.Struct(query); err != nil {
		return mw.ValidationError(err.(validator.ValidationErrors), "wrong token format")
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

// @Summary Finds user by id or login
// @Description Get user using Id or login, both params are optional, but one of them is necessary
// @Produce json
// @Tags users
// @Param id query string false "User id"
// @Param login query string false "User login"
// @Success 200 {object} model.User "Successful response"
// @Failure 400 {object} errormiddleware.Error "Returns when service didn't get a parameters"
// @Failure 404 {object} errormiddleware.Error "Returns when service can't find user by provided credentials (user not found)"
// @Failure 500 {object} errormiddleware.Error "Returns when there's some internal error that needs to be fixed"
// @Router /users [get]
func (h *Handler) FindUser(w http.ResponseWriter, r *http.Request) error {
	user_id := r.URL.Query().Get("id")
	user_login := r.URL.Query().Get("login")
	u, err := h.UserService.FindUser(r.Context(), user_id, user_login)
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

// @Summary Authenticates user
// @Description Finds user by login and password
// @Description Returns a token and refresh token. Token expires in 1 hour, refresh token expires in 7 days and stores in cache (removing after system restart)
// @Description Login field can be provided with user login or email
// @Produce json
// @Tags users
// @Param query body model.UserAuthQuery true "User credentials"
// @Success 200 {object} jwt.JwtResponse "Successful response. Returns user's login, roles and personal token and refresh token. Refresh token stores in cache"
// @Failure 404 {object} errormiddleware.Error "Returns when service can't find user by provided credentials (user not found)"
// @Failure 500 {object} errormiddleware.Error "Returns when there's some internal error that needs to be fixed"
// @Failure 501 {object} errormiddleware.Error "Returns when provided data was not validated"
// @Router /users/login [post]
func (h *Handler) Authenticate(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	defer r.Body.Close()
	var query model.UserAuthQuery
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		return err
	}
	if err := h.Validator.Struct(query); err != nil {
		return mw.ValidationError(err.(validator.ValidationErrors), "wrong query format")
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

// @Summary Register user
// @Description Creates a new instance of user and returns authorization principals
// @Produce json
// @Tags users
// @Param query body model.UserRegisterQuery true "User credentials"
// @Success 200 {object} jwt.JwtResponse "Successful token response. Returns the same response as in authorization"
// @Failure 409 {object} errormiddleware.Error "Returns when there's already exist user with provided login"
// @Failure 500 {object} errormiddleware.Error "Returns when there's some internal error that needs to be fixed"
// @Failure 501 {object} errormiddleware.Error "Returns when provided data was not validated"
// @Router /users/register [post]
func (h *Handler) UserRegister(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	defer r.Body.Close()
	var query model.UserRegisterQuery
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		return err
	}
	if err := h.Validator.Struct(query); err != nil {
		return mw.ValidationError(err.(validator.ValidationErrors), "wrong query format")
	}
	model, err := h.UserService.RegisterUser(r.Context(), &query)
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
