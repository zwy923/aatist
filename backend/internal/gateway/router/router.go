package router

import (
	"time"

	gatewayMiddleware "github.com/aatist/backend/internal/gateway/middleware"
	"github.com/aatist/backend/internal/gateway/proxy"
	"github.com/aatist/backend/internal/platform/auth"
	"github.com/aatist/backend/internal/platform/config"
	"github.com/aatist/backend/internal/platform/log"
	"github.com/aatist/backend/internal/platform/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all routes for the gateway
func RegisterRoutes(router *gin.Engine, cfg *config.Config, logger *log.Logger, jwt *auth.JWT) {
	// Helper function to get service timeout
	getServiceTimeout := func(serviceName string) time.Duration {
		if timeout, ok := cfg.Gateway.ServiceTimeouts[serviceName]; ok {
			return timeout
		}
		return cfg.Gateway.ServiceTimeout
	}

	// API routes with /api/v1 prefix
	api := router.Group("/api/v1")

	// ============================================
	// PROTECTED ROUTES (require auth)
	// ============================================
	protected := api.Group("")
	protected.Use(middleware.GatewayAuthMiddleware(jwt))

	registerProtectedRoutes(protected, getServiceTimeout, logger)

	// ============================================
	// PUBLIC ROUTES (no auth required)
	// ============================================
	public := api.Group("")

	// Auth routes (proxy to user-service)
	authHandler := proxy.NewHandler("user-service", 8081, getServiceTimeout("user-service"), logger)
	public.GET("/auth/*path", authHandler)
	public.POST("/auth/*path", authHandler)

	registerPublicRoutes(public, getServiceTimeout, logger)

	// ============================================
	// INTERNAL API ROUTES
	// ============================================
	internal := api.Group("/internal")
	internal.Use(middleware.InternalServiceMiddleware())

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
		{
			Name: "notification-service", Port: 8085,
			Routes: []RouteDef{
				{"GET", "/notifications"},
				{"GET", "/notifications/*path"},
				{"POST", "/notifications"},
				{"POST", "/notifications/*path"},
				{"PUT", "/notifications/*path"},
				{"PATCH", "/notifications/*path"},
				{"DELETE", "/notifications/*path"},
			},
		},
		{
			Name: "user-service", Port: 8081,
			Routes: []RouteDef{
				{"GET", "/users/me"},
				{"PATCH", "/users/me"},
				{"POST", "/users/me/avatar"},
				{"PATCH", "/users/me/password"},
				{"GET", "/users/me/availability"},
				{"PATCH", "/users/me/availability"},
				{"GET", "/users/me/saved"},
				{"POST", "/users/me/saved"},
				{"DELETE", "/users/me/saved"},
			},
		},
		{
			Name: "portfolio-service", Port: 8082,
			Routes: []RouteDef{
				{"GET", "/users/me/portfolio"},
				{"POST", "/users/me/portfolio"},
				{"PUT", "/users/me/portfolio/:id"},
				{"DELETE", "/users/me/portfolio/:id"},
			},
		},
		{
			Name: "file-service", Port: 8086,
			Routes: []RouteDef{
				{"GET", "/files"},
				{"GET", "/files/*path"},
				{"POST", "/files"},
				{"POST", "/files/*path"},
				{"DELETE", "/files/*path"},
			},
		},
		{
			Name: "opp-service", Port: 8083,
			Routes: []RouteDef{
				{"GET", "/opportunities"},
				{"GET", "/opportunities/*path"},
				{"POST", "/opportunities"},
				{"POST", "/opportunities/*path"},
				{"PUT", "/opportunities/*path"},
				{"PATCH", "/opportunities/*path"},
				{"DELETE", "/opportunities/*path"},
			},
		},
		{
			Name: "events-service", Port: 8084,
			Routes: []RouteDef{
				{"POST", "/events"},
				{"PATCH", "/events/:id"},
				{"DELETE", "/events/:id"},
				{"POST", "/events/:id/interested"},
				{"DELETE", "/events/:id/interested"},
				{"POST", "/events/:id/going"},
				{"DELETE", "/events/:id/going"},
				{"POST", "/events/:id/comments"},
				{"PATCH", "/events/comments/:id"},
				{"DELETE", "/events/comments/:id"},
			},
		},
		{
			Name: "community-service", Port: 8087,
			Routes: []RouteDef{
				{"POST", "/community/posts"},
				{"PUT", "/community/posts/:id"},
				{"DELETE", "/community/posts/:id"},
				{"POST", "/community/posts/:id/like"},
				{"DELETE", "/community/posts/:id/like"},
				{"POST", "/community/posts/:id/comments"},
				{"PUT", "/community/comments/:id"},
				{"DELETE", "/community/comments/:id"},
				{"GET", "/community/users/me/posts"},
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
			Name: "user-service", Port: 8081,
			Routes: []RouteDef{
				{"GET", "/users/check-username"},
				{"GET", "/users/check-email"},
				{"GET", "/users/:id"},
				{"GET", "/users/:id/summary"},
			},
		},
		{
			Name: "portfolio-service", Port: 8082,
			Routes: []RouteDef{
				{"GET", "/portfolio/:id"},
				{"GET", "/users/:id/portfolio"},
			},
		},
		{
			Name: "community-service", Port: 8087,
			Routes: []RouteDef{
				{"GET", "/community/posts"},
				{"GET", "/community/posts/trending"},
				{"GET", "/community/posts/:id"},
				{"GET", "/community/posts/:id/comments"},
				{"GET", "/community/users/:id/posts"},
			},
		},
		{
			Name: "events-service", Port: 8084,
			Routes: []RouteDef{
				{"GET", "/events"},
				{"GET", "/events/:id"},
				{"GET", "/events/:id/comments"},
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
			GroupPath: "/user", ServiceName: "user-service", Port: 8081,
			RewriteFrom: "/api/v1/internal/user", RewriteTo: "/api/v1",
		},
		{
			GroupPath: "/notification", ServiceName: "notification-service", Port: 8085,
			RewriteFrom: "/api/v1/internal/notification", RewriteTo: "/api/v1",
		},
		{
			GroupPath: "/portfolio", ServiceName: "portfolio-service", Port: 8082,
			RewriteFrom: "/api/v1/internal/portfolio", RewriteTo: "/api/v1/portfolio",
		},
		{
			GroupPath: "/file", ServiceName: "file-service", Port: 8086,
			NoRewrite: true,
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
