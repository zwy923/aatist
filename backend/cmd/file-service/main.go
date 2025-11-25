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

	"github.com/aalto-talent-network/backend/internal/file/handler"
	"github.com/aalto-talent-network/backend/internal/file/repository"
	"github.com/aalto-talent-network/backend/internal/file/service"
	"github.com/aalto-talent-network/backend/internal/platform/config"
	"github.com/aalto-talent-network/backend/internal/platform/db"
	"github.com/aalto-talent-network/backend/internal/platform/log"
	"github.com/aalto-talent-network/backend/internal/platform/middleware"
	"github.com/aalto-talent-network/backend/internal/platform/storage"
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

	logger.Info("Starting file service",
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

	// Initialize S3/MinIO storage
	s3Client, err := storage.NewS3(storage.S3Config{
		Endpoint:  cfg.S3.Endpoint,
		AccessKey: cfg.S3.AccessKey,
		SecretKey: cfg.S3.SecretKey,
		UseSSL:    cfg.S3.UseSSL,
		Bucket:    cfg.S3.Bucket,
		PublicURL: cfg.S3.PublicURL,
	})
	if err != nil {
		logger.Fatal("Failed to initialize S3 storage", zap.Error(err))
	}

	// Initialize repositories
	fileRepo := repository.NewPostgresFileRepository(postgres.GetDB())

	// Initialize services
	fileService := service.NewFileService(fileRepo, s3Client, logger)

	// Initialize handler
	fileHandler := handler.NewFileHandler(fileService, logger)

	// Setup Gin router
	if cfg.App.Env == "production" || cfg.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Apply global middlewares
	router.Use(middleware.RecoveryMiddleware(logger))
	router.Use(middleware.RequestIDMiddleware())
	// Note: TrustGatewayMiddleware is only applied to user-facing routes, not globally

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
		c.JSON(http.StatusOK, response.Success(gin.H{"status": "ok", "service": "file"}))
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Internal API routes (for service-to-service communication)
		// These routes require internal authentication
		internal := api.Group("/internal/file")
		internal.Use(middleware.RequireInternalCall()) // Require internal service authentication
		internal.Use(middleware.TrustGatewayMiddleware()) // Extract user identity from headers (set by Gateway)
		{
			internal.POST("/upload", fileHandler.UploadFileHandler)
			internal.GET("/:id", fileHandler.GetFileHandler)
			internal.DELETE("/:id", fileHandler.DeleteFileHandler)
		}

		// User-facing file routes (require auth via Gateway)
		files := api.Group("/files")
		files.Use(middleware.TrustGatewayMiddleware()) // Trust Gateway headers
		files.Use(middleware.RequireGatewayAuth())    // Require Gateway to set user identity
		{
			files.POST("/upload", fileHandler.UploadFileHandler)
			files.POST("/presigned-upload", fileHandler.GeneratePresignedUploadURLHandler)
			files.POST("/confirm-upload", fileHandler.ConfirmUploadHandler)
			files.GET("", fileHandler.GetUserFilesHandler)
			files.GET("/:id", fileHandler.GetFileHandler)
			files.DELETE("/:id", fileHandler.DeleteFileHandler)
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

