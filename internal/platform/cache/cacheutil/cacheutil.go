package cacheutil

import "context"

type contextKey string

const cacheStatusKey contextKey = "x-cache-status"

const (
	StatusHIT  = "HIT"
	StatusMISS = "MISS"
)

func WithCacheStatus(ctx context.Context, status string) context.Context {
	return context.WithValue(ctx, cacheStatusKey, status)
}

func CacheStatusFromContext(ctx context.Context) string {
	val, _ := ctx.Value(cacheStatusKey).(string)
	return val
}
