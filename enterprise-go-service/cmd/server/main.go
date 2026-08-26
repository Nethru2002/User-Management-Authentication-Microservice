package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpdelivery "enterprise-go-service/internal/delivery/http"
	"enterprise-go-service/internal/delivery/http/handler"
	"enterprise-go-service/internal/infrastructure/cache"
	"enterprise-go-service/internal/infrastructure/database"
	"enterprise-go-service/internal/infrastructure/logger"
	"enterprise-go-service/internal/repository/postgres"
	"enterprise-go-service/internal/usecase"

	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	zapLogger, err := logger.NewZapLogger()
	if err != nil {
		log.Fatalf("failed to initialize zap logger: %v", err)
	}
	defer zapLogger.Sync()

	dbPool, err := database.NewPostgresPool(ctx)
	if err != nil {
		zapLogger.Fatal("failed to connect to postgres", zap.Error(err))
	}
	defer dbPool.Close()

	redisClient, err := cache.NewRedisClient(ctx)
	if err != nil {
		zapLogger.Fatal("failed to connect to redis", zap.Error(err))
	}
	defer redisClient.Close()

	userRepo := postgres.NewUserRepository(dbPool)

	authUC := usecase.NewAuthUseCase(userRepo)
	userUC := usecase.NewUserUseCase(userRepo)

	authHandler := handler.NewAuthHandler(authUC)
	userHandler := handler.NewUserHandler(userUC)

	router := httpdelivery.NewRouter(authHandler, userHandler, zapLogger)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		zapLogger.Info("starting server", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("server startup failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLogger.Info("shutting down server gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		zapLogger.Fatal("forced server shutdown", zap.Error(err))
	}

	zapLogger.Info("server exited properly")
}