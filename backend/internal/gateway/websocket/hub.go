package websocket

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aatist/backend/internal/platform/log"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	writeWait           = 10 * time.Second
	pongWait            = 60 * time.Second
	pingPeriod          = (pongWait * 9) / 10
	maxMessageSize      = 512 * 1024
	maxContentLength   = 65536 // text + JSON file attachments (URL + metadata)
)

type connection struct {
	userID string
	conn   *websocket.Conn
	send   chan []byte
	hub    *Hub
	logger *log.Logger
}

func (c *connection) readPump() {
	defer func() {
		c.hub.Unregister(c.userID)
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		var msg ClientMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			c.sendServerMessage(ServerMessage{Type: "error", Error: "invalid json"})
			continue
		}
		switch msg.Type {
		case "ping":
			c.sendServerMessage(ServerMessage{Type: "pong"})
		case "message":
			if msg.ConversationID == "" || msg.Content == "" {
				c.sendServerMessage(ServerMessage{Type: "error", Error: "conversation_id and content required"})
				continue
			}
			if len(msg.Content) > maxContentLength {
				c.sendServerMessage(ServerMessage{Type: "error", Error: "content too long"})
				continue
			}
			otherID := c.hub.otherUserInConversation(msg.ConversationID, c.userID)
			if otherID == "" {
				c.sendServerMessage(ServerMessage{Type: "error", Error: "invalid conversation_id"})
				continue
			}
			createdAt := time.Now().UTC().Format(time.RFC3339)
			payload := ServerMessage{
				Type:           "message",
				ID:             uuid.New().String(),
				ConversationID: msg.ConversationID,
				FromUserID:     c.userID,
				Content:        msg.Content,
				CreatedAt:      createdAt,
				TempID:         msg.TempID,
			}
			// 发给对方
			c.hub.SendToUser(otherID, payload)
			// 回显给发送者（message_sent 便于前端区分“自己发的”）
			payload.Type = "message_sent"
			c.sendServerMessage(payload)
			// 持久化到 chat-service（可选，失败仅打日志）
			if c.hub.persist != nil {
				go func() {
					if err := c.hub.persist(msg.ConversationID, c.userID, msg.Content, createdAt); err != nil {
						c.logger.Warn("persist message failed", zap.String("conversation_id", msg.ConversationID), zap.Error(err))
					}
				}()
			}
		case "typing":
			if msg.ConversationID == "" {
				c.sendServerMessage(ServerMessage{Type: "error", Error: "conversation_id required"})
				continue
			}
			otherID := c.hub.otherUserInConversation(msg.ConversationID, c.userID)
			if otherID == "" {
				continue
			}
			isTyping := msg.IsTyping != nil && *msg.IsTyping
			c.hub.SendToUser(otherID, ServerMessage{
				Type:           "typing",
				ConversationID: msg.ConversationID,
				FromUserID:     c.userID,
				IsTyping:       isTyping,
			})
		default:
			c.sendServerMessage(ServerMessage{Type: "error", Error: "unknown type"})
		}
	}
}

func (c *connection) sendServerMessage(msg ServerMessage) {
	raw, _ := json.Marshal(msg)
	select {
	case c.send <- raw:
	default:
		c.logger.Warn("client send buffer full, dropping", zap.String("userID", c.userID))
	}
}

func (c *connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case data, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// PersistMessageFunc 持久化一条消息到 chat-service（由网关注入，可选）
type PersistMessageFunc func(conversationID, fromUserID, content, createdAt string) error

// Hub 维护 userID -> connection，并支持按会话转发消息
type Hub struct {
	mu      sync.RWMutex
	conns   map[string]*connection
	logger  *log.Logger
	online  map[string]struct{}
	persist PersistMessageFunc
}

// NewHub 创建 Hub，persist 可为 nil 表示不持久化
func NewHub(logger *log.Logger, persist PersistMessageFunc) *Hub {
	return &Hub{
		conns:   make(map[string]*connection),
		logger:  logger,
		online:  make(map[string]struct{}),
		persist: persist,
	}
}

// conversationID 生成规范会话 ID：小ID_大ID
func conversationID(a, b int64) string {
	if a > b {
		a, b = b, a
	}
	return strconv.FormatInt(a, 10) + "_" + strconv.FormatInt(b, 10)
}

// OtherUserInConversation 从 conversation_id 中解析出“对方” userID（当前用户为 fromUserID）
func (h *Hub) otherUserInConversation(conversationID, fromUserID string) string {
	parts := strings.SplitN(conversationID, "_", 2)
	if len(parts) != 2 {
		return ""
	}
	if parts[0] == fromUserID {
		return parts[1]
	}
	if parts[1] == fromUserID {
		return parts[0]
	}
	return ""
}

// ConversationID 根据两个 userID 生成 conversation_id（与前端约定一致）
func ConversationID(userID1, userID2 int64) string {
	if userID1 > userID2 {
		userID1, userID2 = userID2, userID1
	}
	return strconv.FormatInt(userID1, 10) + "_" + strconv.FormatInt(userID2, 10)
}

// Register 注册连接；同一 userID 新连接会踢掉旧连接
func (h *Hub) Register(userID string, conn *websocket.Conn, logger *log.Logger) {
	h.mu.Lock()
	if old, ok := h.conns[userID]; ok {
		close(old.send)
		delete(h.conns, userID)
	}
	c := &connection{userID: userID, conn: conn, send: make(chan []byte, 256), hub: h, logger: logger}
	h.conns[userID] = c
	h.online[userID] = struct{}{}
	onlineList := h.onlineUserIDsLocked()
	// copy current connections for broadcasting after releasing lock
	targets := make([]*connection, 0, len(h.conns))
	for _, conn := range h.conns {
		targets = append(targets, conn)
	}
	h.mu.Unlock()

	// broadcast updated online list to all clients
	msg := ServerMessage{Type: "online", OnlineUserIDs: onlineList}
	for _, conn := range targets {
		conn.sendServerMessage(msg)
	}

	go c.writePump()
	c.readPump()
}

func (h *Hub) onlineUserIDsLocked() []string {
	ids := make([]string, 0, len(h.online))
	for id := range h.online {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// Unregister 移除连接
func (h *Hub) Unregister(userID string) {
	h.mu.Lock()
	if c, ok := h.conns[userID]; ok {
		close(c.send)
		delete(h.conns, userID)
		delete(h.online, userID)
	}
	onlineList := h.onlineUserIDsLocked()
	// copy current connections for broadcasting after releasing lock
	targets := make([]*connection, 0, len(h.conns))
	for _, conn := range h.conns {
		targets = append(targets, conn)
	}
	h.mu.Unlock()

	// broadcast updated online list
	if len(targets) > 0 {
		msg := ServerMessage{Type: "online", OnlineUserIDs: onlineList}
		for _, conn := range targets {
			conn.sendServerMessage(msg)
		}
	}
}

// SendToUser 向指定用户发送一条服务端消息
func (h *Hub) SendToUser(userID string, msg ServerMessage) {
	h.mu.RLock()
	c, ok := h.conns[userID]
	h.mu.RUnlock()
	if !ok {
		return
	}
	raw, _ := json.Marshal(msg)
	select {
	case c.send <- raw:
	default:
		h.logger.Warn("send buffer full", zap.String("userID", userID))
	}
}

// OnlineUserIDs 返回当前在线用户 ID 列表（用于前端显示在线状态）
func (h *Hub) OnlineUserIDs() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.onlineUserIDsLocked()
}
