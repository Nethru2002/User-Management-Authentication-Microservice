package middleware

import (
	"net/http"

	"enterprise-go-service/internal/domain"
	"enterprise-go-service/pkg/utils"
)

func RequirePermission(permission domain.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(UserContextKey).(*domain.JwtCustomClaims)
			if !ok || claims == nil {
				utils.WriteError(w, http.StatusUnauthorized, "unauthorized: session context missing")
				return
			}

			if !domain.HasPermission(claims.Role, permission) {
				utils.WriteError(w, http.StatusForbidden, "forbidden: insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}