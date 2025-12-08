package main

import (
	"fmt"
	"os"

	"github.com/aatist/backend/internal/platform/app"
	"github.com/aatist/backend/internal/platform/auth"
	"github.com/aatist/backend/internal/platform/middleware"
	"github.com/aatist/backend/internal/platform/mq"
	"github.com/aatist/backend/internal/user/handler"
	"github.com/aatist/backend/internal/user/repository"
	"github.com/aatist/backend/internal/user/service"
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

	logger.Info("Starting user service",
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

	// Initialize Redis
	redis, err := app.InitRedis(cfg.Redis.Addr, cfg.Redis.DB, logger)
	if err != nil {
		logger.Fatal("Failed to initialize Redis", zap.Error(err))
	}
	defer redis.Close()

	// Initialize file service client (for avatar uploads)
	fileClient := service.NewHTTPFileServiceClient()

	// Initialize JWT
	jwt := auth.NewJWT(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)

	// Initialize repositories
	userRepo := repository.NewPostgresRepository(postgres.GetDB())
	savedItemRepo := repository.NewPostgresSavedItemRepository(postgres.GetDB())

	// Initialize MQ (optional - only if broker is configured)
	var rabbitMQ *mq.RabbitMQ
	rabbitMQ, err = app.InitRabbitMQ(cfg.MQ.Broker, cfg.MQ.PublishConfirmTimeout, logger)
	if err != nil {
		logger.Warn("Failed to initialize RabbitMQ - email verification will be disabled", zap.Error(err))
	} else if rabbitMQ != nil {
		defer rabbitMQ.Close()
	}

	// Initialize email verification service
	emailVerifSvc := service.NewEmailVerificationService(userRepo, redis, logger)

	// Initialize password reset service
	passwordResetSvc := service.NewPasswordResetService(userRepo, redis, logger)

	// Initialize service with auto-verified email domains from config
	autoVerifiedDomains := cfg.Email.AutoVerifiedDomains
	if len(autoVerifiedDomains) == 0 {
		// Default to @aalto.fi if not configured
		autoVerifiedDomains = []string{"@aalto.fi"}
		logger.Info("No auto-verified domains configured, using default: @aalto.fi")
	} else {
		logger.Info("Auto-verified email domains configured",
			zap.Strings("domains", autoVerifiedDomains),
		)
	}
	authService := service.NewAuthService(userRepo, jwt, redis, logger, emailVerifSvc, autoVerifiedDomains)
	avatarURLPrefix := cfg.S3.PublicURL
	if avatarURLPrefix == "" {
		avatarURLPrefix = "http://localhost:9000/files" // Default file service public URL
	}
	profileService := service.NewProfileService(userRepo, fileClient, redis, logger, avatarURLPrefix)
	savedItemService := service.NewSavedItemService(savedItemRepo)
	notificationClient := service.NewHTTPNotificationClient()

	// Prepare MQ publisher (may be nil)
	var mqPublisher interface {
		PublishEmailVerification(message interface{}) error
		PublishPasswordReset(message interface{}) error
	}
	if rabbitMQ != nil {
		mqPublisher = rabbitMQ
	}

	// Initialize handler
	authHandler := handler.NewAuthHandler(authService, profileService, savedItemService, notificationClient, emailVerifSvc, passwordResetSvc, mqPublisher, logger)

	// Setup Gin router
	router := app.NewDefaultRouter(logger, "user")

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

			// Password reset (forgot password)
			auth.POST("/forgot-password", authHandler.ForgotPasswordHandler)
			auth.POST("/reset-password", authHandler.ResetPasswordHandler)
		}

		// Public user routes
		users := api.Group("/users")
		{
			// Check email availability (for registration validation)
			users.GET("/check-email", authHandler.CheckEmailHandler)

			// Get user by ID
			users.GET("/:id", authHandler.GetUserByIDHandler)

			// Get user summary (lightweight profile for cards/lists)
			users.GET("/:id/summary", authHandler.GetUserSummaryHandler)
		}

		// Protected user routes (require auth via Gateway)
		protectedUsers := api.Group("/users")
		protectedUsers.Use(middleware.RequireGatewayAuth()) // Trust Gateway auth
		{
			protectedUsers.GET("/me", authHandler.GetCurrentUserHandler)
			protectedUsers.PATCH("/me", authHandler.UpdateCurrentUserHandler)
			protectedUsers.POST("/me/avatar", authHandler.UploadAvatarHandler)

			// Password management
			protectedUsers.PATCH("/me/password", authHandler.ChangePasswordHandler)

			// Availability
			protectedUsers.GET("/me/availability", authHandler.GetAvailabilityHandler)
			protectedUsers.PATCH("/me/availability", authHandler.UpdateAvailabilityHandler)

			// Saved items
			protectedUsers.GET("/me/saved", authHandler.GetSavedItemsHandler)
			protectedUsers.POST("/me/saved", authHandler.SaveItemHandler)
			protectedUsers.DELETE("/me/saved", authHandler.UnsaveItemHandler)
		}
	}

	// Start HTTP server with graceful shutdown
	if err := app.RunServer(cfg.App.HTTPPort, router, logger); err != nil {
		logger.Fatal("Server error", zap.Error(err))
	}
}
