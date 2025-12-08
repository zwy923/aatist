package middleware

import (
	"net/http"
	"strconv"

	"github.com/aatist/backend/internal/platform/auth"
	"github.com/aatist/backend/pkg/errs"
	"github.com/aatist/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

const (
	HeaderUserID    = "X-User-ID"
	HeaderUserRole  = "X-User-Role"
	HeaderUserEmail = "X-User-Email"
)

// GatewayAuthMiddleware validates JWT access token and injects user info into headers
// This is used in the API Gateway to forward user identity to downstream services
func GatewayAuthMiddleware(jwt *auth.JWT) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// For public endpoints, continue without authentication
			c.Next()
			return
		}

		// Parse "Bearer <token>"
		const bearerPrefix = "Bearer "
		if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
			c.JSON(http.StatusUnauthorized, response.Error(errs.NewAppError(errs.ErrInvalidToken, http.StatusUnauthorized, "invalid authorization header format").WithCode(errs.CodeInvalidToken)))
			c.Abort()
			return
		}

		token := authHeader[len(bearerPrefix):]
		if token == "" {
			c.JSON(http.StatusUnauthorized, response.Error(errs.NewAppError(errs.ErrInvalidToken, http.StatusUnauthorized, "token required").WithCode(errs.CodeInvalidToken)))
			c.Abort()
			return
		}

		// Validate token
		claims, err := jwt.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, response.Error(errs.NewAppError(errs.ErrInvalidToken, http.StatusUnauthorized, "invalid or expired token").WithCode(errs.CodeInvalidToken)))
			c.Abort()
			return
		}

		// Inject user info into headers for downstream services
		c.Request.Header.Set(HeaderUserID, strconv.FormatInt(claims.UserID, 10))
		c.Request.Header.Set(HeaderUserRole, claims.Role)
		c.Request.Header.Set(HeaderUserEmail, claims.Email)

		// Also set in context for gateway's own use
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("email", claims.Email)
		c.Set("claims", claims)

		c.Next()
	}
}
