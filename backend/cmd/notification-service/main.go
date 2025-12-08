package main

import (
	"fmt"
	"os"

	"github.com/aatist/backend/internal/notification/handler"
	"github.com/aatist/backend/internal/notification/repository"
	"github.com/aatist/backend/internal/notification/service"
	"github.com/aatist/backend/internal/platform/app"
	"github.com/aatist/backend/internal/platform/cache"
	"github.com/aatist/backend/internal/platform/middleware"
	"github.com/aatist/backend/internal/platform/mq"
	"github.com/redis/go-redis/v9"
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

	logger.Info("Starting notification service",
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

	// Initialize Redis (optional - used for batching)
	var redisClient *cache.Redis
	var redisCmd redis.Cmdable
	redisClient, err = app.InitRedis(cfg.Redis.Addr, cfg.Redis.DB, logger)
	if err != nil {
		logger.Warn("Failed to initialize Redis - community batching disabled", zap.Error(err))
	} else if redisClient != nil {
		defer redisClient.Close()
		redisCmd = redisClient.GetClient()
	}

	// Initialize RabbitMQ (optional - used for community events)
	var rabbitMQ *mq.RabbitMQ
	rabbitMQ, err = app.InitRabbitMQ(cfg.MQ.Broker, cfg.MQ.PublishConfirmTimeout, logger)
	if err != nil {
		logger.Warn("Failed to initialize RabbitMQ - community events disabled", zap.Error(err))
	} else if rabbitMQ != nil {
		defer rabbitMQ.Close()
	}

	// Initialize repositories
	notificationRepo := repository.NewPostgresNotificationRepository(postgres.GetDB())

	// Initialize services
	notificationService := service.NewNotificationService(notificationRepo)

	if rabbitMQ != nil {
		consumer := service.NewCommunityEventConsumer(notificationService, redisCmd, logger)
		// Use SERVICE_NAME env var → queues: {SERVICE_NAME}.community, {SERVICE_NAME}.community.dlq, {SERVICE_NAME}.community.retry
		serviceName := os.Getenv("SERVICE_NAME")
		if serviceName == "" {
			serviceName = "notification-service"
		}
		if err := rabbitMQ.ConsumeCommunityEvents(
			serviceName,
			[]string{"community.post.created", "community.post.liked", "community.post.commented"},
			func(eventType string, payload []byte) error {
				return consumer.HandleMessage(eventType, payload)
			},
		); err != nil {
			logger.Warn("Failed to start community event consumer", zap.Error(err))
		}
	}

	// Initialize handler
	notificationHandler := handler.NewNotificationHandler(notificationService, logger)

	// Setup Gin router
	router := app.NewDefaultRouter(logger, "notification")

	// API routes
	api := router.Group("/api/v1")
	{
		// Internal API for creating notifications (for other services)
		// This should only be accessible from internal services (gateway, other microservices)
		internal := api.Group("/internal/notifications")
		internal.Use(middleware.RequireInternalCall()) // Require internal service authentication
		{
			internal.POST("", notificationHandler.CreateNotificationHandler)
		}

		// User-facing notification routes (require auth via Gateway)
		// Use /notifications instead of /me/notifications to avoid Gin wildcard conflicts in Gateway
		userNotifications := api.Group("/notifications")
		userNotifications.Use(middleware.TrustGatewayMiddleware()) // Trust Gateway headers
		userNotifications.Use(middleware.RequireGatewayAuth())     // Require Gateway to set user identity
		{
			userNotifications.GET("", notificationHandler.GetNotificationsHandler)
			userNotifications.GET("/unread-count", notificationHandler.GetUnreadCountHandler)
			userNotifications.PUT("/:id/read", notificationHandler.MarkNotificationAsReadHandler)
			userNotifications.PUT("/read-all", notificationHandler.MarkAllNotificationsAsReadHandler)
			userNotifications.DELETE("/:id", notificationHandler.DeleteNotificationHandler)
			userNotifications.DELETE("", notificationHandler.DeleteMultipleNotificationsHandler) // Batch delete
		}
	}

	// Start HTTP server with graceful shutdown
	if err := app.RunServer(cfg.App.HTTPPort, router, logger); err != nil {
		logger.Fatal("Server error", zap.Error(err))
	}
}
