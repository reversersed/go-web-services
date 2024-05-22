package jwt

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cristalhq/jwt/v3"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/internal/config"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
)

func Middleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.GetLogger()
		header := strings.Split(r.Header.Get("Authorization"), "Bearer ")
		if len(header) != 2 {
			logger.Warnf("Wrong provided token: %s", r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Wrong token form provided"))
			return
		}
		headertoken := header[1]
		key := []byte(config.GetConfig().SecretToken)
		verifier, err := jwt.NewVerifierHS(jwt.HS256, key)
		if err != nil {
			unauthorized(w, err)
			return
		}
		logger.Info("parsing and verifying token...")
		token, err := jwt.ParseAndVerifyString(headertoken, verifier)
		if err != nil {
			unauthorized(w, err)
			return
		}

		var claims UserClaims
		err = json.Unmarshal(token.RawClaims(), &claims)
		if err != nil {
			unauthorized(w, err)
			return
		}
		if valid := claims.IsValidAt(time.Now()); !valid {
			unauthorized(w, fmt.Errorf("token has been expired: %s", err))
			return
		}

		ctx := context.WithValue(r.Context(), UserClaimKey, claims.Login)
		h(w, r.WithContext(ctx))
	}
}
func unauthorized(w http.ResponseWriter, err error) {
	logging.GetLogger().Warn(err)
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("unauthorized due to error. check logs"))
}
