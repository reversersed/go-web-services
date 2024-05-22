package jwt

import (
	"encoding/json"
	"time"

	"github.com/cristalhq/jwt/v3"
	"github.com/google/uuid"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/internal/client/auth"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/internal/config"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/cache"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
)

type key int

const (
	UserClaimKey key = iota
)

type UserClaims struct {
	jwt.RegisteredClaims
	Login string `json:"login"`
}

type RefreshTokenQuery struct {
	RefreshToken string `json:"refreshtoken" validate:"required"`
}

type jwtService struct {
	Logger *logging.Logger
	Cache  cache.Cache
}

type JwtService interface {
	GenerateAccessToken(u *auth.User) ([]byte, error)
	UpdateRefreshToken(query *RefreshTokenQuery) ([]byte, error)
}

func NewService(cache cache.Cache, logger *logging.Logger) JwtService {
	return &jwtService{Logger: logger, Cache: cache}
}
func (j *jwtService) UpdateRefreshToken(rt *RefreshTokenQuery) ([]byte, error) {
	defer j.Cache.Delete([]byte(rt.RefreshToken))

	userBytes, err := j.Cache.Get([]byte(rt.RefreshToken))
	if err != nil {
		j.Logger.Warn(err)
		return nil, errormiddleware.NotFoundError(err.Error(), "couldn't get refresh token from cache")
	}
	var u auth.User
	err = json.Unmarshal(userBytes, &u)
	if err != nil {
		j.Logger.Warn(err)
		return nil, err
	}
	return j.GenerateAccessToken(&u)
}
func (j *jwtService) GenerateAccessToken(u *auth.User) ([]byte, error) {
	key := []byte(config.GetConfig().SecretToken)
	signer, err := jwt.NewSignerHS(jwt.HS256, key)
	if err != nil {
		j.Logger.Warn(err)
		return nil, err
	}
	builder := jwt.NewBuilder(signer)

	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			//ID:        u.UUID,
			Audience:  []string{"users"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 60)),
		},
		Login: u.Login,
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

	jsonBytes, err := json.Marshal(map[string]string{
		"token":        token.String(),
		"refreshtoken": refreshTokenUuid.String(),
	})
	if err != nil {
		j.Logger.Warn(err)
		return nil, err
	}

	return jsonBytes, nil
}
