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
	"github.com/aalto-talent-network/backend/internal/platform/db"
	"github.com/aalto-talent-network/backend/internal/platform/log"
	"github.com/aalto-talent-network/backend/internal/platform/middleware"
	"github.com/aalto-talent-network/backend/internal/platform/mq"
	"github.com/aalto-talent-network/backend/internal/user/handler"
	"github.com/aalto-talent-network/backend/internal/user/repository"
	"github.com/aalto-talent-network/backend/internal/user/service"
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

	logger.Info("Starting user service",
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

	// Initialize Redis
	redis, err := cache.NewRedis(cfg.Redis.Addr, cfg.Redis.DB)
	if err != nil {
		logger.Fatal("Failed to initialize Redis", zap.Error(err))
	}
	defer redis.Close()

	logger.Info("Connected to Redis")

	// Initialize file service client (for avatar uploads)
	fileClient := service.NewHTTPFileServiceClient()

	// Initialize JWT
	jwt := auth.NewJWT(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)

	// Initialize repositories
	userRepo := repository.NewPostgresRepository(postgres.GetDB())
	savedItemRepo := repository.NewPostgresSavedItemRepository(postgres.GetDB())

	// Initialize MQ (optional - only if broker is configured)
	var rabbitMQ *mq.RabbitMQ
	if cfg.MQ.Broker != "" {
		var err error
		rabbitMQ, err = mq.NewRabbitMQ(cfg.MQ.Broker, cfg.MQ.PublishConfirmTimeout, logger)
		if err != nil {
			logger.Warn("Failed to initialize RabbitMQ - email verification will be disabled", zap.Error(err))
		} else {
			defer rabbitMQ.Close()
			logger.Info("Connected to RabbitMQ")
		}
	}

	// Initialize email verification service
	emailVerifSvc := service.NewEmailVerificationService(userRepo, redis, logger)

	// Initialize service
	authService := service.NewAuthService(userRepo, jwt, redis, logger, emailVerifSvc)
	avatarURLPrefix := cfg.S3.PublicURL
	if avatarURLPrefix == "" {
		avatarURLPrefix = "http://localhost:9000/files" // Default file service public URL
	}
	profileService := service.NewProfileService(userRepo, fileClient, redis, logger, avatarURLPrefix)
	savedItemService := service.NewSavedItemService(savedItemRepo)
	notificationClient := service.NewHTTPNotificationClient()

	// Prepare MQ publisher (may be nil)
	var emailPublisher interface {
		PublishEmailVerification(message interface{}) error
	}
	if rabbitMQ != nil {
		emailPublisher = rabbitMQ
	}

	// Initialize handler
	authHandler := handler.NewAuthHandler(authService, profileService, savedItemService, notificationClient, emailVerifSvc, emailPublisher, logger)

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
	router.GET("/user/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, response.Success(gin.H{"status": "ok", "service": "user"}))
	})

	// API routes
	api := router.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.RegisterHandler)
			auth.POST("/login", authHandler.LoginHandler)
			auth.POST("/refresh", authHandler.RefreshTokenHandler)
			auth.POST("/logout", authHandler.LogoutHandler)
			auth.POST("/verify-email", authHandler.VerifyEmailHandler)
			auth.GET("/verify", authHandler.VerifyEmailGetHandler)
		}

		// Public user routes
		users := api.Group("/users")
		{
			users.GET("/:id", authHandler.GetUserByIDHandler)
		}

		// Protected user routes (require auth via Gateway)
		protectedUsers := api.Group("/users")
		protectedUsers.Use(middleware.RequireGatewayAuth()) // Trust Gateway auth
		{
			protectedUsers.GET("/me", authHandler.GetCurrentUserHandler)
			protectedUsers.PATCH("/me", authHandler.UpdateCurrentUserHandler)
			protectedUsers.POST("/me/avatar", authHandler.UploadAvatarHandler)

			// Availability
			protectedUsers.GET("/me/availability", authHandler.GetAvailabilityHandler)
			protectedUsers.PATCH("/me/availability", authHandler.UpdateAvailabilityHandler)

			// Saved items
			protectedUsers.GET("/me/saved", authHandler.GetSavedItemsHandler)
			protectedUsers.POST("/me/saved", authHandler.SaveItemHandler)
			protectedUsers.DELETE("/me/saved", authHandler.UnsaveItemHandler)
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
