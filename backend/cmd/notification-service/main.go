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

	"github.com/aalto-talent-network/backend/internal/notification/handler"
	"github.com/aalto-talent-network/backend/internal/notification/repository"
	"github.com/aalto-talent-network/backend/internal/notification/service"
	"github.com/aalto-talent-network/backend/internal/platform/cache"
	"github.com/aalto-talent-network/backend/internal/platform/config"
	"github.com/aalto-talent-network/backend/internal/platform/db"
	"github.com/aalto-talent-network/backend/internal/platform/log"
	"github.com/aalto-talent-network/backend/internal/platform/middleware"
	"github.com/aalto-talent-network/backend/internal/platform/mq"
	"github.com/aalto-talent-network/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
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

	logger.Info("Starting notification service",
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

	// Initialize Redis (optional - used for batching)
	var redisClient *cache.Redis
	var redisCmd redis.Cmdable
	if cfg.Redis.Addr != "" {
		redisClient, err = cache.NewRedis(cfg.Redis.Addr, cfg.Redis.DB)
		if err != nil {
			logger.Warn("Failed to initialize Redis - community batching disabled", zap.Error(err))
		} else {
			defer redisClient.Close()
			redisCmd = redisClient.GetClient()
			logger.Info("Connected to Redis", zap.String("addr", cfg.Redis.Addr))
		}
	}

	// Initialize RabbitMQ (optional - used for community events)
	var rabbitMQ *mq.RabbitMQ
	if cfg.MQ.Broker != "" {
		rabbitMQ, err = mq.NewRabbitMQ(cfg.MQ.Broker, cfg.MQ.PublishConfirmTimeout, logger)
		if err != nil {
			logger.Warn("Failed to initialize RabbitMQ - community events disabled", zap.Error(err))
		} else {
			defer rabbitMQ.Close()
			logger.Info("Connected to RabbitMQ")
		}
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
	if cfg.App.Env == "production" || cfg.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Apply global middlewares
	router.Use(middleware.RecoveryMiddleware(logger))
	router.Use(middleware.RequestIDMiddleware())
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
	router.GET("/notification/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, response.Success(gin.H{"status": "ok", "service": "notification"}))
	})

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
