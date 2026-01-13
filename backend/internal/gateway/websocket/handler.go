package websocket

import (
	"net/http"
	"time"

	"github.com/aatist/backend/internal/platform/auth"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for now, should be configured based on environment
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandleWebSocket handles WebSocket connections
func (m *Manager) HandleWebSocket(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userClaims, ok := claims.(*auth.Claims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
		return
	}

	userID := userClaims.Subject // Assuming Subject holds the User ID

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		m.logger.Error("Failed to upgrade to websocket", zap.Error(err))
		return
	}
	defer conn.Close()

	m.logger.Info("WebSocket connected", zap.String("userID", userID))

	// Set initial online status
	if err := m.SetUserOnline(c.Request.Context(), userID); err != nil {
		// If we can't set status, maybe we should close connection?
		// For now, just log error.
	}

	// Clean up on exit
	defer func() {
		m.RemoveUserOnline(c.Request.Context(), userID)
		m.logger.Info("WebSocket disconnected", zap.String("userID", userID))
	}()

	// Heartbeat loop
	// We expect the client to send a ping or any message periodically.
	// Or we can just rely on the connection being open and update Redis periodically.
	// Let's update Redis every 30 seconds as long as connection is open.

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Handle close messages
	conn.SetCloseHandler(func(code int, text string) error {
		return nil
	})

	// Read loop to detect client disconnects
	go func() {
		for {
			if _, _, err := conn.NextReader(); err != nil {
				conn.Close()
				break
			}
		}
	}()

	for {
		select {
		case <-ticker.C:
			if err := m.SetUserOnline(c.Request.Context(), userID); err != nil {
				m.logger.Error("Failed to refresh online status", zap.String("userID", userID), zap.Error(err))
			}
		case <-c.Request.Context().Done():
			return
		}
	}
}
