package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	notificationRepository "github.com/taqiyyaghazi/ecosystem-engine/internal/features/notifications/repository"
	notificationUsecase "github.com/taqiyyaghazi/ecosystem-engine/internal/features/notifications/usecase"
	taskRepository "github.com/taqiyyaghazi/ecosystem-engine/internal/features/tasks/repository"
	taskUsecase "github.com/taqiyyaghazi/ecosystem-engine/internal/features/tasks/usecase"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/platform/cache"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/platform/config"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/platform/database"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/platform/logger"
)

const (
	defaultConnTimeout = 10 * time.Second
)

func main() {
	if err := run(); err != nil {
		slog.Error("Worker failed to start", "error", err)
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

	// Tasks feature (Phase 8)
	taskRepo := taskRepository.NewTaskRepository(dbPool, rdb)
	taskUseCase := taskUsecase.NewTaskUsecase(taskRepo)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("Background worker starting", "env", cfg.AppEnv)

	// Start subscriber loop in a goroutine
	// Using "notifications:user:*" as pattern or just let the usecase handle it.
	// We will listen on "notifications:user:*" and "notifications:broadcast"
	// PSUBSCRIBE supports patterns like "notifications:*"
	go notificationUseCase.RunSubscriberLoop(ctx, "notifications:*")

	// Start task worker loop
	go taskUseCase.WorkerLoop(ctx, "tasks:default")

	<-ctx.Done()
	slog.Info("Shutting down gracefully, waiting for active processors...")
	notificationUseCase.Shutdown()

	slog.Info("Worker exited")
	return nil
}
