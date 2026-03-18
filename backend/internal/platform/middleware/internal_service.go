package middleware

import (
	"net/http"
	"os"

	"github.com/aatist/backend/pkg/errs"
	"github.com/aatist/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// ValidateInternalCaller ensures requests to Gateway's internal routes are from authorized services.
// When INTERNAL_API_TOKEN is set, caller must send matching X-Internal-Token.
func ValidateInternalCaller() gin.HandlerFunc {
	return func(c *gin.Context) {
		internalToken := os.Getenv("INTERNAL_API_TOKEN")
		if internalToken == "" {
			c.Next()
			return
		}
		provided := c.GetHeader(HeaderInternalToken)
		if provided != internalToken {
			c.JSON(http.StatusForbidden, response.Error(errs.NewAppError(errs.ErrForbidden, http.StatusForbidden, "invalid internal token").WithCode(errs.CodeForbidden)))
			c.Abort()
			return
		}
		c.Next()
	}
}

// InternalServiceMiddleware injects internal call headers for downstream services.
func InternalServiceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Header.Set(HeaderInternalCall, "true")
		if token := os.Getenv("INTERNAL_API_TOKEN"); token != "" {
			c.Request.Header.Set(HeaderInternalToken, token)
		}
		c.Next()
	}
}

