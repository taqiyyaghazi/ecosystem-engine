package ratelimit

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	defaultLimit  = 10
	defaultWindow = 60 * time.Second
)

// RateLimitMiddleware returns a Gin HandlerFunc that enforces rate limiting
// using the Fixed Window algorithm. It adds standard X-RateLimit-* headers
// to every response and aborts with 429 when the limit is exceeded.
func RateLimitMiddleware(rl *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Construct a unique key per endpoint + client IP.
		key := fmt.Sprintf("ratelimit:%s:%s", c.FullPath(), c.ClientIP())

		allowed, count, err := rl.IsAllowed(c.Request.Context(), key, defaultLimit, defaultWindow)

		// Calculate remaining requests (floor at 0 to avoid negative values).
		remaining := int64(defaultLimit) - count
		if remaining < 0 {
			remaining = 0
		}

		// Always attach rate limit headers so clients can track their quota.
		c.Header("X-RateLimit-Limit", strconv.Itoa(defaultLimit))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		c.Header("X-RateLimit-Reset", strconv.Itoa(int(defaultWindow.Seconds())))

		if err != nil {
			// On Redis failure, fail open (allow the request) to avoid
			// blocking legitimate traffic due to infrastructure issues.
			c.Next()
			return
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":                "Too many requests. Please try again in a minute.",
				"retry_after_seconds":  int(defaultWindow.Seconds()),
			})
			return
		}

		c.Next()
	}
}
