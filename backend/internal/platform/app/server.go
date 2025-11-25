package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aalto-talent-network/backend/internal/platform/log"
	"go.uber.org/zap"
)

// RunServer starts an HTTP server with graceful shutdown
func RunServer(port int, handler http.Handler, logger *log.Logger) error {
	addr := fmt.Sprintf(":%d", port)
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
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
		return err
	}

	logger.Info("Server exited")
	return nil
}
