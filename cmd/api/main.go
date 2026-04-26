package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/services/delivery"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/services/repository"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/services/usecase"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/platform/cache"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/platform/config"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/platform/database"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Configuration error", "error", err)
		os.Exit(1)
	}

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbPool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		dbPool.Close()
		slog.Info("Database connection closed")
	}()

	rdb, err := cache.NewRedisClient(ctx, cfg.RedisURL, cfg.RedisPass, cfg.RedisDB)
	if err != nil {
		slog.Error("Failed to connect to redis", "error", err)
		os.Exit(1)
	}
	defer func() {
		_ = rdb.Close()
		slog.Info("Redis connection closed")
	}()

	serviceRepo := repository.NewServiceRepository(dbPool, rdb)
	serviceUsecase := usecase.NewServiceUsecase(serviceRepo)
	serviceHandler := delivery.NewServiceHandler(serviceUsecase)
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "UP",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	v1 := r.Group("/v1")
	serviceHandler.RegisterRoutes(v1)

	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("Server starting", "port", cfg.AppPort, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server exited gracefully")
}
