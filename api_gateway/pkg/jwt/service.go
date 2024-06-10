package jwt

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cristalhq/jwt/v3"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/internal/client/user"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/cache"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
	valid "github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/validator"
)

type UserClaims struct {
	jwt.RegisteredClaims
	Login string   `json:"login"`
	Roles []string `json:"roles"`
	Email string   `json:"email"`
}

type RefreshTokenQuery struct {
	RefreshToken string `json:"refreshtoken" validate:"required"`
}

type jwtService struct {
	Logger    *logging.Logger
	Cache     cache.Cache
	Validator *valid.Validator
	secret    string
}
type JwtResponse struct {
	Login        string   `json:"login"`
	Roles        []string `json:"roles"`
	Token        string   `json:"token"`
	RefreshToken string   `json:"refreshtoken"`
}
type JwtService interface {
	Middleware(h http.HandlerFunc, roles ...string) http.HandlerFunc
	GenerateAccessToken(u *user.User) (*JwtResponse, error)
	UpdateRefreshToken(query *RefreshTokenQuery) (*JwtResponse, error)
}

func NewService(cache cache.Cache, logger *logging.Logger, validate *valid.Validator, secret string) JwtService {
	return &jwtService{Logger: logger, Cache: cache, Validator: validate, secret: secret}
}
func (j *jwtService) UpdateRefreshToken(rt *RefreshTokenQuery) (*JwtResponse, error) {
	if err := j.Validator.Struct(rt); err != nil {
		tp, ok := err.(validator.ValidationErrors)
		if ok {
			return nil, errormiddleware.ValidationError(tp, "wrong refresh token format")
		} else {
			return nil, errormiddleware.NotFoundError([]string{"wrong refresh token format"}, err.Error())
		}
	}
	defer j.Cache.Delete([]byte(rt.RefreshToken))

	userBytes, err := j.Cache.Get([]byte(rt.RefreshToken))
	if err != nil {
		j.Logger.Warn(err)
		return nil, errormiddleware.NotFoundError([]string{"couldn't get refresh token from cache"}, err.Error())
	}
	var u user.User
	err = json.Unmarshal(userBytes, &u)
	if err != nil {
		j.Logger.Error(err)
		return nil, err
	}
	return j.GenerateAccessToken(&u)
}
func (j *jwtService) GenerateAccessToken(u *user.User) (*JwtResponse, error) {
	signer, err := jwt.NewSignerHS(jwt.HS256, []byte(j.secret))
	if err != nil {
		j.Logger.Warn(err)
		return nil, err
	}
	builder := jwt.NewBuilder(signer)

	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        u.Id,
			Audience:  u.Roles,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 60)),
		},
		Roles: u.Roles,
		Login: u.Login,
		Email: u.Email,
	}
	token, err := builder.Build(claims)
	if err != nil {
		j.Logger.Warn(err)
		return nil, err
	}

	j.Logger.Info("creating refresh token...")
	refreshTokenUuid := uuid.New()
	userBytes, _ := json.Marshal(u)
	j.Cache.Set([]byte(refreshTokenUuid.String()), userBytes, int((7*24*time.Hour)/time.Second))

	responseToken := &JwtResponse{Login: u.Login, Roles: u.Roles, Token: token.String(), RefreshToken: refreshTokenUuid.String()}
	return responseToken, nil
}
