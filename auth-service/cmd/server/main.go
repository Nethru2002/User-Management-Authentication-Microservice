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

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// 1. Automatically load .env file if it exists locally
	_ = godotenv.Load()

	// 2. Initialize Structured Production Zap Logger
	zapLogger, err := logger.NewZapLogger()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer zapLogger.Sync()

	// 3. Load and validate strongly typed configuration
	cfg, err := config.Load()
	if err != nil {
		zapLogger.Fatal("configuration error", zap.Error(err))
	}

	// 4. Initialize Asymmetric Key Manager (RS256 / RSA-4096)
	keyManager, err := utils.NewKeyManager(cfg.RSAPrivateKeyPath)
	if err != nil {
		zapLogger.Fatal("failed to initialize RSA Key Manager", zap.Error(err))
	}

	// 5. Initialize Observability & SRE Tools
	auditLogger := logger.NewAuditLogger(zapLogger)
	emailSender := utils.NewConsoleEmailSender(zapLogger)
	metricsCollector := metrics.NewMetrics()

	// 6. Apply Database Migrations (PostgreSQL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := database.RunMigrations(cfg); err != nil {
		zapLogger.Fatal("database migration failure", zap.Error(err))
	}
	zapLogger.Info("database migrations verified and applied")

	// 7. Connect to PostgreSQL Pool
	dbPool, err := database.NewPostgresPool(ctx, cfg)
	if err != nil {
		zapLogger.Fatal("failed to connect to postgres pool", zap.Error(err))
	}
	defer dbPool.Close()

	// 8. Connect to Redis Client
	redisClient, err := cache.NewRedisClient(ctx, cfg)
	if err != nil {
		zapLogger.Fatal("failed to connect to redis", zap.Error(err))
	}
	defer redisClient.Close()

	// 9. Dependency Injection: Repositories
	userRepo := postgres.NewUserRepository(dbPool)
	sessionRepo := redisrepo.NewSessionRepository(redisClient)

	// 10. Dependency Injection: Use Cases
	authUC := usecase.NewAuthUseCase(userRepo, sessionRepo, keyManager, emailSender, auditLogger, cfg)
	userUC := usecase.NewUserUseCase(userRepo)
	scimUC := usecase.NewSCIMUseCase(userRepo)

	// 11. Dependency Injection: Handlers
	authHandler := handler.NewAuthHandler(authUC, keyManager)
	userHandler := handler.NewUserHandler(userUC)
	healthHandler := handler.NewHealthHandler(dbPool, redisClient)
	scimHandler := handler.NewSCIMHandler(scimUC)

	// 12. Build HTTP Master Router & Middleware Pipeline
	router := httpdelivery.NewRouter(
		authHandler, userHandler, healthHandler, scimHandler,
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

	// 13. Start HTTP Server Non-blocking
	go func() {
		zapLogger.Info("enterprise auth-service running", zap.String("port", cfg.AppPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("server crashed unexpectedly", zap.Error(err))
		}
	}()

	// 14. Graceful Shutdown Signal Interception
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	zapLogger.Info("initiating graceful shutdown", zap.String("signal", sig.String()))

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		zapLogger.Fatal("forced shutdown error", zap.Error(err))
	}

	zapLogger.Info("server exited safely")
}