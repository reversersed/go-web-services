package jwt

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cristalhq/jwt/v3"
	"github.com/go-playground/validator/v10"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/internal/client/user"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/cache"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
	valid "github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/validator"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	TokenCookieName   string = "authTokenCookie"
	RefreshCookieName string = "refreshTokenCookie"
)

type UserClaims struct {
	jwt.RegisteredClaims
	Login string   `json:"login"`
	Roles []string `json:"roles"`
	Email string   `json:"email"`
}

type jwtService struct {
	Logger    *logging.Logger
	Cache     cache.Cache
	Validator *valid.Validator
	secret    string
}

func NewService(cache cache.Cache, logger *logging.Logger, validate *valid.Validator, secret string) *jwtService {
	return &jwtService{Logger: logger, Cache: cache, Validator: validate, secret: secret}
}
func (j *jwtService) UpdateRefreshToken(refreshToken string) (*user.JwtResponse, error) {
	if err := j.Validator.Var(refreshToken, "primitiveid"); err != nil {
		return nil, errormiddleware.ValidationError(err.(validator.ValidationErrors), "wrong refresh token format")
	}
	defer j.Cache.Delete([]byte(refreshToken))

	userBytes, err := j.Cache.Get([]byte(refreshToken))
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
func (j *jwtService) GenerateAccessToken(u *user.User) (*user.JwtResponse, error) {
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
	refreshTokenUuid := primitive.NewObjectID().Hex()
	userBytes, _ := json.Marshal(u)
	j.Cache.Set([]byte(refreshTokenUuid), userBytes, int((7*24*time.Hour)/time.Second))

	responseToken := &user.JwtResponse{
		Login: u.Login,
		Roles: u.Roles,
		Token: &http.Cookie{
			Name:     TokenCookieName,
			Value:    token.String(),
			MaxAge:   (int)((31 * 24 * time.Hour) / time.Second),
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
		},
		RefreshToken: &http.Cookie{
			Name:     RefreshCookieName,
			Value:    refreshTokenUuid,
			MaxAge:   (int)((31 * 24 * time.Hour) / time.Second),
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
		},
	}
	return responseToken, nil
}
