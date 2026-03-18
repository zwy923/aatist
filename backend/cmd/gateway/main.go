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

	"github.com/aatist/backend/internal/gateway/router"
	"github.com/aatist/backend/internal/platform/auth"
	"github.com/aatist/backend/internal/platform/cache"
	"github.com/aatist/backend/internal/platform/config"
	"github.com/aatist/backend/internal/platform/log"
	"github.com/aatist/backend/internal/platform/middleware"
	"github.com/aatist/backend/pkg/response"
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
	if (cfg.App.Env == "production" || cfg.App.Env == "prod") && os.Getenv("INTERNAL_API_TOKEN") == "" {
		logger.Warn("INTERNAL_API_TOKEN is not set in production; internal routes are less secure")
	}

	// Initialize JWT
	jwt := auth.NewJWT(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)

	// Initialize Redis
	redisClient, err := cache.NewRedis(cfg.Redis.Addr, cfg.Redis.DB)
	if err != nil {
		logger.Fatal("Failed to initialize Redis", zap.Error(err))
	}
	defer redisClient.Close()

	// Setup Gin router
	if cfg.App.Env == "production" || cfg.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Apply global middlewares
	r.Use(middleware.RecoveryMiddleware(logger))
	r.Use(middleware.RequestIDMiddleware())

	// CORS configuration
	env := os.Getenv("CORS_ORIGINS")
	var corsOrigins []string
	if env == "" {
		corsOrigins = []string{"*"}
	} else {
		corsOrigins = strings.Split(env, ",")
	}
	r.Use(middleware.CORSMiddleware(corsOrigins))

	// Health check endpoint
	r.GET("/gateway/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, response.Success(gin.H{"status": "ok", "service": "gateway"}))
	})

	// Register all routes
	router.RegisterRoutes(r, cfg, logger, jwt, redisClient)

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.App.HTTPPort),
		Handler: r,
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
