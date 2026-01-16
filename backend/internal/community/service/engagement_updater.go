package service

import (
	"context"
	"time"

	"github.com/aatist/backend/internal/platform/log"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type counterKind int

const (
	counterLike counterKind = iota
	counterComment

	engagementFlushInterval = 500 * time.Millisecond
)

type counterDelta struct {
	postID int64
	delta  int64
	kind   counterKind
}

// EngagementUpdater batches like/comment counter updates and schedules trending refreshes.
type EngagementUpdater struct {
	redis    redis.Cmdable
	trending *TrendingManager
	logger   *log.Logger
	deltaCh  chan counterDelta
	ready    bool
}

func NewEngagementUpdater(redisClient redis.Cmdable, trending *TrendingManager, logger *log.Logger) *EngagementUpdater {
	updater := &EngagementUpdater{
		redis:    redisClient,
		trending: trending,
		logger:   logger,
		deltaCh:  make(chan counterDelta, 2048),
		ready:    redisClient != nil,
	}
	if updater.ready {
		go updater.loop()
	} else {
		logger.Warn("Engagement updater disabled - Redis unavailable")
	}
	return updater
}

func (u *EngagementUpdater) QueueLikeDelta(postID int64, delta int64) {
	u.enqueue(counterDelta{postID: postID, delta: delta, kind: counterLike})
}

func (u *EngagementUpdater) QueueCommentDelta(postID int64, delta int64) {
	u.enqueue(counterDelta{postID: postID, delta: delta, kind: counterComment})
}

func (u *EngagementUpdater) ScheduleRefresh(postID int64) {
	if u.trending != nil {
		u.trending.ScheduleRefresh(postID)
	}
}

func (u *EngagementUpdater) ClearCounters(postID int64) {
	if !u.ready {
		return
	}
	ctx := context.Background()
	if err := u.redis.Del(ctx, likeCountKey(postID), commentCountKey(postID)).Err(); err != nil {
		u.logger.Warn("failed to clear engagement counters", zap.Int64("post_id", postID), zap.Error(err))
	}
}

func (u *EngagementUpdater) enqueue(delta counterDelta) {
	if !u.ready {
		return
	}
	select {
	case u.deltaCh <- delta:
	default:
		u.logger.Warn("engagement delta queue full", zap.Int64("post_id", delta.postID), zap.Int("kind", int(delta.kind)))
	}
}

func (u *EngagementUpdater) loop() {
	ticker := time.NewTicker(engagementFlushInterval)
	defer ticker.Stop()

	likeDeltas := make(map[int64]int64)
	commentDeltas := make(map[int64]int64)

	for {
		select {
		case delta := <-u.deltaCh:
			switch delta.kind {
			case counterLike:
				likeDeltas[delta.postID] += delta.delta
			case counterComment:
				commentDeltas[delta.postID] += delta.delta
			}
		case <-ticker.C:
			u.flush(likeDeltas, commentDeltas)
		}
	}
}

func (u *EngagementUpdater) flush(likeDeltas, commentDeltas map[int64]int64) {
	if len(likeDeltas) == 0 && len(commentDeltas) == 0 {
		return
	}

	ctx := context.Background()
	changed := make(map[int64]struct{})

	for postID, delta := range likeDeltas {
		if delta == 0 {
			continue
		}
		u.applyDelta(ctx, likeCountKey(postID), delta)
		changed[postID] = struct{}{}
		delete(likeDeltas, postID)
	}

	for postID, delta := range commentDeltas {
		if delta == 0 {
			continue
		}
		u.applyDelta(ctx, commentCountKey(postID), delta)
		changed[postID] = struct{}{}
		delete(commentDeltas, postID)
	}

	for postID := range changed {
		u.logger.Info("Engagement flushed, scheduling refresh", zap.Int64("post_id", postID))
		u.ScheduleRefresh(postID)
	}
}

func (u *EngagementUpdater) applyDelta(ctx context.Context, key string, delta int64) {
	if u.redis == nil {
		return
	}
	result, err := u.redis.IncrBy(ctx, key, delta).Result()
	if err != nil {
		u.logger.Warn("failed to update engagement counter", zap.String("key", key), zap.Error(err))
		return
	}
	if result < 0 {
		if err := u.redis.Set(ctx, key, 0, 0).Err(); err != nil {
			u.logger.Warn("failed to reset negative counter", zap.String("key", key), zap.Error(err))
		}
	}
}
