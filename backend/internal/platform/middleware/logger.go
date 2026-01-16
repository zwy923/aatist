package middleware

import (
	"time"

	"github.com/aatist/backend/internal/platform/log"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoggerMiddleware logs incoming requests
func LoggerMiddleware(logger *log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		if len(c.Errors) > 0 {
			for _, e := range c.Errors.Errors() {
				logger.Error("Request error", zap.String("error", e))
			}
		} else {
			logger.Info("Request",
				zap.Int("status", c.Writer.Status()),
				zap.String("method", c.Request.Method),
				zap.String("path", path),
				zap.String("query", query),
				zap.String("ip", c.ClientIP()),
				zap.Duration("latency", latency),
				zap.String("user_id", c.GetHeader(HeaderUserID)),
			)
		}
	}
}
