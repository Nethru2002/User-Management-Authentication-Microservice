package http

import (
	"net/http"

	"auth-service/internal/config"
	"auth-service/internal/delivery/http/handler"
	"auth-service/internal/delivery/http/middleware"
	"auth-service/internal/domain"
	"auth-service/internal/infrastructure/metrics"
	"auth-service/pkg/utils"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func NewRouter(
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	healthHandler *handler.HealthHandler,
	scimHandler *handler.SCIMHandler,
	db *pgxpool.Pool,
	rdb *redis.Client,
	keyManager *utils.KeyManager,
	metricsCollector *metrics.Metrics,
	cfg *config.Config,
	logger *zap.Logger,
) http.Handler {
	mux := http.NewServeMux()

	// 1. Probes & Telemetry
	mux.HandleFunc("GET /healthz", healthHandler.Healthz)
	mux.HandleFunc("GET /readyz", healthHandler.Readyz)
	mux.Handle("GET /metrics", promhttp.Handler())
	mux.HandleFunc("GET /.well-known/jwks.json", authHandler.JWKS)

	// 2. Authentication & MFA
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", authHandler.Logout)
	mux.HandleFunc("POST /api/v1/auth/mfa/challenge", authHandler.VerifyMFA)

	// 3. Email Verification & Password Reset
	mux.HandleFunc("GET /api/v1/auth/verify-email", authHandler.VerifyEmail)
	mux.HandleFunc("POST /api/v1/auth/password/reset-request", authHandler.PasswordResetRequest)
	mux.HandleFunc("POST /api/v1/auth/password/reset-confirm", authHandler.PasswordResetConfirm)

	// 4. Federated Social Login (OAuth2)
	mux.HandleFunc("GET /api/v1/auth/oauth/url", authHandler.OAuthURL)
	mux.HandleFunc("GET /api/v1/auth/oauth/callback", authHandler.OAuthCallback)

	// 5. Authenticated Endpoints
	authMW := middleware.AuthMiddleware(keyManager)

	profileMux := http.NewServeMux()
	profileMux.HandleFunc("GET /api/v1/users/profile", userHandler.GetProfile)
	profileMux.HandleFunc("POST /api/v1/users/mfa/setup", authHandler.SetupMFA)
	profileMux.HandleFunc("POST /api/v1/users/mfa/confirm", authHandler.ConfirmMFASetup)
	mux.Handle("/api/v1/users/", authMW(middleware.RequirePermission(domain.PermissionReadProfile)(profileMux)))

	// 6. Admin Endpoints
	adminMux := http.NewServeMux()
	adminMux.HandleFunc("GET /api/v1/admin/users", userHandler.ListUsers)
	mux.Handle("/api/v1/admin/", authMW(middleware.RequirePermission(domain.PermissionManageSystem)(adminMux)))

	// 7. SCIM 2.0 Identity Provider Synchronization (RFC 7643 / 7644) using Go 1.22 Wildcards
	scimMux := http.NewServeMux()
	scimMux.HandleFunc("GET /scim/v2/ServiceProviderConfig", scimHandler.ServiceProviderConfig)
	scimMux.HandleFunc("GET /scim/v2/Users", scimHandler.ListUsers)
	scimMux.HandleFunc("POST /scim/v2/Users", scimHandler.CreateUser)
	scimMux.HandleFunc("GET /scim/v2/Users/{id}", scimHandler.GetUserByID)
	scimMux.HandleFunc("PUT /scim/v2/Users/{id}", scimHandler.UpdateUserByID)
	scimMux.HandleFunc("DELETE /scim/v2/Users/{id}", scimHandler.DeleteUserByID)
	mux.Handle("/scim/v2/", authMW(middleware.RequirePermission(domain.PermissionManageSystem)(scimMux)))

	// Pipeline Wrapping
	rateLimiter := middleware.NewRedisRateLimiter(rdb, 100)

	var handlerChain http.Handler = mux
	handlerChain = middleware.TenantMiddleware(handlerChain)
	handlerChain = middleware.MetricsMiddleware(metricsCollector)(handlerChain)
	handlerChain = middleware.LoggerMiddleware(logger)(handlerChain)
	handlerChain = rateLimiter.Middleware(handlerChain)
	handlerChain = middleware.CORSMiddleware(handlerChain)
	handlerChain = middleware.TracingMiddleware(handlerChain)
	handlerChain = middleware.RequestIDMiddleware(handlerChain)

	return handlerChain
}