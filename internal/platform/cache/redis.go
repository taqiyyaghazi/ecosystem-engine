package cache

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(ctx context.Context, url, password string, dbStr string) (*redis.Client, error) {
	db, _ := strconv.Atoi(dbStr)

	rdb := redis.NewClient(&redis.Options{
		Addr:     url,
		Password: password,
		DB:       db,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("unable to connect to redis: %w", err)
	}

	_, _ = fmt.Fprintln(os.Stdout, "Successfully connected to Redis")
	return rdb, nil
}
