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

	"github.com/aalto-talent-network/backend/internal/platform/config"
	"github.com/aalto-talent-network/backend/internal/platform/db"
	"github.com/aalto-talent-network/backend/internal/platform/log"
	"github.com/aalto-talent-network/backend/internal/platform/middleware"
	"github.com/aalto-talent-network/backend/internal/portfolio/handler"
	"github.com/aalto-talent-network/backend/internal/portfolio/repository"
	"github.com/aalto-talent-network/backend/internal/portfolio/service"
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

	logger.Info("Starting portfolio service",
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
		c.JSON(http.StatusOK, response.Success(gin.H{"status": "ok", "service": "portfolio"}))
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Public portfolio routes
		portfolio := api.Group("/portfolio")
		{
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
			protectedPortfolio.PUT("/me/portfolio/:id", portfolioHandler.UpdateProjectHandler)
			protectedPortfolio.DELETE("/me/portfolio/:id", portfolioHandler.DeleteProjectHandler)
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
