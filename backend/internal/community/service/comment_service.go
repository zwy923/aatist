package service

import (
	"context"
	"regexp"
	"strconv"

	"github.com/aalto-talent-network/backend/internal/community/model"
	"github.com/aalto-talent-network/backend/internal/community/repository"
	"github.com/aalto-talent-network/backend/internal/platform/log"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var mentionRegex = regexp.MustCompile(`@(?:[^\s@]+?\((\d+)\)|(\d+))`)

// CommentService exposes comment operations.
type CommentService interface {
	CreateComment(ctx context.Context, comment *model.Comment) error
	UpdateComment(ctx context.Context, comment *model.Comment, userID int64) error
	DeleteComment(ctx context.Context, id int64, userID int64) error
	ListComments(ctx context.Context, postID int64, limit, offset int) ([]*model.Comment, error)
}

type commentService struct {
	commentRepo repository.CommentRepository
	postRepo    repository.PostRepository
	publisher   EventPublisher
	trending    *TrendingManager
	engagement  *EngagementUpdater
	logger      *log.Logger
}

func NewCommentService(commentRepo repository.CommentRepository, postRepo repository.PostRepository, redisClient redis.Cmdable, publisher EventPublisher, trending *TrendingManager, engagement *EngagementUpdater, logger *log.Logger) CommentService {
	if trending == nil {
		trending = NewTrendingManager(postRepo, redisClient, logger)
	}
	if engagement == nil {
		engagement = NewEngagementUpdater(redisClient, trending, logger)
	}
	return &commentService{
		commentRepo: commentRepo,
		postRepo:    postRepo,
		publisher:   publisher,
		trending:    trending,
		engagement:  engagement,
		logger:      logger,
	}
}

func (s *commentService) CreateComment(ctx context.Context, comment *model.Comment) error {
	post, err := s.postRepo.FindByID(ctx, comment.PostID)
	if err != nil {
		return err
	}

	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return err
	}

	s.engagement.QueueCommentDelta(comment.PostID, 1)

	if s.publisher != nil {
		payload := PostCommentedEvent{
			PostID:           comment.PostID,
			AuthorID:         post.UserID,
			CommentID:        comment.ID,
			CommentAuthorID:  comment.UserID,
			MentionedUserIDs: extractMentionedUserIDs(comment.Content),
			ParentCommentID:  comment.ParentID,
			CommentSnippet:   snippet(comment.Content, 160),
		}
		if err := s.publisher.PublishCommunityEvent(ctx, EventPostCommented, payload); err != nil {
			s.logger.Warn("failed to publish post commented event", zap.Error(err), zap.Int64("post_id", comment.PostID))
		}
	}

	return nil
}

func (s *commentService) UpdateComment(ctx context.Context, comment *model.Comment, userID int64) error {
	return s.commentRepo.Update(ctx, comment, userID)
}

func (s *commentService) DeleteComment(ctx context.Context, id int64, userID int64) error {
	comment, err := s.commentRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.commentRepo.Delete(ctx, id, userID); err != nil {
		return err
	}
	s.engagement.QueueCommentDelta(comment.PostID, -1)
	return nil
}

func (s *commentService) ListComments(ctx context.Context, postID int64, limit, offset int) ([]*model.Comment, error) {
	comments, err := s.commentRepo.ListByPostID(ctx, postID, limit, offset)
	if err != nil {
		return nil, err
	}
	return buildCommentTree(comments), nil
}

func extractMentionedUserIDs(content string) []int64 {
	matches := mentionRegex.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}

	var ids []int64
	for _, match := range matches {
		var raw string
		if len(match) >= 2 && match[1] != "" {
			raw = match[1]
		} else if len(match) >= 3 && match[2] != "" {
			raw = match[2]
		}
		if raw == "" {
			continue
		}
		id, err := strconv.ParseInt(raw, 10, 64)
		if err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

func snippet(content string, maxLen int) string {
	runes := []rune(content)
	if len(runes) <= maxLen {
		return content
	}
	return string(runes[:maxLen]) + "…"
}

func buildCommentTree(comments []*model.Comment) []*model.Comment {
	if len(comments) == 0 {
		return comments
	}

	nodes := make(map[int64]*model.Comment, len(comments))
	var roots []*model.Comment

	for _, comment := range comments {
		comment.Replies = nil
		nodes[comment.ID] = comment
	}

	for _, comment := range comments {
		if comment.ParentID == nil {
			roots = append(roots, comment)
			continue
		}
		parent, ok := nodes[*comment.ParentID]
		if !ok {
			roots = append(roots, comment)
			continue
		}
		parent.Replies = append(parent.Replies, comment)
	}

	return roots
}
