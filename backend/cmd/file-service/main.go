package main

import (
	"fmt"
	"os"

	"github.com/aatist/backend/internal/file/handler"
	"github.com/aatist/backend/internal/file/repository"
	"github.com/aatist/backend/internal/file/service"
	"github.com/aatist/backend/internal/platform/app"
	"github.com/aatist/backend/internal/platform/middleware"
	"github.com/aatist/backend/internal/platform/storage"
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

	logger.Info("Starting file service",
		zap.String("env", cfg.App.Env),
		zap.Int("port", cfg.App.HTTPPort),
	)

	// Initialize PostgreSQL
	postgres, err := app.InitPostgres(cfg.Postgres.DSN, logger)
	if err != nil {
		logger.Fatal("Failed to initialize PostgreSQL", zap.Error(err))
	}
	defer postgres.Close()

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
	// Note: TrustGatewayMiddleware is only applied to user-facing routes, not globally
	router := app.NewDefaultRouter(logger, "file")

	// API routes
	api := router.Group("/api/v1")
	{
		// Internal API routes (for service-to-service communication)
		// These routes require internal authentication
		internal := api.Group("/internal/file")
		internal.Use(middleware.RequireInternalCall())    // Require internal service authentication
		internal.Use(middleware.TrustGatewayMiddleware()) // Extract user identity from headers (set by Gateway)
		{
			internal.POST("/upload", fileHandler.UploadFileHandler)
			internal.GET("/:id", fileHandler.GetFileHandler)
			internal.DELETE("/:id", fileHandler.DeleteFileHandler)
		}

		// User-facing file routes (require auth via Gateway)
		files := api.Group("/files")
		files.Use(middleware.TrustGatewayMiddleware()) // Trust Gateway headers
		files.Use(middleware.RequireGatewayAuth())     // Require Gateway to set user identity
		{
			files.POST("/upload", fileHandler.UploadFileHandler)
			files.POST("/presigned-upload", fileHandler.GeneratePresignedUploadURLHandler)
			files.POST("/confirm-upload", fileHandler.ConfirmUploadHandler)
			files.GET("", fileHandler.GetUserFilesHandler)
			files.GET("/:id", fileHandler.GetFileHandler)
			files.DELETE("/:id", fileHandler.DeleteFileHandler)
		}
	}

	// Start HTTP server with graceful shutdown
	if err := app.RunServer(cfg.App.HTTPPort, router, logger); err != nil {
		logger.Fatal("Server error", zap.Error(err))
	}
}
