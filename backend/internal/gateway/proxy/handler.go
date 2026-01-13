package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/aatist/backend/internal/platform/log"
	"github.com/aatist/backend/internal/platform/middleware"
	"github.com/aatist/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Hop-by-hop headers that should not be forwarded (RFC 2616)
var hopByHopHeaders = map[string]bool{
	"Connection":          true,
	"Transfer-Encoding":   true,
	"Keep-Alive":          true,
	"Proxy-Authenticate":  true,
	"Proxy-Authorization": true,
	"Trailer":             true,
	"TE":                  true,
	"Upgrade":             true,
}

// Context key type for request ID
type contextKey string

const requestIDKey contextKey = "request_id"

// NewHandler creates a reverse proxy handler with custom timeout
func NewHandler(serviceName string, port int, timeout time.Duration, logger *log.Logger) gin.HandlerFunc {
	// Build target URL
	targetURL, err := url.Parse(fmt.Sprintf("http://%s:%d", serviceName, port))
	if err != nil {
		logger.Fatal("Failed to parse target URL", zap.Error(err))
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Customize Director to modify the request
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		// Call original director first (sets Host, Scheme, etc.)
		originalDirector(req)

		// Remove hop-by-hop headers from request
		for key := range req.Header {
			if hopByHopHeaders[key] {
				delete(req.Header, key)
			}
		}

		// Forward request ID from context if present
		if ctxRequestID, ok := req.Context().Value(requestIDKey).(string); ok {
			req.Header.Set("X-Request-ID", ctxRequestID)
		}

		// Forward user identity headers if present (these are already in headers from GatewayAuthMiddleware)
		// No need to modify them, just ensure they're forwarded

		// Log the proxied request
		logger.Info("Proxying request",
			zap.String("service", serviceName),
			zap.String("path", req.URL.Path),
			zap.String("method", req.Method),
			zap.String("target", req.URL.String()),
		)
	}

	// Customize error handler
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		logger.Error("Reverse proxy error",
			zap.String("service", serviceName),
			zap.String("path", req.URL.Path),
			zap.Error(err),
		)
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(rw).Encode(response.Error(err))
	}

	// Customize ModifyResponse to filter hop-by-hop headers and CORS headers from response
	proxy.ModifyResponse = func(resp *http.Response) error {
		// Remove hop-by-hop headers from response
		for key := range resp.Header {
			if hopByHopHeaders[key] {
				resp.Header.Del(key)
			}
		}

		// Remove CORS headers from backend service response
		// Gateway handles CORS, so we don't want duplicate CORS headers
		corsHeaders := []string{
			"Access-Control-Allow-Origin",
			"Access-Control-Allow-Credentials",
			"Access-Control-Allow-Headers",
			"Access-Control-Allow-Methods",
			"Access-Control-Max-Age",
			"Access-Control-Expose-Headers",
		}
		for _, header := range corsHeaders {
			resp.Header.Del(header)
		}

		return nil
	}

	return func(c *gin.Context) {
		// Store request ID in context for Director to access
		if requestID := middleware.GetRequestID(c); requestID != "" {
			ctx := context.WithValue(c.Request.Context(), requestIDKey, requestID)
			c.Request = c.Request.WithContext(ctx)
		}

		// Apply timeout if specified
		if timeout > 0 {
			ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
			defer cancel()
			c.Request = c.Request.WithContext(ctx)
		}

		// Use ReverseProxy to handle the request
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
