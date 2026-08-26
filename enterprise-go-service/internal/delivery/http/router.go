package http

import (
	"net/http"

	"enterprise-go-service/internal/delivery/http/handler"
	"enterprise-go-service/internal/delivery/http/middleware"

	"go.uber.org/zap"
)

func NewRouter(authHandler *handler.AuthHandler, userHandler *handler.UserHandler, logger *zap.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)

	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("GET /api/v1/users/profile", userHandler.GetProfile)

	mux.Handle("/api/v1/users/", middleware.AuthMiddleware(protectedMux))

	handlerWithLog := middleware.LoggerMiddleware(logger)(mux)
	return handlerWithLog
}