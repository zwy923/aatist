package main

import (
	"fmt"
	"os"

	"github.com/aalto-talent-network/backend/internal/community/handler"
	"github.com/aalto-talent-network/backend/internal/community/repository"
	"github.com/aalto-talent-network/backend/internal/community/service"
	"github.com/aalto-talent-network/backend/internal/platform/app"
	"github.com/aalto-talent-network/backend/internal/platform/cache"
	"github.com/aalto-talent-network/backend/internal/platform/middleware"
	"github.com/aalto-talent-network/backend/internal/platform/mq"
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

	// Initialize handler
	communityHandler := handler.NewCommunityHandler(postSvc, commentSvc, likeSvc, logger)

	// Setup Gin router
	router := app.NewDefaultRouter(logger, "community")

	// API routes
	api := router.Group("/api/v1")
	community := api.Group("/community")
	{
		community.GET("/posts", communityHandler.GetPostsHandler)
		community.GET("/posts/trending", communityHandler.GetTrendingPostsHandler)
		community.GET("/posts/:id", communityHandler.GetPostDetailHandler)
		community.GET("/posts/:id/comments", communityHandler.ListCommentsHandler)

		// Public user posts
		community.GET("/users/:id/posts", communityHandler.GetUserPostsHandler)
	}

	protected := community.Group("")
	protected.Use(middleware.RequireGatewayAuth())
	{
		protected.POST("/posts", communityHandler.CreatePostHandler)
		protected.PUT("/posts/:id", communityHandler.UpdatePostHandler)
		protected.DELETE("/posts/:id", communityHandler.DeletePostHandler)
		protected.POST("/posts/:id/like", communityHandler.LikePostHandler)
		protected.DELETE("/posts/:id/like", communityHandler.UnlikePostHandler)
		protected.POST("/posts/:id/comments", communityHandler.CreateCommentHandler)

		// Current user's posts
		protected.GET("/users/me/posts", communityHandler.GetMyPostsHandler)
	}

	commentRoutes := api.Group("/community/comments")
	commentRoutes.Use(middleware.RequireGatewayAuth())
	{
		commentRoutes.PUT("/:id", communityHandler.UpdateCommentHandler)
		commentRoutes.DELETE("/:id", communityHandler.DeleteCommentHandler)
	}

	// Start HTTP server with graceful shutdown
	if err := app.RunServer(cfg.App.HTTPPort, router, logger); err != nil {
		logger.Fatal("Server error", zap.Error(err))
	}
}
