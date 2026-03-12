package websocket

import (
	"net/http"
	"strconv"

	"github.com/aatist/backend/internal/platform/auth"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandleWebSocket handles WebSocket connections (auth + chat hub)
func (m *Manager) HandleWebSocket(c *gin.Context) {
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
	userIDStr := strconv.FormatInt(userClaims.UserID, 10)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		m.logger.Error("Failed to upgrade to websocket", zap.Error(err))
		return
	}

	m.logger.Info("WebSocket connected", zap.String("userID", userIDStr))
	if err := m.SetUserOnline(c.Request.Context(), userIDStr); err != nil {
		// non-fatal
	}
	defer func() {
		m.RemoveUserOnline(c.Request.Context(), userIDStr)
		m.logger.Info("WebSocket disconnected", zap.String("userID", userIDStr))
	}()

	m.hub.Register(userIDStr, conn, m.logger)
}
