package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"auth-service/internal/config"
	httpdelivery "auth-service/internal/delivery/http"
	"auth-service/internal/delivery/http/handler"
	"auth-service/internal/infrastructure/cache"
	"auth-service/internal/infrastructure/database"
	"auth-service/internal/infrastructure/logger"
	"auth-service/internal/infrastructure/metrics"
	"auth-service/internal/repository/postgres"
	redisrepo "auth-service/internal/repository/redis"
	"auth-service/internal/usecase"
	"auth-service/pkg/utils"

	"go.uber.org/zap"
)

func main() {
	zapLogger, err := logger.NewZapLogger()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer zapLogger.Sync()

	cfg, err := config.Load()
	if err != nil {
		zapLogger.Fatal("configuration load failure", zap.Error(err))
	}

	keyManager, err := utils.NewKeyManager(cfg.RSAPrivateKeyPath)
	if err != nil {
		zapLogger.Fatal("failed to initialize RSA Key Manager", zap.Error(err))
	}

	auditLogger := logger.NewAuditLogger(zapLogger)
	metricsCollector := metrics.NewMetrics()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := database.RunMigrations(cfg); err != nil {
		zapLogger.Fatal("failed to execute migrations", zap.Error(err))
	}
	zapLogger.Info("database migrations verified")

	dbPool, err := database.NewPostgresPool(ctx, cfg)
	if err != nil {
		zapLogger.Fatal("failed to initialize postgres pool", zap.Error(err))
	}
	defer dbPool.Close()

	redisClient, err := cache.NewRedisClient(ctx, cfg)
	if err != nil {
		zapLogger.Fatal("failed to initialize redis client", zap.Error(err))
	}
	defer redisClient.Close()

	// Initialize Repositories
	userRepo := postgres.NewUserRepository(dbPool)
	sessionRepo := redisrepo.NewSessionRepository(redisClient)

	// Initialize Use Cases
	authUC := usecase.NewAuthUseCase(userRepo, sessionRepo, keyManager, auditLogger, cfg)
	userUC := usecase.NewUserUseCase(userRepo)

	// Initialize Handlers
	authHandler := handler.NewAuthHandler(authUC, keyManager)
	userHandler := handler.NewUserHandler(userUC)
	healthHandler := handler.NewHealthHandler(dbPool, redisClient)

	// Router Wiring
	router := httpdelivery.NewRouter(
		authHandler, userHandler, healthHandler,
		dbPool, redisClient, keyManager, metricsCollector,
		cfg, zapLogger,
	)

	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		zapLogger.Info("enterprise auth-service running", zap.String("port", cfg.AppPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("server crashed unexpectedly", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	zapLogger.Info("initiating graceful shutdown", zap.String("signal", sig.String()))

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		zapLogger.Fatal("forced server shutdown", zap.Error(err))
	}

	zapLogger.Info("graceful shutdown completed successfully")
}