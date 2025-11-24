package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const RequestIDKey = "request_id"

// RequestIDMiddleware generates and injects a request ID
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID already exists (from upstream)
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// Generate new request ID
			requestID = uuid.New().String()
		}

		// Set in header
		c.Header("X-Request-ID", requestID)

		// Set in context for logging
		c.Set(RequestIDKey, requestID)

		c.Next()
	}
}

// GetRequestID extracts request ID from context
func GetRequestID(c *gin.Context) string {
	if id, exists := c.Get(RequestIDKey); exists {
		if str, ok := id.(string); ok {
			return str
		}
	}
	return ""
}

// RequestIDLogField returns a zap field for request ID
func RequestIDLogField(c *gin.Context) zap.Field {
	return zap.String("request_id", GetRequestID(c))
}

