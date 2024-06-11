package jwt

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/cristalhq/jwt/v3"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/rest"
)

func (s *jwtService) Middleware(h http.HandlerFunc, roles ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := strings.Split(r.Header.Get("Authorization"), "Bearer ")
		if len(header) != 2 {
			s.Logger.Warnf("Wrong provided token: %s", r.Header.Get("Authorization"))
			unauthorized(w, fmt.Errorf("wrong token form provided"))
			return
		}
		headertoken := header[1]
		key := []byte(s.secret)
		verifier, err := jwt.NewVerifierHS(jwt.HS256, key)
		if err != nil {
			s.Logger.Errorf("error creating verifier for key. key length = %d, error = %v", len(key), err)
			unauthorized(w, err)
			return
		}
		s.Logger.Info("parsing and verifying token...")
		token, err := jwt.ParseAndVerifyString(headertoken, verifier)
		if err != nil {
			unauthorized(w, err)
			return
		}

		var claims UserClaims
		if err := json.Unmarshal(token.RawClaims(), &claims); err != nil {
			unauthorized(w, err)
			return
		}
		if !claims.IsValidAt(time.Now()) {
			unauthorized(w, fmt.Errorf("token has been expired"))
			return
		}
		if len(roles) > 0 {
			var errorRoles []string
			for _, role := range roles {
				if len(role) > 0 && !slices.Contains(claims.Roles, role) {
					errorRoles = append(errorRoles, fmt.Sprintf("user has no %s right", role))
				}
			}
			if len(errorRoles) > 0 {
				forbidden(w, errorRoles)
				return
			}
		}
		ctx := context.WithValue(r.Context(), rest.UserIdKey, claims.ID)
		h(w, r.WithContext(ctx))
	}
}
func unauthorized(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Write(errormiddleware.UnauthorizedError([]string{err.Error()}, "unauthorized due to error, check logs").Marshall())
}
func forbidden(w http.ResponseWriter, errors []string) {
	w.WriteHeader(http.StatusForbidden)
	w.Write(errormiddleware.ForbiddenError(errors, "user rights forbidden").Marshall())
}
