package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/aalto-talent-network/backend/internal/notification/model"
	"github.com/aalto-talent-network/backend/internal/platform/log"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	communityEventPostLiked     = "community.post.liked"
	communityEventPostCommented = "community.post.commented"

	defaultLikeBatchWindow = 30 * time.Second
)

// CommunityEventConsumer processes community.* MQ events and creates notifications.
type CommunityEventConsumer struct {
	notificationSvc NotificationService
	logger          *log.Logger
	likeAggregator  *likeAggregationManager
}

func NewCommunityEventConsumer(notificationSvc NotificationService, redisClient redis.Cmdable, logger *log.Logger) *CommunityEventConsumer {
	return &CommunityEventConsumer{
		notificationSvc: notificationSvc,
		logger:          logger,
		likeAggregator:  newLikeAggregationManager(notificationSvc, redisClient, logger, defaultLikeBatchWindow),
	}
}

// HandleMessage routes community events by type.
func (c *CommunityEventConsumer) HandleMessage(eventType string, payload []byte) error {
	switch eventType {
	case communityEventPostLiked:
		var evt postLikedEvent
		if err := json.Unmarshal(payload, &evt); err != nil {
			return fmt.Errorf("failed to unmarshal post liked event: %w", err)
		}
		c.likeAggregator.Handle(evt)
	case communityEventPostCommented:
		var evt postCommentedEvent
		if err := json.Unmarshal(payload, &evt); err != nil {
			return fmt.Errorf("failed to unmarshal post commented event: %w", err)
		}
		if err := c.handlePostCommented(evt); err != nil {
			return err
		}
	default:
		c.logger.Debug("Ignoring unsupported community event", zap.String("event_type", eventType))
	}
	return nil
}

func (c *CommunityEventConsumer) handlePostCommented(evt postCommentedEvent) error {
	ctx := context.Background()
	// Notify post author.
	if evt.AuthorID != 0 && evt.AuthorID != evt.CommentAuthorID {
		title := "Your post has a new comment"
		message := fmt.Sprintf("New comment: %s", evt.CommentSnippet)
		data := model.NotificationData{
			"post_id":           evt.PostID,
			"comment_id":        evt.CommentID,
			"comment_author_id": evt.CommentAuthorID,
		}
		if err := c.notificationSvc.CreateNotification(ctx, evt.AuthorID, model.NotificationTypeCommunityComment, title, &message, data); err != nil {
			return err
		}
	}

	if len(evt.MentionedUserIDs) == 0 {
		return nil
	}

	title := "Someone mentioned you in a comment"
	message := fmt.Sprintf("Mentioned content: %s", evt.CommentSnippet)
	data := model.NotificationData{
		"post_id":           evt.PostID,
		"comment_id":        evt.CommentID,
		"comment_author_id": evt.CommentAuthorID,
	}
	for _, userID := range evt.MentionedUserIDs {
		if userID == evt.CommentAuthorID {
			continue
		}
		if err := c.notificationSvc.CreateNotification(ctx, userID, model.NotificationTypeCommunityMention, title, &message, data); err != nil {
			return err
		}
	}
	return nil
}

// likeAggregationManager batches like events before sending notifications.
type likeAggregationManager struct {
	notificationSvc NotificationService
	redis           redis.Cmdable
	logger          *log.Logger
	window          time.Duration

	mu     sync.Mutex
	timers map[int64]*time.Timer
}

func newLikeAggregationManager(notificationSvc NotificationService, redisClient redis.Cmdable, logger *log.Logger, window time.Duration) *likeAggregationManager {
	return &likeAggregationManager{
		notificationSvc: notificationSvc,
		redis:           redisClient,
		logger:          logger,
		window:          window,
		timers:          make(map[int64]*time.Timer),
	}
}

func (m *likeAggregationManager) Handle(evt postLikedEvent) {
	if evt.AuthorID == 0 || evt.PostID == 0 {
		return
	}
	if m.redis == nil {
		m.sendImmediate(evt.AuthorID, evt.PostID, 1)
		return
	}

	ctx := context.Background()
	key := likeBatchKey(evt.PostID)
	count, err := m.redis.Incr(ctx, key).Result()
	if err != nil {
		m.logger.Warn("failed to increment like batch count", zap.Error(err))
		m.sendImmediate(evt.AuthorID, evt.PostID, 1)
		return
	}
	if count == 1 {
		m.redis.PExpire(ctx, key, time.Hour)
		m.scheduleFlush(evt.PostID, evt.AuthorID)
	}
}

func (m *likeAggregationManager) scheduleFlush(postID, authorID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.timers[postID]; exists {
		return
	}

	timer := time.AfterFunc(m.window, func() {
		if err := m.flush(postID, authorID); err != nil {
			m.logger.Warn("failed to flush like batch", zap.Error(err))
		}
	})
	m.timers[postID] = timer
}

func (m *likeAggregationManager) flush(postID, authorID int64) error {
	m.mu.Lock()
	if timer, ok := m.timers[postID]; ok {
		timer.Stop()
		delete(m.timers, postID)
	}
	m.mu.Unlock()

	if m.redis == nil {
		m.sendImmediate(authorID, postID, 1)
		return nil
	}

	ctx := context.Background()
	key := likeBatchKey(postID)
	count, err := m.redis.Get(ctx, key).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		return err
	}
	if count <= 0 {
		return nil
	}
	if err := m.redis.Del(ctx, key).Err(); err != nil {
		m.logger.Warn("failed to delete like batch key", zap.Error(err))
	}
	m.sendAggregated(authorID, postID, int(count))
	return nil
}

func (m *likeAggregationManager) sendImmediate(authorID, postID int64, count int) {
	m.sendAggregated(authorID, postID, count)
}

func (m *likeAggregationManager) sendAggregated(authorID, postID int64, count int) {
	if authorID == 0 || postID == 0 || count <= 0 {
		return
	}
	ctx := context.Background()
	title := "Your post has a new like"
	message := fmt.Sprintf("Your post has a new like: %d", count)
	data := model.NotificationData{
		"post_id":    postID,
		"like_count": count,
	}
	if err := m.notificationSvc.CreateNotification(ctx, authorID, model.NotificationTypeCommunityLike, title, &message, data); err != nil {
		m.logger.Warn("failed to create like notification", zap.Error(err))
	}
}

func likeBatchKey(postID int64) string {
	return fmt.Sprintf("cache:notification:post:%d:likes_batch", postID)
}

type postLikedEvent struct {
	PostID   int64 `json:"post_id"`
	AuthorID int64 `json:"author_id"`
	LikerID  int64 `json:"liker_id"`
}

type postCommentedEvent struct {
	PostID           int64   `json:"post_id"`
	AuthorID         int64   `json:"author_id"`
	CommentID        int64   `json:"comment_id"`
	CommentAuthorID  int64   `json:"comment_author_id"`
	MentionedUserIDs []int64 `json:"mentioned_user_ids"`
	CommentSnippet   string  `json:"comment_snippet"`
}
