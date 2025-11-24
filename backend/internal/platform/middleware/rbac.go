package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/aalto-talent-network/backend/internal/user/model"
)

// RequireRole checks if the user has one of the required roles
func RequireRole(roles ...model.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleStr, err := GetRole(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		userRole := model.Role(roleStr)
		if !userRole.IsValid() {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid role"})
			c.Abort()
			return
		}

		// Check if user role is in allowed roles
		allowed := false
		for _, allowedRole := range roles {
			if userRole == allowedRole {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

