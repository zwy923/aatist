package middleware

import (
	"net/http"
	"strconv"
	"strings"

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

func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	const bearerPrefix = "Bearer "
	if len(authHeader) >= len(bearerPrefix) && authHeader[:len(bearerPrefix)] == bearerPrefix {
		t := strings.TrimSpace(authHeader[len(bearerPrefix):])
		if t != "" {
			return t
		}
	}
	// WebSocket 握手无法带自定义头，允许从 query 取 token
	if c.Request.Method == http.MethodGet && (c.Request.URL.Path == "/api/v1/ws" || strings.HasSuffix(c.Request.URL.Path, "/ws")) {
		return c.Query("token")
	}
	return ""
}

// GatewayAuthMiddleware validates JWT access token and injects user info into headers
// This is used in the API Gateway to forward user identity to downstream services
func GatewayAuthMiddleware(jwt *auth.JWT) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
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

// InjectUserFromJWTIfNoGatewayHeaders validates the Bearer access token when X-User-ID is not set
// and injects the same X-User-* headers the gateway would set.
//
// Use this on chat-service (and similar) when HTTP traffic can reach the service without passing
// through the API gateway — e.g. nginx routes /api/v1/conversations directly to chat-service.
// If the gateway already set X-User-ID, this middleware is a no-op.
func InjectUserFromJWTIfNoGatewayHeaders(j *auth.JWT) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.TrimSpace(c.GetHeader(HeaderUserID)) != "" {
			c.Next()
			return
		}
		token := extractToken(c)
		if token == "" {
			c.Next()
			return
		}
		claims, err := j.ValidateToken(token)
		if err != nil {
			c.Next()
			return
		}
		c.Request.Header.Set(HeaderUserID, strconv.FormatInt(claims.UserID, 10))
		c.Request.Header.Set(HeaderUserRole, claims.Role)
		c.Request.Header.Set(HeaderUserEmail, claims.Email)
		// Match GatewayAuthMiddleware + TrustGateway: handlers use middleware.GetUserID (context), not headers.
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("email", claims.Email)
		c.Set("claims", claims)
		c.Next()
	}
}
