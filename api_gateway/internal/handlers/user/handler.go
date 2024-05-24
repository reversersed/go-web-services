package user

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/internal/client/jwt"
	model "github.com/reversersed/go-web-services/tree/main/api_gateway/internal/client/user"
	mw "github.com/reversersed/go-web-services/tree/main/api_gateway/internal/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
)

const (
	url_auth     = "/api/v1/users/login"
	url_refresh  = "/api/v1/users/refresh"
	url_register = "/api/v1/users/register"
)

type UserService interface {
	AuthByLoginAndPassword(ctx context.Context, query *model.UserAuthQuery) (*model.User, error)
	RegisterUser(ctx context.Context, query *model.UserRegisterQuery) (*model.User, error)
}

type Handler struct {
	Logger      *logging.Logger
	JwtService  jwt.JwtService
	UserService UserService
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, url_auth, mw.Middleware(h.Authenticate))
	router.HandlerFunc(http.MethodPost, url_refresh, mw.Middleware(h.UpdateToken))
	router.HandlerFunc(http.MethodPost, url_register, mw.Middleware(h.UserRegister))
	h.Logger.Info("auth service registered")
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
	if err := validator.New().Struct(query); err != nil {
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
	if err := validator.New().Struct(query); err != nil {
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
	if err := validator.New().Struct(query); err != nil {
		return mw.ValidationError(err.(validator.ValidationErrors), "wrong query format")
	}
	errs := h.passwordValidation(query.Password)
	if errs != nil {
		return mw.ValidationErrorByString(errs, "wrong password format")
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
func (h *Handler) passwordValidation(password string) []string {
	var errors []string

	//lowercase character
	mathed, err := regexp.MatchString("[a-z]+", password)
	if err != nil {
		h.Logger.Errorf("cant check user password: %v", err)
		return []string{"internal error. check logs"}
	}
	if !mathed {
		errors = append(errors, "password must contain at least one lowercase letter")
	}
	//uppercase character
	mathed, err = regexp.MatchString("[A-Z]+", password)
	if err != nil {
		h.Logger.Errorf("cant check user password: %v", err)
		return []string{"internal error. check logs"}
	}
	if !mathed {
		errors = append(errors, "password must contain at least one uppercase letter")
	}
	//one digit
	mathed, err = regexp.MatchString("[0-9]+", password)
	if err != nil {
		h.Logger.Errorf("cant check user password: %v", err)
		return []string{"internal error. check logs"}
	}
	if !mathed {
		errors = append(errors, "password must contain at least one digit")
	}
	//special symbol
	mathed, err = regexp.MatchString("[!@#\\$%\\^&*()_\\+-.,]+", password)
	if err != nil {
		h.Logger.Errorf("cant check user password: %v", err)
		return []string{"internal error. check logs"}
	}
	if !mathed {
		errors = append(errors, "password must contain at least one special character")
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}
