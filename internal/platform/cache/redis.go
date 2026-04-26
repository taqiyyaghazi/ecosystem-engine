package cache

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(ctx context.Context, url, password string, dbStr string) (*redis.Client, error) {
	db, err := strconv.Atoi(dbStr)
	if err != nil {
		slog.Warn("Invalid REDIS_DB value, defaulting to 0", "value", dbStr, "error", err)
		db = 0
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     url,
		Password: password,
		DB:       db,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("unable to connect to redis: %w", err)
	}

	slog.Info("Successfully connected to Redis")
	return rdb, nil
}
