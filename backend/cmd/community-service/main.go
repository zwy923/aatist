package main

import (
	"fmt"
	"os"

	"github.com/aatist/backend/internal/community/handler"
	"github.com/aatist/backend/internal/community/repository"
	"github.com/aatist/backend/internal/community/service"
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
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger, err := app.InitLogger(cfg.App.Env)
	if err != nil {
		fmt.Printf("failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting community service",
		zap.String("env", cfg.App.Env),
		zap.Int("port", cfg.App.HTTPPort),
	)

	// Initialize PostgreSQL
	postgres, err := app.InitPostgres(cfg.Postgres.DSN, logger)
	if err != nil {
		logger.Fatal("failed to initialize postgres", zap.Error(err))
	}
	defer postgres.Close()

	// Run database migrations
	if err := app.RunMigrations(postgres, logger); err != nil {
		logger.Fatal("failed to run migrations", zap.Error(err))
	}

	// Initialize Redis (optional)
	var redisClient *cache.Redis
	var redisCmd redis.Cmdable
	redisClient, err = app.InitRedis(cfg.Redis.Addr, cfg.Redis.DB, logger)
	if err != nil {
		logger.Warn("failed to initialize redis", zap.Error(err))
	} else if redisClient != nil {
		defer redisClient.Close()
		redisCmd = redisClient.GetClient()
	} else {
		logger.Warn("Redis configuration missing - trending features disabled")
	}

	// Initialize RabbitMQ (optional)
	var rabbitMQ *mq.RabbitMQ
	var eventPublisher service.EventPublisher
	rabbitMQ, err = app.InitRabbitMQ(cfg.MQ.Broker, cfg.MQ.PublishConfirmTimeout, logger)
	if err != nil {
		logger.Warn("failed to initialize RabbitMQ", zap.Error(err))
	} else if rabbitMQ != nil {
		defer rabbitMQ.Close()
		eventPublisher = rabbitMQ
	}

	// Initialize repositories
	postRepo := repository.NewPostgresPostRepository(postgres.GetDB())
	commentRepo := repository.NewPostgresCommentRepository(postgres.GetDB())
	likeRepo := repository.NewPostgresLikeRepository(postgres.GetDB())
	trendingMgr := service.NewTrendingManager(postRepo, redisCmd, logger)
	engagementUpdater := service.NewEngagementUpdater(redisCmd, trendingMgr, logger)

	// Initialize services
	postSvc := service.NewPostService(postRepo, redisCmd, eventPublisher, trendingMgr, engagementUpdater, logger)
	commentSvc := service.NewCommentService(commentRepo, postRepo, redisCmd, eventPublisher, trendingMgr, engagementUpdater, logger)
	likeSvc := service.NewLikeService(likeRepo, postRepo, redisCmd, eventPublisher, trendingMgr, engagementUpdater, logger)

	// Initialize handlers
	postHandler := handler.NewPostHandler(postSvc, logger)
	commentHandler := handler.NewCommentHandler(commentSvc, logger)
	likeHandler := handler.NewLikeHandler(likeSvc, logger)

	// Setup Gin router
	router := app.NewDefaultRouter(logger, "community")

	// API routes
	api := router.Group("/api/v1")
	community := api.Group("/community")
	{
		// Public post routes
		community.GET("/posts", postHandler.GetPostsHandler)
		community.GET("/posts/trending", postHandler.GetTrendingPostsHandler)
		community.GET("/posts/:id", postHandler.GetPostDetailHandler)
		community.GET("/posts/:id/comments", commentHandler.ListCommentsHandler)

		// Public user posts
		community.GET("/users/:id/posts", postHandler.GetUserPostsHandler)
	}

	// Protected routes
	protected := community.Group("")
	protected.Use(middleware.RequireGatewayAuth())
	{
		// Post routes
		protected.POST("/posts", postHandler.CreatePostHandler)
		protected.PUT("/posts/:id", postHandler.UpdatePostHandler)
		protected.DELETE("/posts/:id", postHandler.DeletePostHandler)

		// Like routes
		protected.POST("/posts/:id/like", likeHandler.LikePostHandler)
		protected.DELETE("/posts/:id/like", likeHandler.UnlikePostHandler)

		// Comment routes
		protected.POST("/posts/:id/comments", commentHandler.CreateCommentHandler)

		// Current user's posts
		protected.GET("/users/me/posts", postHandler.GetMyPostsHandler)
	}

	// Comment management routes
	commentRoutes := api.Group("/community/comments")
	commentRoutes.Use(middleware.RequireGatewayAuth())
	{
		commentRoutes.PUT("/:id", commentHandler.UpdateCommentHandler)
		commentRoutes.DELETE("/:id", commentHandler.DeleteCommentHandler)
	}

	// Start HTTP server with graceful shutdown
	if err := app.RunServer(cfg.App.HTTPPort, router, logger); err != nil {
		logger.Fatal("Server error", zap.Error(err))
	}
}
