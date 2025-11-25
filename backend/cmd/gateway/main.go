package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aalto-talent-network/backend/internal/platform/auth"
	"github.com/aalto-talent-network/backend/internal/platform/cache"
	"github.com/aalto-talent-network/backend/internal/platform/config"
	"github.com/aalto-talent-network/backend/internal/platform/log"
	"github.com/aalto-talent-network/backend/internal/platform/middleware"
	"github.com/aalto-talent-network/backend/pkg/response"
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

func main() {
	// Load configuration
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "configs/config.yaml"
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger, err := log.NewLogger(cfg.App.Env)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting API Gateway",
		zap.String("env", cfg.App.Env),
		zap.Int("port", cfg.App.HTTPPort),
	)

	// Initialize Redis (for rate limiting, etc.)
	// Redis is optional for Gateway - it's mainly used for rate limiting
	// If Redis is not available, Gateway can still function (rate limiting will be disabled)
	redis, err := cache.NewRedis(cfg.Redis.Addr, cfg.Redis.DB)
	if err != nil {
		logger.Warn("Failed to initialize Redis - Gateway will continue without rate limiting",
			zap.String("redis_addr", cfg.Redis.Addr),
			zap.Error(err),
		)
		redis = nil // Set to nil so we can check later
	} else {
		defer redis.Close()
		logger.Info("Connected to Redis", zap.String("addr", cfg.Redis.Addr))
	}

	// Initialize JWT
	jwt := auth.NewJWT(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)

	// Helper function to get service timeout
	getServiceTimeout := func(serviceName string) time.Duration {
		if timeout, ok := cfg.Gateway.ServiceTimeouts[serviceName]; ok {
			return timeout
		}
		return cfg.Gateway.ServiceTimeout
	}

	// Setup Gin router
	if cfg.App.Env == "production" || cfg.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Apply global middlewares
	router.Use(middleware.RecoveryMiddleware(logger))
	router.Use(middleware.RequestIDMiddleware())

	// CORS configuration
	env := os.Getenv("CORS_ORIGINS")
	var corsOrigins []string
	if env == "" {
		// Default to allow all in development
		corsOrigins = []string{"*"}
	} else {
		corsOrigins = strings.Split(env, ",")
	}
	router.Use(middleware.CORSMiddleware(corsOrigins))

	// Health check endpoint
	router.GET("/gateway/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, response.Success(gin.H{"status": "ok", "service": "gateway"}))
	})

	// API routes with /api/v1 prefix
	api := router.Group("/api/v1")
	{
		// Public routes (no auth required)
		public := api.Group("")
		{
			// Proxy to user-service for auth endpoints
			public.Any("/auth/*path", proxyToServiceWithTimeout("user-service", 8081, getServiceTimeout("user-service"), logger))

			// Public user routes → user-service
			// Check username/email availability (for registration validation)
			public.GET("/users/check-username", proxyToServiceWithTimeout("user-service", 8081, getServiceTimeout("user-service"), logger))
			public.GET("/users/check-email", proxyToServiceWithTimeout("user-service", 8081, getServiceTimeout("user-service"), logger))
			// GET /users/:id - view user profile (public)
			public.GET("/users/:id", proxyToServiceWithTimeout("user-service", 8081, getServiceTimeout("user-service"), logger))
			// GET /users/:id/summary - view user summary (public, lightweight profile for cards)
			public.GET("/users/:id/summary", proxyToServiceWithTimeout("user-service", 8081, getServiceTimeout("user-service"), logger))

			// Public portfolio routes → portfolio-service
			// GET /portfolio/:id - view single project (public)
			public.GET("/portfolio/:id", proxyToServiceWithTimeout("portfolio-service", 8082, getServiceTimeout("portfolio-service"), logger))
			// GET /users/:id/portfolio - view user's portfolio (public)
			public.GET("/users/:id/portfolio", proxyToServiceWithTimeout("portfolio-service", 8082, getServiceTimeout("portfolio-service"), logger))
		}

		// Public community routes (no auth required) → community-service
		// These must be before protected routes to avoid auth requirement
		public.GET("/community/posts", proxyToServiceWithTimeout("community-service", 8087, getServiceTimeout("community-service"), logger))
		public.GET("/community/posts/trending", proxyToServiceWithTimeout("community-service", 8087, getServiceTimeout("community-service"), logger))
		public.GET("/community/posts/:id", proxyToServiceWithTimeout("community-service", 8087, getServiceTimeout("community-service"), logger))
		public.GET("/community/posts/:id/comments", proxyToServiceWithTimeout("community-service", 8087, getServiceTimeout("community-service"), logger))
		// Public user posts
		public.GET("/community/users/:id/posts", proxyToServiceWithTimeout("community-service", 8087, getServiceTimeout("community-service"), logger))

		// Public events routes (no auth required) → events-service
		public.GET("/events", proxyToServiceWithTimeout("events-service", 8084, getServiceTimeout("events-service"), logger))
		public.GET("/events/:id", proxyToServiceWithTimeout("events-service", 8084, getServiceTimeout("events-service"), logger))
		public.GET("/events/:id/comments", proxyToServiceWithTimeout("events-service", 8084, getServiceTimeout("events-service"), logger))

		// Protected routes (require auth)
		protected := api.Group("")
		protected.Use(middleware.GatewayAuthMiddleware(jwt))
		{
			// Notification routes → notification-service
			protected.Any("/notifications/*path", proxyToServiceWithTimeout("notification-service", 8085, getServiceTimeout("notification-service"), logger))
			protected.Any("/notifications", proxyToServiceWithTimeout("notification-service", 8085, getServiceTimeout("notification-service"), logger))

			// Current user routes → route based on path
			// /users/me/portfolio* → portfolio-service
			// /users/me/* (other) → user-service
			protected.Any("/users/me/*path", func(c *gin.Context) {
				path := c.Param("path")
				if strings.HasPrefix(path, "/portfolio") {
					// Portfolio routes → portfolio-service
					proxyToServiceWithTimeout("portfolio-service", 8082, getServiceTimeout("portfolio-service"), logger)(c)
				} else {
					// All other /users/me/* routes → user-service
					proxyToServiceWithTimeout("user-service", 8081, getServiceTimeout("user-service"), logger)(c)
				}
			})
			protected.Any("/users/me", proxyToServiceWithTimeout("user-service", 8081, getServiceTimeout("user-service"), logger))

			// File routes → file-service
			protected.Any("/files/*path", proxyToServiceWithTimeout("file-service", 8086, getServiceTimeout("file-service"), logger))
			protected.Any("/files", proxyToServiceWithTimeout("file-service", 8086, getServiceTimeout("file-service"), logger))

			// Opportunities routes → opp-service
			protected.Any("/opportunities/*path", proxyToServiceWithTimeout("opp-service", 8083, getServiceTimeout("opp-service"), logger))
			protected.Any("/opportunities", proxyToServiceWithTimeout("opp-service", 8083, getServiceTimeout("opp-service"), logger))

			// Events routes → events-service
			protected.Any("/events/*path", proxyToServiceWithTimeout("events-service", 8084, getServiceTimeout("events-service"), logger))
			protected.Any("/events", proxyToServiceWithTimeout("events-service", 8084, getServiceTimeout("events-service"), logger))

			// Community protected routes → community-service
			protected.POST("/community/posts", proxyToServiceWithTimeout("community-service", 8087, getServiceTimeout("community-service"), logger))
			protected.PUT("/community/posts/:id", proxyToServiceWithTimeout("community-service", 8087, getServiceTimeout("community-service"), logger))
			protected.DELETE("/community/posts/:id", proxyToServiceWithTimeout("community-service", 8087, getServiceTimeout("community-service"), logger))
			protected.POST("/community/posts/:id/like", proxyToServiceWithTimeout("community-service", 8087, getServiceTimeout("community-service"), logger))
			protected.DELETE("/community/posts/:id/like", proxyToServiceWithTimeout("community-service", 8087, getServiceTimeout("community-service"), logger))
			protected.POST("/community/posts/:id/comments", proxyToServiceWithTimeout("community-service", 8087, getServiceTimeout("community-service"), logger))
			protected.PUT("/community/comments/:id", proxyToServiceWithTimeout("community-service", 8087, getServiceTimeout("community-service"), logger))
			protected.DELETE("/community/comments/:id", proxyToServiceWithTimeout("community-service", 8087, getServiceTimeout("community-service"), logger))
			// Current user's posts
			protected.GET("/community/users/me/posts", proxyToServiceWithTimeout("community-service", 8087, getServiceTimeout("community-service"), logger))
		}

		// Internal API routes (for service-to-service communication)
		// These routes bypass Gateway auth but require internal authentication
		// All internal traffic goes through Gateway for unified monitoring, rate limiting, and tracing
		internal := api.Group("/internal")
		internal.Use(middleware.InternalServiceMiddleware())
		{
			// Internal user API (for other services to check user profile visibility, etc.)
			internal.Any("/user/*path", func(c *gin.Context) {
				// Rewrite path: /api/v1/internal/user/users/:id -> /api/v1/users/:id
				originalPath := c.Request.URL.Path
				if strings.HasPrefix(originalPath, "/api/v1/internal/user/") {
					newPath := strings.Replace(originalPath, "/api/v1/internal/user/", "/api/v1/", 1)
					c.Request.URL.Path = newPath
				}
				proxyToServiceWithTimeout("user-service", 8081, getServiceTimeout("user-service"), logger)(c)
			})

			// Internal notification API (for other services to create notifications)
			internal.Any("/notification/*path", func(c *gin.Context) {
				// Rewrite path: /api/v1/internal/notification/notifications -> /api/v1/internal/notifications
				originalPath := c.Request.URL.Path
				if strings.HasPrefix(originalPath, "/api/v1/internal/notification/") {
					newPath := strings.Replace(originalPath, "/api/v1/internal/notification/", "/api/v1/internal/", 1)
					c.Request.URL.Path = newPath
				}
				proxyToServiceWithTimeout("notification-service", 8085, getServiceTimeout("notification-service"), logger)(c)
			})

			// Internal portfolio API (for future service-to-service calls)
			internal.Any("/portfolio/*path", proxyToServiceWithTimeout("portfolio-service", 8082, getServiceTimeout("portfolio-service"), logger))

			// Internal file API (for other services to upload files)
			// Keep the path as-is since file-service expects /api/v1/internal/file/*
			internal.Any("/file/*path", proxyToServiceWithTimeout("file-service", 8086, getServiceTimeout("file-service"), logger))
		}
	}

	// Also support /auth/* for convenience (forwards to /api/v1/auth/*)
	router.Any("/auth/*path", func(c *gin.Context) {
		path := c.Param("path")
		// Rewrite the path to include /api/v1 prefix
		c.Request.URL.Path = "/api/v1/auth" + path
		// Create a new handler and call it
		proxyToServiceWithTimeout("user-service", 8081, getServiceTimeout("user-service"), logger)(c)
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.App.HTTPPort),
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		logger.Info("HTTP server starting", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

// proxyToService creates a reverse proxy handler to forward requests to downstream services
// Uses httputil.ReverseProxy for better performance, connection reuse, and proper streaming
func proxyToService(serviceName string, port int, logger *log.Logger) gin.HandlerFunc {
	return proxyToServiceWithTimeout(serviceName, port, 0, logger)
}

// proxyToServiceWithTimeout creates a reverse proxy handler with custom timeout
func proxyToServiceWithTimeout(serviceName string, port int, timeout time.Duration, logger *log.Logger) gin.HandlerFunc {
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

	// Customize ModifyResponse to filter hop-by-hop headers from response
	proxy.ModifyResponse = func(resp *http.Response) error {
		// Remove hop-by-hop headers from response
		for key := range resp.Header {
			if hopByHopHeaders[key] {
				resp.Header.Del(key)
			}
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
