package jwt

import (
	"encoding/json"
	"time"

	"github.com/cristalhq/jwt/v3"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/internal/client/user"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/internal/config"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/internal/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/cache"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
)

type key int

const (
	UserClaimKey key = iota
)

type UserClaims struct {
	jwt.RegisteredClaims
	Login string   `json:"login"`
	Roles []string `json:"roles"`
	Email string   `json:"email"`
}

type RefreshTokenQuery struct {
	RefreshToken string `json:"refreshtoken" validate:"required,jwt"`
}

type jwtService struct {
	Logger *logging.Logger
	Cache  cache.Cache
}
type JwtResponse struct {
	Login        string   `json:"login"`
	Roles        []string `json:"roles"`
	Token        string   `json:"token"`
	RefreshToken string   `json:"refreshtoken"`
}
type JwtService interface {
	GenerateAccessToken(u *user.User) (*JwtResponse, error)
	UpdateRefreshToken(query *RefreshTokenQuery) (*JwtResponse, error)
}

func NewService(cache cache.Cache, logger *logging.Logger) JwtService {
	return &jwtService{Logger: logger, Cache: cache}
}
func (j *jwtService) UpdateRefreshToken(rt *RefreshTokenQuery) (*JwtResponse, error) {
	if err := validator.New().Struct(rt); err != nil {
		tp, ok := err.(validator.ValidationErrors)
		if ok {
			return nil, errormiddleware.ValidationError(tp, "wrong refresh token format")
		} else {
			return nil, errormiddleware.NotFoundError("wrong refresh token format", []string{err.Error()})
		}
	}
	defer j.Cache.Delete([]byte(rt.RefreshToken))

	userBytes, err := j.Cache.Get([]byte(rt.RefreshToken))
	if err != nil {
		j.Logger.Warn(err)
		return nil, errormiddleware.NotFoundError("couldn't get refresh token from cache", []string{err.Error()})
	}
	var u user.User
	err = json.Unmarshal(userBytes, &u)
	if err != nil {
		j.Logger.Warn(err)
		return nil, err
	}
	return j.GenerateAccessToken(&u)
}
func (j *jwtService) GenerateAccessToken(u *user.User) (*JwtResponse, error) {
	key := []byte(config.GetConfig().SecretToken)
	signer, err := jwt.NewSignerHS(jwt.HS256, key)
	if err != nil {
		j.Logger.Warn(err)
		return nil, err
	}
	builder := jwt.NewBuilder(signer)

	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        u.Id,
			Audience:  []string{"users"},
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
	err = j.Cache.Set([]byte(refreshTokenUuid.String()), userBytes, int(7*24*time.Hour))
	if err != nil {
		j.Logger.Warn(err)
		return nil, err
	}
	responseToken := &JwtResponse{Login: u.Login, Roles: u.Roles, Token: token.String(), RefreshToken: refreshTokenUuid.String()}
	return responseToken, nil
}
