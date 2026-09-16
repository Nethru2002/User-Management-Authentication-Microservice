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
	db *pgxpool.Pool,
	rdb *redis.Client,
	keyManager *utils.KeyManager,
	metricsCollector *metrics.Metrics,
	cfg *config.Config,
	logger *zap.Logger,
) http.Handler {
	mux := http.NewServeMux()

	// 1. Kubernetes Health & Readiness Probes
	mux.HandleFunc("GET /healthz", healthHandler.Healthz)
	mux.HandleFunc("GET /readyz", healthHandler.Readyz)

	// 2. Prometheus Metrics & RFC 7517 Public JWKS
	mux.Handle("GET /metrics", promhttp.Handler())
	mux.HandleFunc("GET /.well-known/jwks.json", authHandler.JWKS)

	// 3. Public Authentication Routes
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", authHandler.Logout)

	// 4. Authenticated RBAC Protected Routes
	authMW := middleware.AuthMiddleware(keyManager)

	profileMux := http.NewServeMux()
	profileMux.HandleFunc("GET /api/v1/users/profile", userHandler.GetProfile)
	mux.Handle("/api/v1/users/profile", authMW(middleware.RequirePermission(domain.PermissionReadProfile)(profileMux)))

	adminMux := http.NewServeMux()
	adminMux.HandleFunc("GET /api/v1/admin/users", userHandler.ListUsers)
	mux.Handle("/api/v1/admin/", authMW(middleware.RequirePermission(domain.PermissionManageSystem)(adminMux)))

	// 5. Middleware Pipeline
	rateLimiter := middleware.NewRedisRateLimiter(rdb, 100)

	var handlerChain http.Handler = mux
	handlerChain = middleware.MetricsMiddleware(metricsCollector)(handlerChain)
	handlerChain = middleware.LoggerMiddleware(logger)(handlerChain)
	handlerChain = rateLimiter.Middleware(handlerChain)
	handlerChain = middleware.CORSMiddleware(handlerChain)
	handlerChain = middleware.TracingMiddleware(handlerChain)
	handlerChain = middleware.RequestIDMiddleware(handlerChain)

	return handlerChain
}