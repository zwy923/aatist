package main

import (
	"fmt"
	"os"

	"github.com/aatist/backend/internal/platform/app"
	"github.com/aatist/backend/internal/platform/middleware"
	"github.com/aatist/backend/internal/portfolio/handler"
	"github.com/aatist/backend/internal/portfolio/repository"
	"github.com/aatist/backend/internal/portfolio/service"
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

	logger.Info("Starting portfolio service",
		zap.String("env", cfg.App.Env),
		zap.Int("port", cfg.App.HTTPPort),
	)

	// Initialize PostgreSQL
	postgres, err := app.InitPostgres(cfg.Postgres.DSN, logger)
	if err != nil {
		logger.Fatal("Failed to initialize PostgreSQL", zap.Error(err))
	}
	defer postgres.Close()

	// Initialize repositories
	projectRepo := repository.NewPostgresProjectRepository(postgres.GetDB())

	// Initialize user service client (for checking profile visibility)
	// All internal calls go through Gateway
	userClient := service.NewHTTPUserServiceClient(os.Getenv("GATEWAY_URL"))

	// Initialize services
	projectService := service.NewProjectService(projectRepo, userClient)

	// Initialize handler
	portfolioHandler := handler.NewPortfolioHandler(projectService, logger)

	// Setup Gin router
	router := app.NewDefaultRouter(logger, "portfolio")

	// API routes
	api := router.Group("/api/v1")
	{
		// Public portfolio routes
		portfolio := api.Group("/portfolio")
		{
			portfolio.GET("", portfolioHandler.GetPublicProjectsHandler)
			portfolio.GET("/:id", portfolioHandler.GetProjectDetailHandler)
		}

		// Public user portfolio route
		users := api.Group("/users")
		{
			users.GET("/:id/portfolio", portfolioHandler.GetUserPortfolioHandler)
		}

		// Protected portfolio routes (require auth via Gateway)
		protectedPortfolio := api.Group("/users")
		protectedPortfolio.Use(middleware.RequireGatewayAuth()) // Require Gateway to set user identity
		{
			protectedPortfolio.GET("/me/portfolio", portfolioHandler.GetMyPortfolioHandler)
			protectedPortfolio.POST("/me/portfolio", portfolioHandler.CreateProjectHandler)
			protectedPortfolio.PATCH("/me/portfolio/:id", portfolioHandler.UpdateProjectHandler)
			protectedPortfolio.DELETE("/me/portfolio/:id", portfolioHandler.DeleteProjectHandler)
		}
	}

	// Start HTTP server with graceful shutdown
	if err := app.RunServer(cfg.App.HTTPPort, router, logger); err != nil {
		logger.Fatal("Server error", zap.Error(err))
	}
}
