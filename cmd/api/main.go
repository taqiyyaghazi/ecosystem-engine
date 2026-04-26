package main

import (
	"context"
	"log"
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
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbPool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	rdb, err := cache.NewRedisClient(ctx, cfg.RedisURL, cfg.RedisPass, cfg.RedisDB)
	if err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}
	defer func() { _ = rdb.Close() }()

	serviceRepo := repository.NewServiceRepository(dbPool, rdb)
	serviceUsecase := usecase.NewServiceUsecase(serviceRepo)
	serviceHandler := delivery.NewServiceHandler(serviceUsecase)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	v1 := r.Group("/v1")
	serviceHandler.RegisterRoutes(v1)

	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: r,
	}

	go func() {
		slog.Info("Server starting", "port", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server...")

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	slog.Info("Server exited")
}
