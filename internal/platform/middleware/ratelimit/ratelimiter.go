package ratelimit

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter holds the Redis client used for rate limit tracking.
type RateLimiter struct {
	redis *redis.Client
}

// NewRateLimiter creates a new RateLimiter backed by the given Redis client.
func NewRateLimiter(rdb *redis.Client) *RateLimiter {
	return &RateLimiter{redis: rdb}
}

// IsAllowed applies the Fixed Window algorithm.
// It atomically increments the counter for the given key and sets an
// expiration only on the very first request of a new window.
// Returns (allowed bool, currentCount int64, error).
func (rl *RateLimiter) IsAllowed(ctx context.Context, key string, limit int, window time.Duration) (bool, int64, error) {
	// 1. Atomically increment the counter.
	count, err := rl.redis.Incr(ctx, key).Result()
	if err != nil {
		return false, 0, err
	}

	// 2. Set expiration only on the first request of a new window.
	//    This is safe: if INCR returned 1 this is definitively the first request.
	if count == 1 {
		rl.redis.Expire(ctx, key, window)
	}

	// 3. Check against the threshold.
	return count <= int64(limit), count, nil
}
