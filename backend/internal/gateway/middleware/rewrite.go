package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// InternalRewrite returns a middleware that rewrites paths for internal requests
// prefixToRemove: the prefix to strip from the path (e.g., "/api/v1/internal/user")
// prefixToAdd: the prefix to add to the path (e.g., "/api/v1")
func InternalRewrite(prefixToRemove, prefixToAdd string) gin.HandlerFunc {
	return func(c *gin.Context) {
		originalPath := c.Request.URL.Path
		if strings.HasPrefix(originalPath, prefixToRemove) {
			newPath := strings.TrimPrefix(originalPath, prefixToRemove)
			c.Request.URL.Path = prefixToAdd + newPath
			// Add header to identify internal request
			c.Request.Header.Set("X-Internal-Request", "true")
		}
		c.Next()
	}
}
