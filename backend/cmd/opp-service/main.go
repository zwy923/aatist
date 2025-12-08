package main

import (
	"fmt"
	"os"

	"github.com/aatist/backend/internal/opportunity/handler"
	"github.com/aatist/backend/internal/opportunity/repository"
	oppservice "github.com/aatist/backend/internal/opportunity/service"
	"github.com/aatist/backend/internal/platform/app"
	"github.com/aatist/backend/internal/platform/middleware"
	userservice "github.com/aatist/backend/internal/user/service"
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

	logger.Info("Starting opportunity service",
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
	oppRepo := repository.NewPostgresOpportunityRepository(postgres.GetDB())
	appRepo := repository.NewPostgresOpportunityApplicationRepository(postgres.GetDB())

	// Initialize services
	oppService := oppservice.NewOpportunityService(oppRepo)
	savedItemClient := userservice.NewHTTPSavedItemClient() // Uses user-service's saved items API
	applicationService := oppservice.NewApplicationService(appRepo, oppRepo)

	// Initialize handler
	oppHandler := handler.NewOpportunityHandler(oppService, savedItemClient, applicationService, logger)

	// Setup Gin router
	router := app.NewDefaultRouter(logger, "opportunity")

	// API routes
	api := router.Group("/api/v1")
	{
		// Public routes
		opportunities := api.Group("/opportunities")
		{
			// List opportunities (public, with optional auth for favorites)
			opportunities.GET("", oppHandler.ListOpportunitiesHandler)
			// Get opportunity by ID (public, with optional auth for favorites)
			opportunities.GET("/:id", oppHandler.GetOpportunityHandler)
		}

		// Protected routes (require auth via Gateway)
		protected := api.Group("/opportunities")
		protected.Use(middleware.RequireGatewayAuth()) // Trust Gateway auth
		{
			// Create opportunity
			protected.POST("", oppHandler.CreateOpportunityHandler)
			// Update opportunity
			protected.PATCH("/:id", oppHandler.UpdateOpportunityHandler)
			// Delete opportunity
			protected.DELETE("/:id", oppHandler.DeleteOpportunityHandler)

			// Save/unsave operations (uses user-service's saved items API)
			protected.POST("/:id/favorite", oppHandler.SaveOpportunityHandler)
			protected.DELETE("/:id/favorite", oppHandler.UnsaveOpportunityHandler)

			// Application operations
			protected.POST("/:id/apply", oppHandler.CreateApplicationHandler)
			// List my applications
			protected.GET("/applications", oppHandler.ListMyApplicationsHandler)
			// List applications for an opportunity (only creator can view)
			protected.GET("/:id/applications", oppHandler.ListOpportunityApplicationsHandler)
		}
	}

	// Start HTTP server with graceful shutdown
	if err := app.RunServer(cfg.App.HTTPPort, router, logger); err != nil {
		logger.Fatal("Server error", zap.Error(err))
	}
}
