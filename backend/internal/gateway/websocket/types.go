package websocket

// ClientMessage 客户端 -> 服务端
type ClientMessage struct {
	Type           string `json:"type"`                     // "message" | "ping" | "typing"
	ConversationID string `json:"conversation_id"`          // 会话 ID，格式 "userID1_userID2"（按数字升序）
	Content        string `json:"content,omitempty"`        // 消息正文
	TempID         string `json:"temp_id,omitempty"`        // 客户端临时 ID，用于确认送达
	IsTyping       *bool  `json:"is_typing,omitempty"`      // 输入状态，仅 type=typing 时使用
}

// ServerMessage 服务端 -> 客户端
type ServerMessage struct {
	Type           string   `json:"type"`                        // "message" | "message_sent" | "error" | "pong" | "online" | "typing"
	ID             string   `json:"id,omitempty"`                // 服务端消息 ID
	ConversationID string   `json:"conversation_id,omitempty"`
	FromUserID     string   `json:"from_user_id,omitempty"`
	Content        string   `json:"content,omitempty"`
	CreatedAt      string   `json:"created_at,omitempty"`        // ISO8601
	TempID         string   `json:"temp_id,omitempty"`           // 回显客户端 temp_id
	Error          string   `json:"error,omitempty"`
	OnlineUserIDs  []string `json:"online_user_ids,omitempty"`
	IsTyping       bool     `json:"is_typing,omitempty"`         // 输入状态，仅 type=typing 时使用
}
