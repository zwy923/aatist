package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/aalto-talent-network/backend/internal/platform/auth"
	"github.com/aalto-talent-network/backend/pkg/errs"
)

// AuthMiddleware validates JWT access token and injects user info into context
func AuthMiddleware(jwt *auth.JWT) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		// Parse "Bearer <token>"
		const bearerPrefix = "Bearer "
		if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		token := authHeader[len(bearerPrefix):]
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
			c.Abort()
			return
		}

		// Validate token
		claims, err := jwt.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// Inject user info into context
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("email", claims.Email)
		c.Set("claims", claims)

		c.Next()
	}
}

// GetUserID extracts user ID from context (must be used after AuthMiddleware)
func GetUserID(c *gin.Context) (int64, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, errs.ErrUnauthorized
	}

	id, ok := userID.(int64)
	if !ok {
		return 0, errs.ErrUnauthorized
	}

	return id, nil
}

// GetRole extracts role from context (must be used after AuthMiddleware)
func GetRole(c *gin.Context) (string, error) {
	role, exists := c.Get("role")
	if !exists {
		return "", errs.ErrUnauthorized
	}

	r, ok := role.(string)
	if !ok {
		return "", errs.ErrUnauthorized
	}

	return r, nil
}

