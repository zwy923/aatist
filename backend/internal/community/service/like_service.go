package service

import (
	"context"

	"github.com/aatist/backend/internal/community/repository"
	"github.com/aatist/backend/internal/platform/log"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// LikeService exposes like/unlike operations.
type LikeService interface {
	LikePost(ctx context.Context, postID int64, likerID int64) (bool, error)
	UnlikePost(ctx context.Context, postID int64, likerID int64) error
	HasLiked(ctx context.Context, postID int64, likerID int64) (bool, error)
}

type likeService struct {
	likeRepo   repository.LikeRepository
	postRepo   repository.PostRepository
	publisher  EventPublisher
	trending   *TrendingManager
	engagement *EngagementUpdater
	logger     *log.Logger
}

func NewLikeService(likeRepo repository.LikeRepository, postRepo repository.PostRepository, redisClient redis.Cmdable, publisher EventPublisher, trending *TrendingManager, engagement *EngagementUpdater, logger *log.Logger) LikeService {
	if trending == nil {
		trending = NewTrendingManager(postRepo, redisClient, logger)
	}
	if engagement == nil {
		engagement = NewEngagementUpdater(redisClient, trending, logger)
	}
	return &likeService{
		likeRepo:   likeRepo,
		postRepo:   postRepo,
		publisher:  publisher,
		trending:   trending,
		engagement: engagement,
		logger:     logger,
	}
}

func (s *likeService) LikePost(ctx context.Context, postID int64, likerID int64) (bool, error) {
	exists, err := s.likeRepo.Exists(ctx, postID, likerID)
	if err != nil {
		return false, err
	}
	if exists {
		return false, nil
	}

	if err := s.likeRepo.Create(ctx, postID, likerID); err != nil {
		return false, err
	}

	s.engagement.QueueLikeDelta(postID, 1)

	if s.publisher != nil {
		post, err := s.postRepo.FindByID(ctx, postID)
		if err != nil {
			s.logger.Warn("failed to fetch post for like event", zap.Error(err), zap.Int64("post_id", postID))
		} else {
			payload := PostLikedEvent{
				PostID:   postID,
				AuthorID: post.UserID,
				LikerID:  likerID,
			}
			if err := s.publisher.PublishCommunityEvent(ctx, EventPostLiked, payload); err != nil {
				s.logger.Warn("failed to publish post liked event", zap.Error(err), zap.Int64("post_id", postID))
			}
		}
	}

	return true, nil
}

func (s *likeService) UnlikePost(ctx context.Context, postID int64, likerID int64) error {
	if err := s.likeRepo.Delete(ctx, postID, likerID); err != nil {
		return err
	}
	s.engagement.QueueLikeDelta(postID, -1)
	return nil
}

func (s *likeService) HasLiked(ctx context.Context, postID int64, likerID int64) (bool, error) {
	return s.likeRepo.Exists(ctx, postID, likerID)
}
