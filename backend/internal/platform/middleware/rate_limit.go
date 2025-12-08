package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/aatist/backend/internal/platform/cache"
	"github.com/aatist/backend/pkg/errs"
)

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(redis *cache.Redis, keyPrefix string, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate key based on IP or user ID
		ip := c.ClientIP()
		key := fmt.Sprintf("%s:%s", keyPrefix, ip)

		client := redis.GetClient()
		ctx := c.Request.Context()

		count, err := client.Incr(ctx, key).Result()
		if err != nil {
			// Fail open - allow request if Redis is down
			c.Next()
			return
		}

		if count == 1 {
			// Set expiration on first request
			client.Expire(ctx, key, window)
		}

		if count > int64(limit) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": errs.ErrRateLimitExceeded.Error(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

