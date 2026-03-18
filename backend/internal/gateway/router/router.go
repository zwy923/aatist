package router

import (
	"time"

	"github.com/aatist/backend/internal/gateway/chatclient"
	gatewayMiddleware "github.com/aatist/backend/internal/gateway/middleware"
	"github.com/aatist/backend/internal/gateway/proxy"
	"github.com/aatist/backend/internal/gateway/websocket"
	"github.com/aatist/backend/internal/platform/auth"
	"github.com/aatist/backend/internal/platform/cache"
	"github.com/aatist/backend/internal/platform/config"
	"github.com/aatist/backend/internal/platform/log"
	"github.com/aatist/backend/internal/platform/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all routes for the gateway
func RegisterRoutes(router *gin.Engine, cfg *config.Config, logger *log.Logger, jwt *auth.JWT, redis *cache.Redis) {
	// Helper function to get service timeout
	getServiceTimeout := func(serviceName string) time.Duration {
		if timeout, ok := cfg.Gateway.ServiceTimeouts[serviceName]; ok {
			return timeout
		}
		return cfg.Gateway.ServiceTimeout
	}

	// Persist chat messages to chat-service (optional; set CHAT_SERVICE_URL if chat-service is running)
	persistFunc := chatclient.NewPersistFunc(logger)
	wsManager := websocket.NewManager(redis, logger, persistFunc)

	// API routes with /api/v1 prefix
	api := router.Group("/api/v1")

	// ============================================
	// PROTECTED ROUTES (require auth)
	// ============================================
	protected := api.Group("")
	protected.Use(middleware.GatewayAuthMiddleware(jwt))

	// WebSocket route
	protected.GET("/ws", wsManager.HandleWebSocket)

	registerProtectedRoutes(protected, getServiceTimeout, logger)

	// ============================================
	// PUBLIC ROUTES (no auth required)
	// ============================================
	public := api.Group("")

	// Auth routes (proxy to backend monolith)
	authHandler := proxy.NewHandler("backend", 8081, getServiceTimeout("backend"), logger)
	public.GET("/auth/*path", authHandler)
	public.POST("/auth/*path", authHandler)

	registerPublicRoutes(public, getServiceTimeout, logger)

	// ============================================
	// INTERNAL API ROUTES (service-to-service only)
	// ============================================
	internal := api.Group("/internal")
	internal.Use(middleware.ValidateInternalCaller())  // Require X-Internal-Token when INTERNAL_API_TOKEN is set
	internal.Use(middleware.InternalServiceMiddleware()) // Inject headers for downstream

	registerInternalRoutes(internal, getServiceTimeout, logger)
}

// Helper structs for route registration
type RouteDef struct {
	Method string
	Path   string
}

type ServiceRoutes struct {
	Name   string
	Port   int
	Routes []RouteDef
}

func registerProtectedRoutes(group *gin.RouterGroup, getTimeout func(string) time.Duration, logger *log.Logger) {
	services := []ServiceRoutes{
		// Backend monolith: user, portfolio, opportunity, notification
		{
			Name: "backend", Port: 8081,
			Routes: []RouteDef{
				{"GET", "/notifications"},
				{"GET", "/notifications/:id"},
				{"PATCH", "/notifications/read-all"},
				{"PATCH", "/notifications/:id/read"},
				{"DELETE", "/notifications/:id"},
				{"GET", "/users/search"},
				{"GET", "/users/me"},
				{"PATCH", "/users/me"},
				{"POST", "/users/me/avatar"},
				{"PATCH", "/users/me/password"},
				{"GET", "/users/me/saved"},
				{"POST", "/users/me/saved"},
				{"DELETE", "/users/me/saved"},
				{"DELETE", "/users/me/saved/:id"},
				{"POST", "/users/me/skills"},
				{"DELETE", "/users/me/skills/:name"},
				{"POST", "/users/me/courses"},
				{"DELETE", "/users/me/courses/:code"},
				{"GET", "/users/me/services"},
				{"POST", "/users/me/services"},
				{"PATCH", "/users/me/services/:id"},
				{"DELETE", "/users/me/services/:id"},
				{"GET", "/users/me/portfolio"},
				{"POST", "/users/me/portfolio"},
				{"PATCH", "/users/me/portfolio/:id"},
				{"DELETE", "/users/me/portfolio/:id"},
				{"GET", "/opportunities"},
				{"GET", "/opportunities/:id"},
				{"POST", "/opportunities"},
				{"PATCH", "/opportunities/:id"},
				{"DELETE", "/opportunities/:id"},
				{"POST", "/opportunities/:id/favorite"},
				{"DELETE", "/opportunities/:id/favorite"},
				{"POST", "/opportunities/:id/apply"},
				{"GET", "/opportunities/me"},
				{"PATCH", "/opportunities/:id/status"},
				{"GET", "/opportunities/:id/stats"},
				{"GET", "/opportunities/:id/applications"},
				{"GET", "/applications/:id"},
				{"PATCH", "/applications/:id"},
				{"DELETE", "/applications/:id"},
				{"GET", "/users/me/applications"},
			},
		},
		{
			Name: "file-service", Port: 8086,
			Routes: []RouteDef{
				{"GET", "/files"},
				{"GET", "/files/:id"},
				{"GET", "/files/:id/download"},
				{"POST", "/files"},
				{"POST", "/files/upload"},
				{"POST", "/files/presigned-upload"},
				{"POST", "/files/confirm-upload"},
				{"DELETE", "/files/:id"},
			},
		},
		{
			Name: "chat-service", Port: 8088,
			Routes: []RouteDef{
				{"POST", "/conversations/start"},
				{"GET", "/conversations"},
				{"GET", "/conversations/:id/messages"},
				{"PUT", "/conversations/:id/read"},
				{"PATCH", "/conversations/:id/read"},
				{"DELETE", "/conversations/:id"},
			},
		},
	}

	for _, svc := range services {
		handler := proxy.NewHandler(svc.Name, svc.Port, getTimeout(svc.Name), logger)
		for _, route := range svc.Routes {
			group.Handle(route.Method, route.Path, handler)
		}
	}
}

func registerPublicRoutes(group *gin.RouterGroup, getTimeout func(string) time.Duration, logger *log.Logger) {
	services := []ServiceRoutes{
		{
			Name: "backend", Port: 8081,
			Routes: []RouteDef{
				{"GET", "/users/check-username"},
				{"GET", "/users/check-email"},
				{"GET", "/users/:id"},
				{"GET", "/users/:id/summary"},
				{"GET", "/stats/overview"},
				{"GET", "/skills/popular"},
				{"GET", "/skills"},
				{"GET", "/courses"},
				{"GET", "/tags"},
				{"GET", "/portfolio"},
				{"GET", "/portfolio/:id"},
				{"GET", "/users/:id/portfolio"},
			},
		},
	}

	for _, svc := range services {
		handler := proxy.NewHandler(svc.Name, svc.Port, getTimeout(svc.Name), logger)
		for _, route := range svc.Routes {
			group.Handle(route.Method, route.Path, handler)
		}
	}
}

func registerInternalRoutes(group *gin.RouterGroup, getTimeout func(string) time.Duration, logger *log.Logger) {
	type InternalConfig struct {
		GroupPath   string
		ServiceName string
		Port        int
		RewriteFrom string
		RewriteTo   string
		NoRewrite   bool
	}

	configs := []InternalConfig{
		{
			GroupPath: "/user", ServiceName: "backend", Port: 8081,
			RewriteFrom: "/api/v1/internal/user", RewriteTo: "/api/v1",
		},
		{
			GroupPath: "/notification", ServiceName: "backend", Port: 8081,
			RewriteFrom: "/api/v1/internal/notification", RewriteTo: "/api/v1",
		},
		{
			GroupPath: "/portfolio", ServiceName: "backend", Port: 8081,
			RewriteFrom: "/api/v1/internal/portfolio", RewriteTo: "/api/v1/portfolio",
		},
		{
			GroupPath: "/file", ServiceName: "file-service", Port: 8086,
			NoRewrite: true,
		},
		{
			GroupPath: "/chat", ServiceName: "chat-service", Port: 8088,
			RewriteFrom: "/api/v1/internal/chat", RewriteTo: "/api/v1/internal",
		},
	}

	for _, cfg := range configs {
		subGroup := group.Group(cfg.GroupPath)
		if !cfg.NoRewrite {
			subGroup.Use(gatewayMiddleware.InternalRewrite(cfg.RewriteFrom, cfg.RewriteTo))
		}

		handler := proxy.NewHandler(cfg.ServiceName, cfg.Port, getTimeout(cfg.ServiceName), logger)
		subGroup.Any("/*path", handler)
	}
}
