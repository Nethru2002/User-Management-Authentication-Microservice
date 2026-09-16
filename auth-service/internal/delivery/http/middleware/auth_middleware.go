package middleware

import (
	"context"
	"net/http"
	"strings"

	"auth-service/pkg/utils"
)

type contextKey string

const UserContextKey contextKey = "user"

func AuthMiddleware(keyManager *utils.KeyManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				utils.WriteError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				utils.WriteError(w, http.StatusUnauthorized, "invalid authorization header format: must be Bearer <token>")
				return
			}

			claims, err := keyManager.ValidateAccessToken(parts[1])
			if err != nil {
				utils.WriteError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}