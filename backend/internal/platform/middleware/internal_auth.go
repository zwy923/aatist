package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/aatist/backend/pkg/errs"
	"github.com/aatist/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

const (
	HeaderInternalCall  = "X-Internal-Call"
	HeaderInternalToken = "X-Internal-Token"
)

// RequireInternalCall ensures that the request is from an internal service
// It checks for X-Internal-Call header and optionally validates X-Internal-Token
func RequireInternalCall() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for internal call header
		internalCall := c.GetHeader(HeaderInternalCall)
		if internalCall != "true" {
			c.JSON(http.StatusForbidden, response.Error(errs.NewAppError(errs.ErrForbidden, http.StatusForbidden, "internal API access only").WithCode(errs.CodeForbidden)))
			c.Abort()
			return
		}

		// Optionally validate internal token if configured
		internalToken := os.Getenv("INTERNAL_API_TOKEN")
		if internalToken != "" {
			providedToken := c.GetHeader(HeaderInternalToken)
			if providedToken != internalToken {
				c.JSON(http.StatusForbidden, response.Error(errs.NewAppError(errs.ErrForbidden, http.StatusForbidden, "invalid internal token").WithCode(errs.CodeForbidden)))
				c.Abort()
				return
			}
		}

		// Optional: Check if request is from allowed IPs (gateway internal network)
		// This is a simple check - in production, you might want more sophisticated IP whitelisting
		allowedIPs := os.Getenv("INTERNAL_ALLOWED_IPS")
		if allowedIPs != "" {
			clientIP := c.ClientIP()
			allowedList := strings.Split(allowedIPs, ",")
			allowed := false
			for _, ip := range allowedList {
				if strings.TrimSpace(ip) == clientIP {
					allowed = true
					break
				}
			}
			if !allowed {
				c.JSON(http.StatusForbidden, response.Error(errs.NewAppError(errs.ErrForbidden, http.StatusForbidden, "request not from allowed IP").WithCode(errs.CodeForbidden)))
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
