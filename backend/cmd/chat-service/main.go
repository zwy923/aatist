package main

import (
	"fmt"
	"os"

	"github.com/aatist/backend/internal/chat/handler"
	"github.com/aatist/backend/internal/chat/repository"
	"github.com/aatist/backend/internal/chat/service"
	"github.com/aatist/backend/internal/platform/app"
	"github.com/aatist/backend/internal/platform/middleware"
	"go.uber.org/zap"
)

func main() {
	cfg, err := app.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger, err := app.InitLogger(cfg.App.Env)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	port := cfg.App.HTTPPort
	if p := os.Getenv("HTTP_PORT"); p != "" {
		fmt.Sscanf(p, "%d", &port)
	}
	if port == 0 {
		port = 8088
	}
	cfg.App.HTTPPort = port

	logger.Info("Starting chat service",
		zap.String("env", cfg.App.Env),
		zap.Int("port", cfg.App.HTTPPort),
	)

	postgres, err := app.InitPostgres(cfg.Postgres.DSN, logger)
	if err != nil {
		logger.Fatal("Failed to initialize PostgreSQL", zap.Error(err))
	}
	defer postgres.Close()

	repo := repository.NewPostgresChatRepository(postgres.GetDB())
	chatSvc := service.NewChatService(repo)
	chatHandler := handler.NewChatHandler(chatSvc, logger)

	router := app.NewDefaultRouter(logger, "chat")

	api := router.Group("/api/v1")
	{
		internal := api.Group("/internal/messages")
		internal.Use(middleware.RequireInternalCall())
		{
			internal.POST("", chatHandler.CreateMessageHandler)
		}

		conversations := api.Group("/conversations")
		conversations.Use(middleware.TrustGatewayMiddleware())
		conversations.Use(middleware.RequireGatewayAuth())
		{
			conversations.POST("/start", chatHandler.StartConversationHandler)
			conversations.GET("", chatHandler.GetConversationsHandler)
			conversations.GET("/:id/messages", chatHandler.GetMessagesHandler)
			conversations.PUT("/:id/read", chatHandler.MarkConversationAsReadHandler)
			conversations.PATCH("/:id/read", chatHandler.MarkConversationAsReadHandler)
			conversations.DELETE("/:id", chatHandler.DeleteConversationHandler)
		}
	}

	if err := app.RunServer(cfg.App.HTTPPort, router, logger); err != nil {
		logger.Fatal("Server error", zap.Error(err))
	}
}
