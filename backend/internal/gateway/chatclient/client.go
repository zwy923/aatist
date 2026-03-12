package chatclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aatist/backend/internal/gateway/websocket"
	"github.com/aatist/backend/internal/platform/log"
	"go.uber.org/zap"
)

const defaultChatServiceURL = "http://localhost:8088"

// NewPersistFunc returns a PersistMessageFunc that POSTs to chat-service internal API.
// If CHAT_SERVICE_URL is empty or the request fails, it returns an error (caller may ignore).
func NewPersistFunc(logger *log.Logger) websocket.PersistMessageFunc {
	baseURL := os.Getenv("CHAT_SERVICE_URL")
	if baseURL == "" {
		baseURL = defaultChatServiceURL
	}
	url := baseURL + "/api/v1/internal/messages"
	client := &http.Client{Timeout: 5 * time.Second}
	return func(conversationID, fromUserID, content, createdAt string) error {
		fromID, err := strconv.ParseInt(fromUserID, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid from_user_id: %w", err)
		}
		body := map[string]interface{}{
			"conversation_id": conversationID,
			"from_user_id":   fromID,
			"content":        content,
			"created_at":     createdAt,
		}
		raw, _ := json.Marshal(body)
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(raw))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Internal-Call", "true")
		if token := os.Getenv("INTERNAL_API_TOKEN"); token != "" {
			req.Header.Set("X-Internal-Token", token)
		}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			logger.Warn("chat-service persist returned non-2xx",
				zap.Int("status", resp.StatusCode),
				zap.String("conversation_id", conversationID))
			return fmt.Errorf("chat-service returned %d", resp.StatusCode)
		}
		return nil
	}
}
