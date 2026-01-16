package app

import (
	"net/http"
	"os"
	"strings"

	"github.com/aatist/backend/internal/platform/log"
	"github.com/aatist/backend/internal/platform/middleware"
	"github.com/aatist/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// NewDefaultRouter creates a new Gin router with default middlewares
func NewDefaultRouter(logger *log.Logger, serviceName string) *gin.Engine {
	// Set Gin mode based on environment
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = os.Getenv("ENV")
	}
	if env == "production" || env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Apply global middlewares
	router.Use(middleware.LoggerMiddleware(logger))
	router.Use(middleware.RecoveryMiddleware(logger))
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.TrustGatewayMiddleware()) // Trust headers from Gateway

	// CORS configuration
	corsOrigins := getCORSOrigins()
	router.Use(middleware.CORSMiddleware(corsOrigins))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, response.Success(map[string]interface{}{"status": "ok", "service": serviceName}))
	})

	return router
}

// getCORSOrigins parses CORS origins from environment variable
func getCORSOrigins() []string {
	env := os.Getenv("CORS_ORIGINS")
	if env == "" {
		return []string{"*"} // Default to allow all in development
	}
	return strings.Split(env, ",")
}
