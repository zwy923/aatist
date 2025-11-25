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

	"github.com/aalto-talent-network/backend/internal/community/handler"
	"github.com/aalto-talent-network/backend/internal/community/repository"
	"github.com/aalto-talent-network/backend/internal/community/service"
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
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "configs/config.yaml"
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger, err := log.NewLogger(cfg.App.Env)
	if err != nil {
		fmt.Printf("failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting community service",
		zap.String("env", cfg.App.Env),
		zap.Int("port", cfg.App.HTTPPort),
	)

	postgres, err := db.NewPostgres(cfg.Postgres.DSN)
	if err != nil {
		logger.Fatal("failed to initialize postgres", zap.Error(err))
	}
	defer postgres.Close()

	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}
	if err := db.RunMigrations(postgres.GetSQLDB(), migrationsDir); err != nil {
		logger.Fatal("failed to run migrations", zap.Error(err))
	}
	logger.Info("Database migrations completed")

	var redisClient *cache.Redis
	var redisCmd redis.Cmdable
	if cfg.Redis.Addr != "" {
		redisClient, err = cache.NewRedis(cfg.Redis.Addr, cfg.Redis.DB)
		if err != nil {
			logger.Warn("failed to initialize redis", zap.Error(err))
		} else {
			defer redisClient.Close()
			redisCmd = redisClient.GetClient()
			logger.Info("Connected to Redis", zap.String("addr", cfg.Redis.Addr))
		}
	} else {
		logger.Warn("Redis configuration missing - trending features disabled")
	}

	var rabbitMQ *mq.RabbitMQ
	var eventPublisher service.EventPublisher
	if cfg.MQ.Broker != "" {
		rabbitMQ, err = mq.NewRabbitMQ(cfg.MQ.Broker, cfg.MQ.PublishConfirmTimeout, logger)
		if err != nil {
			logger.Warn("failed to initialize RabbitMQ", zap.Error(err))
		} else {
			defer rabbitMQ.Close()
			eventPublisher = rabbitMQ
			logger.Info("Connected to RabbitMQ")
		}
	}

	postRepo := repository.NewPostgresPostRepository(postgres.GetDB())
	commentRepo := repository.NewPostgresCommentRepository(postgres.GetDB())
	likeRepo := repository.NewPostgresLikeRepository(postgres.GetDB())
	trendingMgr := service.NewTrendingManager(postRepo, redisCmd, logger)
	engagementUpdater := service.NewEngagementUpdater(redisCmd, trendingMgr, logger)

	postSvc := service.NewPostService(postRepo, redisCmd, eventPublisher, trendingMgr, engagementUpdater, logger)
	commentSvc := service.NewCommentService(commentRepo, postRepo, redisCmd, eventPublisher, trendingMgr, engagementUpdater, logger)
	likeSvc := service.NewLikeService(likeRepo, postRepo, redisCmd, eventPublisher, trendingMgr, engagementUpdater, logger)

	communityHandler := handler.NewCommunityHandler(postSvc, commentSvc, likeSvc, logger)

	if cfg.App.Env == "production" || cfg.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(middleware.RecoveryMiddleware(logger))
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.TrustGatewayMiddleware())

	env := os.Getenv("CORS_ORIGINS")
	var corsOrigins []string
	if env == "" {
		corsOrigins = []string{"*"}
	} else {
		corsOrigins = strings.Split(env, ",")
	}
	router.Use(middleware.CORSMiddleware(corsOrigins))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, response.Success(gin.H{"status": "ok", "service": "community"}))
	})

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

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.App.HTTPPort),
		Handler: router,
	}

	go func() {
		logger.Info("HTTP server starting", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down community service")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("forced shutdown", zap.Error(err))
	}
	logger.Info("Community service exited")
}
