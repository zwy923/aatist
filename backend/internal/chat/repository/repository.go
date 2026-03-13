package repository

import (
	"context"

	"github.com/aatist/backend/internal/chat/model"
)

type ChatRepository interface {
	Create(ctx context.Context, msg *model.ChatMessage) error
	ListByConversation(ctx context.Context, conversationID string, limit, offset int) ([]*model.ChatMessage, error)
	ListConversationsForUser(ctx context.Context, userID int64, limit int) ([]*model.ConversationSummary, error)
	DeleteConversation(ctx context.Context, conversationID string) error
}
