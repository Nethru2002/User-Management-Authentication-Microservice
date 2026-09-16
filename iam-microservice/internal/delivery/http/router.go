package http

import (
	"encoding/json"
	"net/http"

	"enterprise-go-service/internal/delivery/http/handler"
	"enterprise-go-service/internal/delivery/http/middleware"
	"enterprise-go-service/internal/domain"

	"go.uber.org/zap"
)

func NewRouter(authHandler *handler.AuthHandler, userHandler *handler.UserHandler, logger *zap.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	})

	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)

	profileMux := http.NewServeMux()
	profileMux.HandleFunc("GET /api/v1/users/profile", userHandler.GetProfile)
	mux.Handle("/api/v1/users/profile", middleware.AuthMiddleware(middleware.RequirePermission(domain.PermissionReadProfile)(profileMux)))

	adminMux := http.NewServeMux()
	adminMux.HandleFunc("GET /api/v1/admin/users", userHandler.ListUsers)
	mux.Handle("/api/v1/admin/", middleware.AuthMiddleware(middleware.RequirePermission(domain.PermissionManageSystem)(adminMux)))

	rateLimiter := middleware.NewRateLimiter()

	var handlerChain http.Handler = mux
	handlerChain = middleware.LoggerMiddleware(logger)(handlerChain)
	handlerChain = rateLimiter.Middleware(handlerChain)
	handlerChain = middleware.CORSMiddleware(handlerChain)
	handlerChain = middleware.RequestIDMiddleware(handlerChain)

	return handlerChain
}