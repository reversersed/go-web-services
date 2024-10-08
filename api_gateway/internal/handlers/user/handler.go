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
	url_auth              = "/api/v1/users/auth"
	url_register          = "/api/v1/users/register"
	url_confirm_email     = "/api/v1/users/email"
	url_find_user         = "/api/v1/users"
	url_delete_user       = "/api/v1/users/delete"
	url_update_user_login = "/api/v1/users/changename"
)

//go:generate mockgen -source=handler.go -destination=mocks/service_mock.go

type UserService interface {
	AuthByLoginAndPassword(ctx context.Context, query *model.UserAuthQuery) (*model.User, error)
	RegisterUser(ctx context.Context, query *model.UserRegisterQuery) (*model.User, error)
	UserEmailConfirmation(ctx context.Context, code string) (int, error)
	FindUser(ctx context.Context, userid string, login string) (*model.User, error)
	DeleteUser(ctx context.Context, query *model.DeleteUserQuery) error
	UpdateUserLogin(ctx context.Context, query *model.UpdateUserLoginQuery) (*model.User, error)
}
type JwtService interface {
	Middleware(h http.HandlerFunc, roles ...string) http.HandlerFunc
	GenerateAccessToken(u *model.User) (*model.JwtResponse, error)
	GetUserClaims(token string) (*model.JwtResponse, error)
}
type Handler struct {
	Logger      *logging.Logger
	JwtService  JwtService
	UserService UserService
	Validator   *valid.Validator
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, url_auth, h.Logger.Middleware(mw.Middleware(h.Authenticate)))
	router.HandlerFunc(http.MethodGet, url_auth, h.JwtService.Middleware(h.Logger.Middleware(mw.Middleware(h.Authorize))))
	router.HandlerFunc(http.MethodPost, url_register, h.Logger.Middleware(mw.Middleware(h.UserRegister)))
	router.HandlerFunc(http.MethodPost, url_confirm_email, h.JwtService.Middleware(h.Logger.Middleware(mw.Middleware(h.EmailConfirmation))))
	router.HandlerFunc(http.MethodGet, url_find_user, h.Logger.Middleware(mw.Middleware(h.FindUser)))
	router.HandlerFunc(http.MethodDelete, url_delete_user, h.JwtService.Middleware(h.Logger.Middleware(mw.Middleware(h.DeleteUser))))
	router.HandlerFunc(http.MethodPatch, url_update_user_login, h.JwtService.Middleware(h.Logger.Middleware(mw.Middleware(h.UpdateUserLogin))))
	h.Logger.Info("auth handlers registered")
}

// @Summary Update user's login
// @Description New login must be unique. Login changing are available only 1 time per month
// @Tags users
// @Produce json
// @Param NewLogin body model.UpdateUserLoginQuery true "New user login. Must be unique"
// @Success 200 {object} model.JwtResponse "Successful response. User's login was updated"
// @Failure 401 {object} errormiddleware.Error "Return's if service can't authorize user"
// @Failure 403 {object} errormiddleware.Error "Return's if user has login changing cooldown"
// @Failure 404 {object} errormiddleware.Error "Return's if user is not authorized"
// @Failure 409 {object} errormiddleware.Error "Return's if new user's login already taken"
// @Failure 500 {object} errormiddleware.Error "Returns when there's some internal error that needs to be fixed or smtp server is not responding"
// @Failure 501 {object} errormiddleware.Error "Returns if query was incorrect"
// @Security ApiKeyAuth
// @Router /users/changename [patch]
func (h *Handler) UpdateUserLogin(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	var query model.UpdateUserLoginQuery
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		return err
	}
	if err := h.Validator.Struct(query); err != nil {
		return mw.ValidationError(err.(validator.ValidationErrors), "wrong request")
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	user, err := h.UserService.UpdateUserLogin(ctx, &query)
	if err != nil {
		return err
	}
	token, err := h.JwtService.GenerateAccessToken(user)
	if err != nil {
		h.Logger.Warn(err)
		return err
	}
	data, _ := json.Marshal(token)

	http.SetCookie(w, token.Token)
	http.SetCookie(w, token.RefreshToken)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return nil
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
// @Failure 404 {object} errormiddleware.Error "Return's if service can't find user's code or code is expired"
// @Failure 500 {object} errormiddleware.Error "Returns when there's some internal error that needs to be fixed or smtp server is not responding"
// @Security ApiKeyAuth
// @Router /users/email [post]
func (h *Handler) EmailConfirmation(w http.ResponseWriter, r *http.Request) error {
	responseCode, err := h.UserService.UserEmailConfirmation(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		return err
	}
	if responseCode != http.StatusOK && responseCode != http.StatusNoContent {
		h.Logger.Errorf("user service returned invalid status code (%d) for email confirmation request", responseCode)
		return fmt.Errorf("service responded with invalid status code: %d", responseCode)
	}
	w.WriteHeader(responseCode)
	return nil
}

// @Summary Finds user by id or login
// @Description Get user using Id or login, both params are optional, but one of them is necessary
// @Produce json
// @Tags users
// @Param id query string false "User id"
// @Param login query string false "User login" example(admin)
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
	bytes, _ := json.Marshal(u)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
	return nil
}

// @Summary Authorizes user
// @Description Authorizes user's credentials by token. This needs to check if user's token is valid or get current authenticated user
// @Produce json
// @Tags users
// @Success 200 {object} model.JwtResponse "Successful response. Returns user's login, roles and personal token and refresh token. Refresh token stores in cache"
// @Failure 401 {object} errormiddleware.Error "Returns if user not authorized"
// @Failure 500 {object} errormiddleware.Error "Returns when there's some internal error that needs to be fixed"
// @Security ApiKeyAuth
// @Router /users/auth [get]
func (h *Handler) Authorize(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	tokenString, err := r.Cookie(jwt.TokenCookieName)
	if err != nil {
		return mw.UnauthorizedError([]string{"user not authorized"}, err.Error())
	}

	token, err := h.JwtService.GetUserClaims(tokenString.Value)
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
// @Description Sets token to cookies
// @Description Login field can be provided with user login or email
// @Produce json
// @Tags users
// @Param query body model.UserAuthQuery true "User credentials"
// @Success 200 {object} model.JwtResponse "Successful response. Returns user's login and roles
// @Failure 404 {object} errormiddleware.Error "Returns when service can't find user by provided credentials (user not found)"
// @Failure 500 {object} errormiddleware.Error "Returns when there's some internal error that needs to be fixed"
// @Failure 501 {object} errormiddleware.Error "Returns when provided data was not validated"
// @Router /users/auth [post]
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

	http.SetCookie(w, token.Token)
	http.SetCookie(w, token.RefreshToken)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return nil
}

// @Summary Register user
// @Description Creates a new instance of user and returns authorization principals. Sets the token cookies
// @Produce json
// @Tags users
// @Param query body model.UserRegisterQuery true "User credentials"
// @Success 200 {object} model.JwtResponse "Successful token response. Returns the same response as in authorization"
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

	http.SetCookie(w, token.Token)
	http.SetCookie(w, token.RefreshToken)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return nil
}
