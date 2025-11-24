package main

import (
	"context"
	"fmt"
	"net/http"
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

	// Setup Gin router
	if cfg.App.Env == "production" || cfg.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Apply global middlewares
	router.Use(middleware.RecoveryMiddleware(logger))
	router.Use(middleware.RequestIDMiddleware())

	// CORS configuration
	corsOrigins := strings.Split(os.Getenv("CORS_ORIGINS"), ",")
	if len(corsOrigins) == 0 || corsOrigins[0] == "" {
		// Default to allow all in development
		corsOrigins = []string{"*"}
	}
	router.Use(middleware.CORSMiddleware(corsOrigins))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, response.Success(gin.H{"status": "ok"}))
	})

	// API routes with /api/v1 prefix
	api := router.Group("/api/v1")
	{
		// Public routes (no auth required)
		public := api.Group("")
		{
			// Proxy to user-service for auth endpoints
			public.Any("/auth/*path", proxyToService("user-service", 8081, logger))
			// Public portfolio route (for viewing project details)
			public.Any("/portfolio/*path", proxyToService("user-service", 8081, logger))
			// Public user profile routes
			public.Any("/users/:id", proxyToService("user-service", 8081, logger))
			public.Any("/users/:id/portfolio", proxyToService("user-service", 8081, logger))
		}

		// Protected routes (require auth)
		protected := api.Group("")
		protected.Use(middleware.GatewayAuthMiddleware(jwt))
		{
			// Proxy to various services
			protected.Any("/users/*path", proxyToService("user-service", 8081, logger))
			// Portfolio routes are handled by user-service
			protected.Any("/portfolio/*path", proxyToService("user-service", 8081, logger))
			protected.Any("/opportunities/*path", proxyToService("opp-service", 8083, logger))
			protected.Any("/community/*path", proxyToService("community-service", 8084, logger))
		}
	}

	// Also support /auth/* for convenience (forwards to /api/v1/auth/*)
	router.Any("/auth/*path", func(c *gin.Context) {
		path := c.Param("path")
		// Rewrite the path to include /api/v1 prefix
		c.Request.URL.Path = "/api/v1/auth" + path
		// Create a new handler and call it
		proxyToService("user-service", 8081, logger)(c)
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

// proxyToService creates a proxy handler to forward requests to downstream services
func proxyToService(serviceName string, port int, logger *log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Simple proxy implementation
		// In production, use a proper reverse proxy library like httputil.ReverseProxy

		// Get the full request path
		// This will be the complete path like /api/v1/auth/register
		fullPath := c.Request.URL.Path

		// Build target URL - forward the full path to maintain consistency
		targetURL := fmt.Sprintf("http://%s:%d%s", serviceName, port, fullPath)
		if c.Request.URL.RawQuery != "" {
			targetURL += "?" + c.Request.URL.RawQuery
		}

		logger.Info("Proxying request",
			zap.String("service", serviceName),
			zap.String("path", fullPath),
			zap.String("method", c.Request.Method),
			middleware.RequestIDLogField(c),
		)

		// Create request to downstream service
		req, err := http.NewRequest(c.Request.Method, targetURL, c.Request.Body)
		if err != nil {
			logger.Error("Failed to create proxy request", zap.Error(err))
			c.JSON(http.StatusInternalServerError, response.Error(err))
			return
		}

		// Copy headers (including user identity headers set by GatewayAuthMiddleware)
		for key, values := range c.Request.Header {
			req.Header[key] = values
		}

		// Forward request ID
		if requestID := middleware.GetRequestID(c); requestID != "" {
			req.Header.Set("X-Request-ID", requestID)
		}

		// Make request
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			logger.Error("Failed to proxy request", zap.Error(err))
			c.JSON(http.StatusBadGateway, response.Error(err))
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}

		// Copy status code and body
		c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
	}
}
