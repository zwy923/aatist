package middleware

import (
	"os"

	"github.com/gin-gonic/gin"
)

// InternalServiceMiddleware automatically sets internal call headers for Gateway's internal routes
// This middleware should be used in Gateway's internal API group to automatically inject
// X-Internal-Call and X-Internal-Token headers for downstream services
func InternalServiceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set internal call header
		c.Request.Header.Set(HeaderInternalCall, "true")

		// Set internal token if configured
		internalToken := os.Getenv("INTERNAL_API_TOKEN")
		if internalToken != "" {
			c.Request.Header.Set(HeaderInternalToken, internalToken)
		}

		c.Next()
	}
}

