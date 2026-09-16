package middleware

import (
	"context"
	"net/http"

	"auth-service/internal/domain"

	"github.com/google/uuid"
)

type tenantContextKeyType struct{}

var TenantContextKey = tenantContextKeyType{}

func TenantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantHeader := r.Header.Get("X-Tenant-ID")
		tenantID := domain.DefaultTenantID

		if tenantHeader != "" {
			if parsed, err := uuid.Parse(tenantHeader); err == nil {
				tenantID = parsed
			}
		}

		ctx := context.WithValue(r.Context(), TenantContextKey, tenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}