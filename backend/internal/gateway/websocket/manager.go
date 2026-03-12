package websocket

import (
	"context"
	"fmt"
	"time"

	"github.com/aatist/backend/internal/platform/cache"
	"github.com/aatist/backend/internal/platform/log"
	"go.uber.org/zap"
)

const (
	// OnlineStatusKeyPrefix is the prefix for online status keys in Redis
	OnlineStatusKeyPrefix = "user:online:"
	// OnlineStatusTTL is the time-to-live for online status keys
	OnlineStatusTTL = 60 * time.Second
)

// Manager handles WebSocket connections, online status and chat hub
type Manager struct {
	redis  *cache.Redis
	logger *log.Logger
	hub    *Hub
}

// NewManager creates a new WebSocket manager. persist can be nil to skip saving messages.
func NewManager(redis *cache.Redis, logger *log.Logger, persist PersistMessageFunc) *Manager {
	return &Manager{
		redis:  redis,
		logger: logger,
		hub:    NewHub(logger, persist),
	}
}

// SetUserOnline sets the user as online in Redis
func (m *Manager) SetUserOnline(ctx context.Context, userID string) error {
	key := fmt.Sprintf("%s%s", OnlineStatusKeyPrefix, userID)
	// Set key with TTL. Value can be anything, e.g., timestamp or "1"
	err := m.redis.GetClient().Set(ctx, key, time.Now().Unix(), OnlineStatusTTL).Err()
	if err != nil {
		m.logger.Error("Failed to set user online status", zap.String("userID", userID), zap.Error(err))
		return err
	}
	return nil
}

// RemoveUserOnline removes the user online status from Redis
func (m *Manager) RemoveUserOnline(ctx context.Context, userID string) error {
	key := fmt.Sprintf("%s%s", OnlineStatusKeyPrefix, userID)
	err := m.redis.GetClient().Del(ctx, key).Err()
	if err != nil {
		m.logger.Error("Failed to remove user online status", zap.String("userID", userID), zap.Error(err))
		return err
	}
	return nil
}
