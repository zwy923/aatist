package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aatist/backend/internal/chat/model"
	"github.com/aatist/backend/internal/chat/repository"
	"github.com/aatist/backend/pkg/errs"
)

func participantsFromConversationID(conversationID string) (int64, int64, error) {
	parts := strings.SplitN(conversationID, "_", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid conversation id")
	}
	a, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || a <= 0 {
		return 0, 0, fmt.Errorf("invalid conversation participant")
	}
	b, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || b <= 0 {
		return 0, 0, fmt.Errorf("invalid conversation participant")
	}
	if a == b {
		return 0, 0, fmt.Errorf("invalid conversation participants")
	}
	if a > b {
		return 0, 0, fmt.Errorf("conversation id must be ordered")
	}
	return a, b, nil
}

func participantInConversation(conversationID string, userID int64) bool {
	a, b, err := participantsFromConversationID(conversationID)
	if err != nil {
		return false
	}
	return a == userID || b == userID
}

type ChatService interface {
	CreateMessage(ctx context.Context, conversationID string, fromUserID int64, content, createdAt string) (*model.ChatMessage, error)
	ListConversations(ctx context.Context, userID int64, limit int) ([]*model.ConversationSummary, error)
	ListMessages(ctx context.Context, conversationID string, userID int64, limit, offset int) ([]*model.ChatMessage, error)
	DeleteConversation(ctx context.Context, conversationID string, userID int64) error
}

type chatService struct {
	repo repository.ChatRepository
}

func NewChatService(repo repository.ChatRepository) ChatService {
	return &chatService{repo: repo}
}

func (s *chatService) CreateMessage(ctx context.Context, conversationID string, fromUserID int64, content, createdAt string) (*model.ChatMessage, error) {
	conversationID = strings.TrimSpace(conversationID)
	content = strings.TrimSpace(content)
	if conversationID == "" || fromUserID <= 0 || content == "" {
		return nil, errs.NewAppError(errs.ErrInvalidInput, 400, "conversation_id, from_user_id and content are required")
	}
	if !participantInConversation(conversationID, fromUserID) {
		return nil, errs.NewAppError(errs.ErrUnauthorized, 403, "sender is not in conversation")
	}

	msg := &model.ChatMessage{
		ConversationID: conversationID,
		FromUserID:     fromUserID,
		Content:        content,
	}
	if createdAt != "" {
		if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
			msg.CreatedAt = t
		}
	}
	if err := s.repo.Create(ctx, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (s *chatService) ListConversations(ctx context.Context, userID int64, limit int) ([]*model.ConversationSummary, error) {
	if userID <= 0 {
		return nil, errs.NewAppError(errs.ErrInvalidInput, 400, "invalid user id")
	}
	return s.repo.ListConversationsForUser(ctx, userID, limit)
}

func (s *chatService) ListMessages(ctx context.Context, conversationID string, userID int64, limit, offset int) ([]*model.ChatMessage, error) {
	conversationID = strings.TrimSpace(conversationID)
	if conversationID == "" || userID <= 0 {
		return nil, errs.NewAppError(errs.ErrInvalidInput, 400, "invalid conversation id or user id")
	}
	if !participantInConversation(conversationID, userID) {
		return nil, errs.NewAppError(errs.ErrUnauthorized, 403, "conversation access denied")
	}
	return s.repo.ListByConversation(ctx, conversationID, limit, offset)
}

func (s *chatService) DeleteConversation(ctx context.Context, conversationID string, userID int64) error {
	conversationID = strings.TrimSpace(conversationID)
	if conversationID == "" || userID <= 0 {
		return errs.NewAppError(errs.ErrInvalidInput, 400, "invalid conversation id or user id")
	}
	if !participantInConversation(conversationID, userID) {
		return errs.NewAppError(errs.ErrUnauthorized, 403, "conversation access denied")
	}
	return s.repo.DeleteConversation(ctx, conversationID)
}
