package model

import "time"

// ChatMessage represents a single chat message
type ChatMessage struct {
	ID             int64     `db:"id" json:"id"`
	ConversationID string    `db:"conversation_id" json:"conversation_id"`
	FromUserID     int64     `db:"from_user_id" json:"from_user_id"`
	Content        string    `db:"content" json:"content"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

// CreateChatMessageInput for inserting a new message (internal API)
type CreateChatMessageInput struct {
	ConversationID string `json:"conversation_id"`
	FromUserID     int64  `json:"from_user_id"`
	Content        string `json:"content"`
	CreatedAt      string `json:"created_at"` // ISO8601, optional
}

// ConversationSummary for listing user's conversations (last message + other user info)
type ConversationSummary struct {
	ConversationID string    `json:"conversation_id"`
	OtherUserID    int64     `json:"other_user_id"`
	OtherUserName  string    `json:"other_user_name,omitempty"`
	OtherUserAvatar string   `json:"other_user_avatar,omitempty"`
	OrganizationName string  `json:"organization_name,omitempty"`
	LastMessage    string    `json:"last_message"`
	LastAt         time.Time `json:"last_at"`
	UnreadCount    int       `json:"unread_count"`
}
