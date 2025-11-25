package main

import (
	"fmt"
	"os"

	"github.com/aalto-talent-network/backend/internal/event/handler"
	"github.com/aalto-talent-network/backend/internal/event/repository"
	"github.com/aalto-talent-network/backend/internal/event/service"
	"github.com/aalto-talent-network/backend/internal/platform/app"
	"github.com/aalto-talent-network/backend/internal/platform/middleware"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := app.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger, err := app.InitLogger(cfg.App.Env)
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
	postgres, err := app.InitPostgres(cfg.Postgres.DSN, logger)
	if err != nil {
		logger.Fatal("Failed to initialize PostgreSQL", zap.Error(err))
	}
	defer postgres.Close()

	// Run database migrations
	if err := app.RunMigrations(postgres, logger); err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}

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
	router := app.NewDefaultRouter(logger, "events")

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

	// Start HTTP server with graceful shutdown
	if err := app.RunServer(cfg.App.HTTPPort, router, logger); err != nil {
		logger.Fatal("Server error", zap.Error(err))
	}
}
