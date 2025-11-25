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

	"github.com/aalto-talent-network/backend/internal/event/handler"
	"github.com/aalto-talent-network/backend/internal/event/repository"
	"github.com/aalto-talent-network/backend/internal/event/service"
	"github.com/aalto-talent-network/backend/internal/platform/config"
	"github.com/aalto-talent-network/backend/internal/platform/db"
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

	logger.Info("Starting events service",
		zap.String("env", cfg.App.Env),
		zap.Int("port", cfg.App.HTTPPort),
	)

	// Initialize PostgreSQL
	postgres, err := db.NewPostgres(cfg.Postgres.DSN)
	if err != nil {
		logger.Fatal("Failed to initialize PostgreSQL", zap.Error(err))
	}
	defer postgres.Close()

	logger.Info("Connected to PostgreSQL")

	// Run database migrations
	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}
	if err := db.RunMigrations(postgres.GetSQLDB(), migrationsDir); err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}
	logger.Info("Database migrations completed")

	// Initialize repositories
	eventRepo := repository.NewPostgresEventRepository(postgres.GetDB())
	interestRepo := repository.NewPostgresEventInterestRepository(postgres.GetDB())
	goingRepo := repository.NewPostgresEventGoingRepository(postgres.GetDB())
	commentRepo := repository.NewPostgresEventCommentRepository(postgres.GetDB())

	// Initialize services
	eventService := service.NewEventService(eventRepo, interestRepo, goingRepo, commentRepo)
	interestService := service.NewEventInterestService(interestRepo)
	goingService := service.NewEventGoingService(goingRepo)
	commentService := service.NewEventCommentService(commentRepo)

	// Initialize handler
	eventHandler := handler.NewEventHandler(eventService, interestService, goingService, commentService, logger)

	// Setup Gin router
	if cfg.App.Env == "production" || cfg.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Apply global middlewares
	router.Use(middleware.RecoveryMiddleware(logger))
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.TrustGatewayMiddleware()) // Trust headers from Gateway
	// CORS configuration
	env := os.Getenv("CORS_ORIGINS")
	var corsOrigins []string
	if env == "" {
		corsOrigins = []string{"*"}
	} else {
		corsOrigins = strings.Split(env, ",")
	}
	router.Use(middleware.CORSMiddleware(corsOrigins))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, response.Success(gin.H{"status": "ok", "service": "events"}))
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Public routes
		events := api.Group("/events")
		{
			// List events (public, with optional auth for user status)
			events.GET("", eventHandler.ListEventsHandler)
			// Get event by ID (public, with optional auth for user status)
			events.GET("/:id", eventHandler.GetEventHandler)
			// List comments (public)
			events.GET("/:id/comments", eventHandler.ListCommentsHandler)
		}

		// Protected routes (require auth via Gateway)
		protected := api.Group("/events")
		protected.Use(middleware.RequireGatewayAuth()) // Trust Gateway auth
		{
			// Create event
			protected.POST("", eventHandler.CreateEventHandler)
			// Update event
			protected.PATCH("/:id", eventHandler.UpdateEventHandler)
			// Delete event
			protected.DELETE("/:id", eventHandler.DeleteEventHandler)

			// Interest operations
			protected.POST("/:id/interested", eventHandler.AddInterestHandler)
			protected.DELETE("/:id/interested", eventHandler.RemoveInterestHandler)

			// Going operations
			protected.POST("/:id/going", eventHandler.AddGoingHandler)
			protected.DELETE("/:id/going", eventHandler.RemoveGoingHandler)

			// Comment operations
			protected.POST("/:id/comments", eventHandler.CreateCommentHandler)
			protected.DELETE("/:id/comments/:commentId", eventHandler.DeleteCommentHandler)
		}
	}

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
