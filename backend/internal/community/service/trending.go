package service

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/aalto-talent-network/backend/internal/community/model"
	"github.com/aalto-talent-network/backend/internal/community/repository"
	"github.com/aalto-talent-network/backend/internal/platform/log"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const trendingDecayFactor = 0.05

// TrendingManager encapsulates logic for updating trending scores.
type TrendingManager struct {
	postRepo  repository.PostRepository
	redis     redis.Cmdable
	logger    *log.Logger
	refreshCh chan int64
}

func NewTrendingManager(postRepo repository.PostRepository, redisClient redis.Cmdable, logger *log.Logger) *TrendingManager {
	t := &TrendingManager{
		postRepo:  postRepo,
		redis:     redisClient,
		logger:    logger,
		refreshCh: make(chan int64, 2048),
	}
	go t.run()
	return t
}

// Refresh schedules an asynchronous refresh for the given postID.
func (t *TrendingManager) Refresh(ctx context.Context, postID int64) {
	t.ScheduleRefresh(postID)
}

// Remove deletes a post from the trending sorted set.
func (t *TrendingManager) Remove(ctx context.Context, postID int64) {
	if t.redis == nil {
		return
	}
	if err := t.redis.ZRem(ctx, redisTrendingKey, postID).Err(); err != nil {
		t.logger.Warn("failed to remove post from trending", zap.Int64("post_id", postID), zap.Error(err))
	}
}

// GetTopIDs returns ordered post IDs from trending sorted set.
func (t *TrendingManager) GetTopIDs(ctx context.Context, limit int64) ([]int64, error) {
	if t.redis == nil {
		return []int64{}, nil
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	ids, err := t.redis.ZRevRange(ctx, redisTrendingKey, 0, limit-1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to read trending posts: %w", err)
	}

	result := make([]int64, 0, len(ids))
	for _, idStr := range ids {
		postID, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			result = append(result, postID)
		}
	}
	return result, nil
}

// calculateScore computes Reddit-style hot score.
func (t *TrendingManager) calculateScore(ctx context.Context, post *model.DiscussionPost) float64 {
	likes := t.readCount(ctx, likeCountKey(post.ID))
	comments := t.readCount(ctx, commentCountKey(post.ID))

	engagement := float64(likes + comments + 1) // avoid log(0)
	timeSince := time.Since(post.CreatedAt).Hours()
	if timeSince < 0 {
		timeSince = 0
	}
	return math.Log(engagement) - timeSince*trendingDecayFactor
}

func (t *TrendingManager) readCount(ctx context.Context, key string) int64 {
	if t.redis == nil {
		return 0
	}
	val, err := t.redis.Get(ctx, key).Int64()
	if err != nil {
		return 0
	}
	return val
}

// ScheduleRefresh queues a postID for background refresh.
func (t *TrendingManager) ScheduleRefresh(postID int64) {
	if t.redis == nil {
		return
	}
	select {
	case t.refreshCh <- postID:
	default:
		t.logger.Warn("trending refresh queue full", zap.Int64("post_id", postID))
	}
}

func (t *TrendingManager) run() {
	if t.redis == nil {
		return
	}
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	pending := make(map[int64]struct{})
	ctx := context.Background()

	for {
		select {
		case postID := <-t.refreshCh:
			pending[postID] = struct{}{}
		case <-ticker.C:
			if len(pending) == 0 {
				continue
			}
			for id := range pending {
				t.refreshNow(ctx, id)
				delete(pending, id)
			}
		}
	}
}

func (t *TrendingManager) refreshNow(ctx context.Context, postID int64) {
	post, err := t.postRepo.FindByID(ctx, postID)
	if err != nil {
		t.logger.Warn("failed to fetch post for trending refresh", zap.Int64("post_id", postID), zap.Error(err))
		return
	}

	score := t.calculateScore(ctx, post)
	if err := t.redis.ZAdd(ctx, redisTrendingKey, redis.Z{
		Member: postID,
		Score:  score,
	}).Err(); err != nil {
		t.logger.Warn("failed to update trending score", zap.Int64("post_id", postID), zap.Error(err))
	}
}
