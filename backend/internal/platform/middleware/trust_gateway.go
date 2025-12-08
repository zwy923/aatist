package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/aatist/backend/pkg/errs"
	"github.com/aatist/backend/pkg/response"
)

// TrustGatewayMiddleware extracts user identity from headers set by Gateway
// This middleware should be used in all microservices to trust Gateway-injected headers
func TrustGatewayMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract user info from headers (set by Gateway)
		userIDStr := c.GetHeader(HeaderUserID)
		userRole := c.GetHeader(HeaderUserRole)
		userEmail := c.GetHeader(HeaderUserEmail)

		// If headers are present, trust them (Gateway has already validated JWT)
		if userIDStr != "" && userRole != "" && userEmail != "" {
			userID, err := strconv.ParseInt(userIDStr, 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, response.Error(errs.NewAppError(errs.ErrInvalidInput, http.StatusBadRequest, "invalid user ID header").WithCode(errs.CodeInvalidInput)))
				c.Abort()
				return
			}

			// Inject into context
			c.Set("user_id", userID)
			c.Set("role", userRole)
			c.Set("email", userEmail)
		}

		c.Next()
	}
}

// RequireGatewayAuth ensures that user identity headers are present
// Use this for protected endpoints in microservices
func RequireGatewayAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.GetHeader(HeaderUserID)
		if userIDStr == "" {
			c.JSON(http.StatusUnauthorized, response.Error(errs.NewAppError(errs.ErrUnauthorized, http.StatusUnauthorized, "authentication required").WithCode(errs.CodeUnauthorized)))
			c.Abort()
			return
		}

		c.Next()
	}
}

