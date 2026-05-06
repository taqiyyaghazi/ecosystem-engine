package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	authDelivery "github.com/taqiyyaghazi/ecosystem-engine/internal/features/auth/delivery"
	authRepository "github.com/taqiyyaghazi/ecosystem-engine/internal/features/auth/repository"
	authUsecase "github.com/taqiyyaghazi/ecosystem-engine/internal/features/auth/usecase"
	discoveryDelivery "github.com/taqiyyaghazi/ecosystem-engine/internal/features/discovery/delivery"
	discoveryRepository "github.com/taqiyyaghazi/ecosystem-engine/internal/features/discovery/repository"
	discoveryUsecase "github.com/taqiyyaghazi/ecosystem-engine/internal/features/discovery/usecase"
	leaderboardDelivery "github.com/taqiyyaghazi/ecosystem-engine/internal/features/leaderboard/delivery"
	leaderboardRepository "github.com/taqiyyaghazi/ecosystem-engine/internal/features/leaderboard/repository"
	leaderboardUsecase "github.com/taqiyyaghazi/ecosystem-engine/internal/features/leaderboard/usecase"
	notificationRepository "github.com/taqiyyaghazi/ecosystem-engine/internal/features/notifications/repository"
	notificationUsecase "github.com/taqiyyaghazi/ecosystem-engine/internal/features/notifications/usecase"
	serviceDelivery "github.com/taqiyyaghazi/ecosystem-engine/internal/features/services/delivery"
	serviceRepository "github.com/taqiyyaghazi/ecosystem-engine/internal/features/services/repository"
	serviceUsecase "github.com/taqiyyaghazi/ecosystem-engine/internal/features/services/usecase"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/platform/cache"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/platform/config"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/platform/database"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/platform/logger"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/platform/middleware/ratelimit"
)

const (
	defaultReadTimeout     = 5 * time.Second
	defaultWriteTimeout    = 10 * time.Second
	defaultIdleTimeout     = 120 * time.Second
	defaultShutdownTimeout = 10 * time.Second
	defaultConnTimeout     = 10 * time.Second
)

func main() {
	if err := run(); err != nil {
		slog.Error("Application failed to start", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	logger.SetupLogger(cfg.AppEnv)

	dbCtx, dbCancel := context.WithTimeout(context.Background(), defaultConnTimeout)
	defer dbCancel()

	dbPool, err := database.NewPostgresPool(dbCtx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer func() {
		dbPool.Close()
		slog.Info("Database connection closed")
	}()

	rdbCtx, rdbCancel := context.WithTimeout(context.Background(), defaultConnTimeout)
	defer rdbCancel()

	rdb, err := cache.NewRedisClient(rdbCtx, cfg.RedisURL, cfg.RedisPass, cfg.RedisDB)
	if err != nil {
		return fmt.Errorf("connecting to redis: %w", err)
	}
	defer func() {
		_ = rdb.Close()
		slog.Info("Redis connection closed")
	}()

	// Notifications feature (Phase 7)
	notificationRepo := notificationRepository.NewNotificationRepository(dbPool, rdb)
	notificationUseCase := notificationUsecase.NewNotificationUsecase(notificationRepo)

	// Services feature (Phase 1)
	serviceRepo := serviceRepository.NewServiceRepository(dbPool, rdb)
	partnerRepo := serviceRepository.NewPartnerRepository(dbPool)
	serviceUseCase := serviceUsecase.NewServiceUsecase(serviceRepo)
	partnerUseCase := serviceUsecase.NewPartnerUsecase(serviceRepo, partnerRepo, notificationUseCase)
	serviceHandler := serviceDelivery.NewServiceHandler(serviceUseCase, partnerUseCase)

	// Auth feature (Phase 2)
	userRepo := authRepository.NewUserRepository(dbPool)
	sessionRepo := authRepository.NewSessionRepository(rdb)
	authUseCase := authUsecase.NewAuthUsecase(userRepo, sessionRepo)
	authHandler := authDelivery.NewAuthHandler(authUseCase)
	authMiddleware := authDelivery.RequireAuth(authUseCase)

	// Rate Limiter (Phase 3)
	rateLimiter := ratelimit.NewRateLimiter(rdb)

	// Discovery feature (Phase 4)
	discoveryRepo := discoveryRepository.NewDiscoveryRepository(rdb, dbPool)
	discoveryUseCase := discoveryUsecase.NewDiscoveryUsecase(discoveryRepo)
	discoveryHandler := discoveryDelivery.NewDiscoveryHandler(discoveryUseCase)

	// Start background worker for stale data management
	discoveryUseCase.StartCleanupWorker(context.Background())

	// Leaderboard feature (Phase 5)
	leaderboardRepo := leaderboardRepository.NewLeaderboardRepository(dbPool, rdb)
	leaderboardUseCase := leaderboardUsecase.NewLeaderboardUseCase(leaderboardRepo, notificationUseCase)
	leaderboardHandler := leaderboardDelivery.NewLeaderboardHandler(leaderboardUseCase)

	router := setupRouter(cfg.AppEnv, serviceHandler, authHandler, authMiddleware, rateLimiter, discoveryHandler, leaderboardHandler)

	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      router,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
		IdleTimeout:  defaultIdleTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("Server starting", "port", cfg.AppPort, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server failed", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	slog.Info("Shutting down gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	slog.Info("Server exited")
	return nil
}



func setupRouter(
	env string,
	serviceHandler *serviceDelivery.ServiceHandler,
	authHandler *authDelivery.AuthHandler,
	authMiddleware gin.HandlerFunc,
	rateLimiter *ratelimit.RateLimiter,
	discoveryHandler *discoveryDelivery.DiscoveryHandler,
	leaderboardHandler *leaderboardDelivery.LeaderboardHandler,
) *gin.Engine {
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Apply rate limiting globally (Phase 3)
	r.Use(ratelimit.RateLimitMiddleware(rateLimiter))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "UP",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	v1 := r.Group("/v1")
	{
		authHandler.RegisterRoutes(v1)

		// Services routes are protected by session middleware (Phase 2 requirement)
		protected := v1.Group("", authMiddleware)
		serviceHandler.RegisterRoutes(protected)

		// Discovery routes (Phase 4):
		// POST /location requires auth to identify the partner from session.
		// GET  /nearby is public (users searching for nearby partners).
		discoveryProtected := v1.Group("", authMiddleware)
		discoveryProtected.POST("/discovery/location", discoveryHandler.UpdateLocation)
		discoveryProtected.POST("/discovery/offline", discoveryHandler.SetOffline)
		v1.GET("/discovery/nearby", discoveryHandler.GetNearby)

		// Leaderboard routes (Phase 5)
		leaderboardHandler.RegisterRoutes(v1, protected)
	}

	return r
}
