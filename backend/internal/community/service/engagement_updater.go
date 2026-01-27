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

	if u.redis == nil {
		// Should not happen if loop is running, but defensive check
		for k := range likeDeltas {
			delete(likeDeltas, k)
		}
		for k := range commentDeltas {
			delete(commentDeltas, k)
		}
		return
	}

	ctx := context.Background()
	changed := make(map[int64]struct{})
	pipe := u.redis.Pipeline()

	type pendingCmd struct {
		key string
		cmd *redis.IntCmd
	}
	var cmds []pendingCmd

	for postID, delta := range likeDeltas {
		if delta == 0 {
			continue
		}
		key := likeCountKey(postID)
		cmd := pipe.IncrBy(ctx, key, delta)
		cmds = append(cmds, pendingCmd{key: key, cmd: cmd})
		changed[postID] = struct{}{}
		delete(likeDeltas, postID)
	}

	for postID, delta := range commentDeltas {
		if delta == 0 {
			continue
		}
		key := commentCountKey(postID)
		cmd := pipe.IncrBy(ctx, key, delta)
		cmds = append(cmds, pendingCmd{key: key, cmd: cmd})
		changed[postID] = struct{}{}
		delete(commentDeltas, postID)
	}

	if len(cmds) > 0 {
		_, err := pipe.Exec(ctx)
		if err != nil && err != redis.Nil {
			u.logger.Warn("engagement pipeline exec failed", zap.Error(err))
		}

		for _, item := range cmds {
			res, err := item.cmd.Result()
			if err != nil {
				u.logger.Warn("failed to update engagement counter", zap.String("key", item.key), zap.Error(err))
				continue
			}
			if res < 0 {
				if err := u.redis.Set(ctx, item.key, 0, 0).Err(); err != nil {
					u.logger.Warn("failed to reset negative counter", zap.String("key", item.key), zap.Error(err))
				}
			}
		}
	}

	for postID := range changed {
		u.logger.Info("Engagement flushed, scheduling refresh", zap.Int64("post_id", postID))
		u.ScheduleRefresh(postID)
	}
}
