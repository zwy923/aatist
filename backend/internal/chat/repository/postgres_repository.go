package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aatist/backend/internal/chat/model"
	"github.com/jmoiron/sqlx"
)

type postgresChatRepository struct {
	db *sqlx.DB
}

// NewPostgresChatRepository creates a new Postgres-backed chat repository
func NewPostgresChatRepository(db *sqlx.DB) ChatRepository {
	return &postgresChatRepository{db: db}
}

func (r *postgresChatRepository) Create(ctx context.Context, msg *model.ChatMessage) error {
	query := `INSERT INTO chat_messages (conversation_id, from_user_id, content, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`
	var err error
	if !msg.CreatedAt.IsZero() {
		err = r.db.QueryRowContext(ctx, query,
			msg.ConversationID, msg.FromUserID, msg.Content, msg.CreatedAt,
		).Scan(&msg.ID)
	} else {
		msg.CreatedAt = time.Now()
		err = r.db.QueryRowContext(ctx, query,
			msg.ConversationID, msg.FromUserID, msg.Content, msg.CreatedAt,
		).Scan(&msg.ID)
	}
	if err != nil {
		return fmt.Errorf("create chat message: %w", err)
	}
	return nil
}

func (r *postgresChatRepository) ListByConversation(ctx context.Context, conversationID string, limit, offset int) ([]*model.ChatMessage, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	query := `SELECT id, conversation_id, from_user_id, content, created_at
		FROM chat_messages
		WHERE conversation_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`
	var list []*model.ChatMessage
	if err := r.db.SelectContext(ctx, &list, query, conversationID, limit, offset); err != nil {
		return nil, fmt.Errorf("list by conversation: %w", err)
	}
	// Reverse so oldest first for display
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}
	return list, nil
}

// otherUserFromConversation parses "id1_id2" and returns the id that is not the given userID
func otherUserFromConversation(conversationID string, userID int64) (int64, bool) {
	parts := strings.SplitN(conversationID, "_", 2)
	if len(parts) != 2 {
		return 0, false
	}
	a, _ := strconv.ParseInt(parts[0], 10, 64)
	b, _ := strconv.ParseInt(parts[1], 10, 64)
	if a == userID {
		return b, true
	}
	if b == userID {
		return a, true
	}
	return 0, false
}

func (r *postgresChatRepository) ListConversationsForUser(ctx context.Context, userID int64, limit int) ([]*model.ConversationSummary, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	userPrefix := fmt.Sprintf("%d_%%", userID)
	userSuffix := fmt.Sprintf("%%_%d", userID)
	query := `
		WITH conv_last AS (
			SELECT DISTINCT ON (conversation_id) conversation_id, content, created_at
			FROM chat_messages
			WHERE conversation_id LIKE $1 OR conversation_id LIKE $2
			ORDER BY conversation_id, created_at DESC
		)
		SELECT conversation_id, content AS last_message, created_at AS last_at FROM conv_last
		ORDER BY last_at DESC LIMIT $3`
	var rows []struct {
		ConversationID string    `db:"conversation_id"`
		LastMessage    string    `db:"last_message"`
		LastAt         time.Time `db:"last_at"`
	}
	if err := r.db.SelectContext(ctx, &rows, query, userPrefix, userSuffix, limit); err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}
	out := make([]*model.ConversationSummary, 0, len(rows))
	for _, row := range rows {
		otherID, ok := otherUserFromConversation(row.ConversationID, userID)
		if !ok {
			continue
		}
		out = append(out, &model.ConversationSummary{
			ConversationID: row.ConversationID,
			OtherUserID:    otherID,
			LastMessage:    row.LastMessage,
			LastAt:         row.LastAt,
		})
	}
	return out, nil
}
