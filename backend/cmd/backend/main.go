package main

import (
	"fmt"
	"os"

	"github.com/aatist/backend/internal/backend/adapters"
	notifhandler "github.com/aatist/backend/internal/notification/handler"
	notifrepo "github.com/aatist/backend/internal/notification/repository"
	notifservice "github.com/aatist/backend/internal/notification/service"
	opphandler "github.com/aatist/backend/internal/opportunity/handler"
	opprepository "github.com/aatist/backend/internal/opportunity/repository"
	oppservice "github.com/aatist/backend/internal/opportunity/service"
	"github.com/aatist/backend/internal/platform/app"
	"github.com/aatist/backend/internal/platform/auth"
	"github.com/aatist/backend/internal/platform/middleware"
	"github.com/aatist/backend/internal/platform/mq"
	portfoliohandler "github.com/aatist/backend/internal/portfolio/handler"
	portfoliorepo "github.com/aatist/backend/internal/portfolio/repository"
	portfolioservice "github.com/aatist/backend/internal/portfolio/service"
	authhandler "github.com/aatist/backend/internal/user/handler"
	userrepository "github.com/aatist/backend/internal/user/repository"
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

	logger.Info("Starting backend (modular monolith)",
		zap.String("env", cfg.App.Env),
		zap.Int("port", cfg.App.HTTPPort),
	)

	// Initialize PostgreSQL
	postgres, err := app.InitPostgres(cfg.Postgres.DSN, logger)
	if err != nil {
		logger.Fatal("Failed to initialize PostgreSQL", zap.Error(err))
	}
	defer postgres.Close()

	// Initialize Redis
	redis, err := app.InitRedis(cfg.Redis.Addr, cfg.Redis.DB, logger)
	if err != nil {
		logger.Fatal("Failed to initialize Redis", zap.Error(err))
	}
	defer redis.Close()

	// Initialize file service client (external service - HTTP)
	fileClient := service.NewHTTPFileServiceClient()

	// Initialize JWT
	jwt := auth.NewJWT(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)

	// Initialize MQ (optional)
	var rabbitMQ *mq.RabbitMQ
	rabbitMQ, err = app.InitRabbitMQ(cfg.MQ.Broker, cfg.MQ.PublishConfirmTimeout, logger)
	if err != nil {
		logger.Warn("Failed to initialize RabbitMQ - email verification will be disabled", zap.Error(err))
	} else if rabbitMQ != nil {
		defer rabbitMQ.Close()
	}

	// ========== User module ==========
	userRepo := userrepository.NewPostgresRepository(postgres.GetDB())
	savedItemRepo := userrepository.NewPostgresSavedItemRepository(postgres.GetDB())
	userServiceRepo := userrepository.NewPostgresUserServiceRepository(postgres.GetDB())

	emailVerifSvc := service.NewEmailVerificationService(userRepo, redis, logger)
	passwordResetSvc := service.NewPasswordResetService(userRepo, redis, logger)

	autoVerifiedDomains := cfg.Email.AutoVerifiedDomains
	if len(autoVerifiedDomains) == 0 {
		autoVerifiedDomains = []string{"@aalto.fi"}
		logger.Info("No auto-verified domains configured, using default: @aalto.fi")
	}

	authService := service.NewAuthService(userRepo, jwt, redis, logger, emailVerifSvc, autoVerifiedDomains, cfg.Email.DisableEmailVerification)
	avatarURLPrefix := cfg.S3.PublicURL
	if avatarURLPrefix == "" {
		avatarURLPrefix = "http://localhost:9000/files"
	}
	profileService, err := service.NewProfileService(userRepo, fileClient, redis, logger, avatarURLPrefix)
	if err != nil {
		logger.Fatal("Failed to initialize profile service", zap.Error(err))
	}
	savedItemService := service.NewSavedItemService(savedItemRepo)

	// Initialize notification service (in-process) and notification client
	notificationRepo := notifrepo.NewPostgresNotificationRepository(postgres.GetDB())
	notificationService := notifservice.NewNotificationService(notificationRepo)
	notificationClient := adapters.NewLocalNotificationClient(notificationService)

	var mqPublisher interface {
		PublishEmailVerification(message interface{}) error
		PublishPasswordReset(message interface{}) error
	}
	if rabbitMQ != nil {
		mqPublisher = rabbitMQ
	}

	authHandler := authhandler.NewAuthHandler(authService, profileService, savedItemService, userServiceRepo, notificationClient, emailVerifSvc, passwordResetSvc, mqPublisher, cfg.Email.DisableEmailVerification, logger)

	// ========== Portfolio module ==========
	projectRepo := portfoliorepo.NewPostgresProjectRepository(postgres.GetDB())
	userClient := adapters.NewLocalUserServiceClient(userRepo)
	projectService := portfolioservice.NewProjectService(projectRepo, userClient)
	portfolioHandler := portfoliohandler.NewPortfolioHandler(projectService, logger)

	// ========== Opportunity module ==========
	oppRepo := opprepository.NewPostgresOpportunityRepository(postgres.GetDB())
	appRepo := opprepository.NewPostgresOpportunityApplicationRepository(postgres.GetDB())
	oppService := oppservice.NewOpportunityService(oppRepo)
	savedItemClient := adapters.NewLocalSavedItemClient(savedItemService)
	applicationService := oppservice.NewApplicationService(appRepo, oppRepo)
	oppHandler := opphandler.NewOpportunityHandler(oppService, savedItemClient, applicationService, logger)

	// ========== Notification module ==========
	notificationHandler := notifhandler.NewNotificationHandler(notificationService, logger)

	// ========== Setup router ==========
	router := app.NewDefaultRouter(logger, "backend")
	api := router.Group("/api/v1")

	// ----- Auth routes (public) -----
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/register", authHandler.RegisterHandler)
		authGroup.POST("/login", authHandler.LoginHandler)
		authGroup.POST("/refresh", authHandler.RefreshTokenHandler)
		authGroup.POST("/logout", authHandler.LogoutHandler)
		authGroup.POST("/verify-email", authHandler.VerifyEmailHandler)
		authGroup.GET("/verify", authHandler.VerifyEmailGetHandler)
		authGroup.POST("/forgot-password", authHandler.ForgotPasswordHandler)
		authGroup.POST("/reset-password", authHandler.ResetPasswordHandler)
	}

	// ----- Public user routes -----
	usersPublic := api.Group("/users")
	{
		usersPublic.GET("/check-email", authHandler.CheckEmailHandler)
		usersPublic.GET("/:id", authHandler.GetUserByIDHandler)
		usersPublic.GET("/:id/summary", authHandler.GetUserSummaryHandler)
		usersPublic.GET("/search", authHandler.SearchUsersHandler)
	}

	// ----- Metadata routes -----
	api.GET("/skills", authHandler.SearchSkillsHandler)
	api.GET("/courses", authHandler.SearchCoursesHandler)
	api.GET("/tags", authHandler.SearchTagsHandler)

	// ----- Protected user routes -----
	protectedUsers := api.Group("/users")
	protectedUsers.Use(middleware.RequireGatewayAuth())
	{
		protectedUsers.GET("/me", authHandler.GetCurrentUserHandler)
		protectedUsers.PATCH("/me", authHandler.UpdateCurrentUserHandler)
		protectedUsers.POST("/me/avatar", authHandler.UploadAvatarHandler)
		protectedUsers.POST("/me/banner", authHandler.UploadProfileBannerHandler)
		protectedUsers.PATCH("/me/password", authHandler.ChangePasswordHandler)
		protectedUsers.GET("/me/saved", authHandler.GetSavedItemsHandler)
		protectedUsers.POST("/me/saved", authHandler.SaveItemHandler)
		protectedUsers.DELETE("/me/saved", authHandler.UnsaveItemHandler)
		protectedUsers.DELETE("/me/saved/:id", authHandler.UnsaveItemHandler)
		protectedUsers.POST("/me/skills", authHandler.AddUserSkillHandler)
		protectedUsers.DELETE("/me/skills/:name", authHandler.RemoveUserSkillHandler)
		protectedUsers.POST("/me/courses", authHandler.AddUserCourseHandler)
		protectedUsers.DELETE("/me/courses/:code", authHandler.RemoveUserCourseHandler)
		protectedUsers.GET("/me/services", authHandler.GetUserServicesHandler)
		protectedUsers.POST("/me/services", authHandler.CreateUserServiceHandler)
		protectedUsers.PATCH("/me/services/:id", authHandler.UpdateUserServiceHandler)
		protectedUsers.DELETE("/me/services/:id", authHandler.DeleteUserServiceHandler)
	}

	// ----- Portfolio routes -----
	portfolio := api.Group("/portfolio")
	{
		portfolio.GET("", portfolioHandler.GetPublicProjectsHandler)
		portfolio.GET("/:id", portfolioHandler.GetProjectDetailHandler)
	}
	api.Group("/users").GET("/:id/portfolio", portfolioHandler.GetUserPortfolioHandler)
	protectedPortfolio := api.Group("/users")
	protectedPortfolio.Use(middleware.RequireGatewayAuth())
	{
		protectedPortfolio.GET("/me/portfolio", portfolioHandler.GetMyPortfolioHandler)
		protectedPortfolio.POST("/me/portfolio", portfolioHandler.CreateProjectHandler)
		protectedPortfolio.PATCH("/me/portfolio/:id", portfolioHandler.UpdateProjectHandler)
		protectedPortfolio.DELETE("/me/portfolio/:id", portfolioHandler.DeleteProjectHandler)
	}

	// ----- Opportunity routes -----
	opportunities := api.Group("/opportunities")
	{
		opportunities.GET("", oppHandler.ListOpportunitiesHandler)
		opportunities.GET("/:id", oppHandler.GetOpportunityHandler)
	}
	protectedOpp := api.Group("/opportunities")
	protectedOpp.Use(middleware.RequireGatewayAuth())
	{
		protectedOpp.POST("", oppHandler.CreateOpportunityHandler)
		protectedOpp.PATCH("/:id", oppHandler.UpdateOpportunityHandler)
		protectedOpp.DELETE("/:id", oppHandler.DeleteOpportunityHandler)
		protectedOpp.GET("/me", oppHandler.ListMyOpportunitiesHandler)
		protectedOpp.PATCH("/:id/status", oppHandler.UpdateOpportunityStatusHandler)
		protectedOpp.GET("/:id/stats", oppHandler.GetOpportunityStatsHandler)
		protectedOpp.POST("/:id/favorite", oppHandler.SaveOpportunityHandler)
		protectedOpp.DELETE("/:id/favorite", oppHandler.UnsaveOpportunityHandler)
		protectedOpp.POST("/:id/apply", oppHandler.CreateApplicationHandler)
		protectedOpp.GET("/applications", oppHandler.ListMyApplicationsHandler)
		protectedOpp.GET("/:id/applications", oppHandler.ListOpportunityApplicationsHandler)
	}
	protectedUsersOpp := api.Group("/users")
	protectedUsersOpp.Use(middleware.RequireGatewayAuth())
	{
		protectedUsersOpp.GET("/me/applications", oppHandler.ListMyApplicationsHandler)
	}

	// ----- Notification routes -----
	internalNotif := api.Group("/internal/notifications")
	internalNotif.Use(middleware.RequireInternalCall())
	{
		internalNotif.POST("", notificationHandler.CreateNotificationHandler)
	}
	userNotifications := api.Group("/notifications")
	userNotifications.Use(middleware.TrustGatewayMiddleware())
	userNotifications.Use(middleware.RequireGatewayAuth())
	{
		userNotifications.GET("", notificationHandler.GetNotificationsHandler)
		userNotifications.GET("/unread-count", notificationHandler.GetUnreadCountHandler)
		userNotifications.PUT("/:id/read", notificationHandler.MarkNotificationAsReadHandler)
		userNotifications.PATCH("/:id/read", notificationHandler.MarkNotificationAsReadHandler)
		userNotifications.PUT("/read-all", notificationHandler.MarkAllNotificationsAsReadHandler)
		userNotifications.PATCH("/read-all", notificationHandler.MarkAllNotificationsAsReadHandler)
		userNotifications.DELETE("/:id", notificationHandler.DeleteNotificationHandler)
		userNotifications.DELETE("", notificationHandler.DeleteMultipleNotificationsHandler)
	}

	// Start HTTP server
	if err := app.RunServer(cfg.App.HTTPPort, router, logger); err != nil {
		logger.Fatal("Server error", zap.Error(err))
	}
}
