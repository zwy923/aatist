package service

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/aatist/backend/internal/community/model"
	"github.com/aatist/backend/internal/community/repository"
	"github.com/aatist/backend/internal/platform/log"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// PostService exposes discussion post operations.
type PostService interface {
	CreatePost(ctx context.Context, post *model.DiscussionPost) error
	UpdatePost(ctx context.Context, post *model.DiscussionPost) error
	DeletePost(ctx context.Context, id int64, userID int64) error
	GetPost(ctx context.Context, id int64) (*model.DiscussionPost, error)
	ListPosts(ctx context.Context, filter repository.PostListFilter) ([]*model.DiscussionPost, error)
	ListUserPosts(ctx context.Context, userID int64, limit, offset int) ([]*model.DiscussionPost, error)
	SearchPosts(ctx context.Context, filter repository.PostSearchFilter) ([]*model.DiscussionPost, error)
	SearchPostsTrending(ctx context.Context, filter repository.PostSearchFilter) ([]*model.DiscussionPost, error)
	GetTrendingPosts(ctx context.Context, limit int) ([]*model.DiscussionPost, error)
	GetStickyPosts(ctx context.Context, limit int) ([]*model.DiscussionPost, error)
}

type postService struct {
	postRepo   repository.PostRepository
	redis      redis.Cmdable
	publisher  EventPublisher
	trending   *TrendingManager
	engagement *EngagementUpdater
	logger     *log.Logger
}

func NewPostService(postRepo repository.PostRepository, redisClient redis.Cmdable, publisher EventPublisher, trending *TrendingManager, engagement *EngagementUpdater, logger *log.Logger) PostService {
	if trending == nil {
		trending = NewTrendingManager(postRepo, redisClient, logger)
	}
	if engagement == nil {
		engagement = NewEngagementUpdater(redisClient, trending, logger)
	}
	return &postService{
		postRepo:   postRepo,
		redis:      redisClient,
		publisher:  publisher,
		trending:   trending,
		engagement: engagement,
		logger:     logger,
	}
}

func (s *postService) CreatePost(ctx context.Context, post *model.DiscussionPost) error {
	if post.Category == "" {
		post.Category = model.PostCategoryGeneral
	}
	if !post.Category.IsValid() {
		post.Category = model.PostCategoryGeneral
	}
	if err := s.postRepo.Create(ctx, post); err != nil {
		return err
	}

	if s.publisher != nil {
		payload := PostCreatedEvent{
			PostID:    post.ID,
			AuthorID:  post.UserID,
			Category:  string(post.Category),
			CreatedAt: post.CreatedAt,
			Tags:      []string(post.Tags),
		}
		if err := s.publisher.PublishCommunityEvent(ctx, EventPostCreated, payload); err != nil {
			s.logger.Warn("failed to publish post created event", zap.Error(err), zap.Int64("post_id", post.ID))
		}
	}

	s.engagement.ScheduleRefresh(post.ID)
	return nil
}

func (s *postService) UpdatePost(ctx context.Context, post *model.DiscussionPost) error {
	if !post.Category.IsValid() {
		post.Category = model.PostCategoryGeneral
	}
	if err := s.postRepo.Update(ctx, post); err != nil {
		return err
	}
	s.engagement.ScheduleRefresh(post.ID)
	return nil
}

func (s *postService) DeletePost(ctx context.Context, id int64, userID int64) error {
	if err := s.postRepo.Delete(ctx, id, userID); err != nil {
		return err
	}
	s.engagement.ClearCounters(id)
	s.trending.Remove(ctx, id)
	return nil
}

func (s *postService) GetPost(ctx context.Context, id int64) (*model.DiscussionPost, error) {
	return s.postRepo.FindByID(ctx, id)
}

func (s *postService) ListPosts(ctx context.Context, filter repository.PostListFilter) ([]*model.DiscussionPost, error) {
	return s.postRepo.List(ctx, filter)
}

func (s *postService) ListUserPosts(ctx context.Context, userID int64, limit, offset int) ([]*model.DiscussionPost, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.postRepo.ListByUserID(ctx, userID, limit, offset)
}

func (s *postService) SearchPosts(ctx context.Context, filter repository.PostSearchFilter) ([]*model.DiscussionPost, error) {
	return s.postRepo.Search(ctx, filter)
}

func (s *postService) SearchPostsTrending(ctx context.Context, filter repository.PostSearchFilter) ([]*model.DiscussionPost, error) {
	ids, err := s.postRepo.SearchIDs(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []*model.DiscussionPost{}, nil
	}
	orderedIDs := s.orderIDsByTrending(ctx, ids)
	return s.postRepo.GetPostsByIDs(ctx, orderedIDs)
}

func (s *postService) GetTrendingPosts(ctx context.Context, limit int) ([]*model.DiscussionPost, error) {
	if limit <= 0 {
		limit = 10
	}
	ids, err := s.trending.GetTopIDs(ctx, int64(limit))
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []*model.DiscussionPost{}, nil
	}
	return s.postRepo.GetPostsByIDs(ctx, ids)
}

func (s *postService) GetStickyPosts(ctx context.Context, limit int) ([]*model.DiscussionPost, error) {
	if limit <= 0 {
		limit = 6
	}
	category := model.PostCategorySticky
	return s.postRepo.List(ctx, repository.PostListFilter{
		Category: &category,
		Limit:    limit,
		Sort:     repository.PostListSortNewest,
	})
}

func (s *postService) orderIDsByTrending(ctx context.Context, ids []int64) []int64 {
	if s.redis == nil || len(ids) == 0 {
		return ids
	}

	ordered, err := s.orderIDsViaRedis(ctx, ids)
	if err != nil {
		s.logger.Warn("failed to order trending posts via redis", zap.Error(err))
		return s.orderIDsByLocalScore(ctx, ids)
	}
	return ordered
}

func (s *postService) orderIDsViaRedis(ctx context.Context, ids []int64) ([]int64, error) {
	if len(ids) == 0 {
		return ids, nil
	}

	tempKey := fmt.Sprintf("community:search:%d", time.Now().UnixNano())
	destKey := tempKey + ":ordered"

	members := make([]redis.Z, 0, len(ids))
	for _, id := range ids {
		members = append(members, redis.Z{
			Member: id,
			Score:  0,
		})
	}

	if err := s.redis.ZAdd(ctx, tempKey, members...).Err(); err != nil {
		return nil, err
	}
	defer s.redis.Del(ctx, tempKey, destKey)

	store := &redis.ZStore{
		Keys:    []string{redisTrendingKey, tempKey},
		Weights: []float64{1, 0},
	}
	if err := s.redis.ZInterStore(ctx, destKey, store).Err(); err != nil {
		return nil, err
	}

	orderedStr, err := s.redis.ZRevRange(ctx, destKey, 0, int64(len(ids))-1).Result()
	if err != nil {
		return nil, err
	}

	seen := make(map[int64]struct{}, len(orderedStr))
	ordered := make([]int64, 0, len(ids))
	for _, member := range orderedStr {
		id, err := strconv.ParseInt(member, 10, 64)
		if err != nil {
			continue
		}
		ordered = append(ordered, id)
		seen[id] = struct{}{}
	}

	for _, id := range ids {
		if _, ok := seen[id]; !ok {
			ordered = append(ordered, id)
		}
	}
	return ordered, nil
}

func (s *postService) orderIDsByLocalScore(ctx context.Context, ids []int64) []int64 {
	if len(ids) == 0 || s.redis == nil {
		return ids
	}

	members := make([]string, len(ids))
	for i, id := range ids {
		members[i] = formatID(id)
	}

	scores, err := s.redis.ZMScore(ctx, redisTrendingKey, members...).Result()
	if err != nil {
		return ids
	}

	type pair struct {
		id    int64
		score float64
		idx   int
	}
	pairs := make([]pair, 0, len(ids))
	for idx, id := range ids {
		score := scores[idx]
		pairs = append(pairs, pair{id: id, score: score, idx: idx})
	}

	sort.SliceStable(pairs, func(i, j int) bool {
		if pairs[i].score == pairs[j].score {
			return pairs[i].idx < pairs[j].idx
		}
		return pairs[i].score > pairs[j].score
	})

	ordered := make([]int64, 0, len(ids))
	for _, p := range pairs {
		ordered = append(ordered, p.id)
	}
	return ordered
}

func formatID(id int64) string {
	return fmt.Sprintf("%d", id)
}
