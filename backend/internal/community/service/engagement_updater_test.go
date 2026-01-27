package service

import (
	"context"
	"testing"
	"time"

	"github.com/aatist/backend/internal/platform/log"
	"github.com/redis/go-redis/v9"
)

// MockRedis satisfies redis.Cmdable
type MockRedis struct {
	redis.Cmdable
	latency time.Duration
	ops     int
}

func (m *MockRedis) IncrBy(ctx context.Context, key string, value int64) *redis.IntCmd {
	time.Sleep(m.latency)
	m.ops++
	return redis.NewIntResult(value, nil)
}

func (m *MockRedis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	time.Sleep(m.latency)
	m.ops++
	return redis.NewStatusResult("OK", nil)
}

func (m *MockRedis) Pipeline() redis.Pipeliner {
	return &MockPipeliner{parent: m}
}

// MockPipeliner satisfies redis.Pipeliner
type MockPipeliner struct {
	redis.Pipeliner
	parent *MockRedis
}

func (mp *MockPipeliner) IncrBy(ctx context.Context, key string, value int64) *redis.IntCmd {
	// No sleep, just queuing
	return redis.NewIntResult(value, nil)
}

func (mp *MockPipeliner) Exec(ctx context.Context) ([]redis.Cmder, error) {
	time.Sleep(mp.parent.latency) // One latency for the batch
	mp.parent.ops++
	return nil, nil
}

func BenchmarkEngagementUpdater_Flush(b *testing.B) {
	logger, _ := log.NewLogger("test")
	mockRedis := &MockRedis{latency: 1 * time.Millisecond}
	// We use struct literal to avoid starting the background loop
	updater := &EngagementUpdater{
		redis:    mockRedis,
		trending: nil,
		logger:   logger,
		deltaCh:  make(chan counterDelta, 2048),
		ready:    true,
	}

	// Pre-allocate a map to copy from
	baseMap := make(map[int64]int64)
	for i := 0; i < 50; i++ { // 50 updates per flush
		baseMap[int64(i)] = 1
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Create fresh maps for this iteration
		likes := make(map[int64]int64)
		for k, v := range baseMap {
			likes[k] = v
		}
		b.StartTimer()
		updater.flush(likes, nil)
	}
}
