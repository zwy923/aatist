package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Redis wraps Redis client
type Redis struct {
	client *redis.Client
}

// NewRedis creates a new Redis client
func NewRedis(addr string, db int) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Redis{client: client}, nil
}

// GetClient returns the underlying Redis client
func (r *Redis) GetClient() *redis.Client {
	return r.client
}

// Close closes the Redis connection
func (r *Redis) Close() error {
	return r.client.Close()
}

// HealthCheck checks if Redis is healthy
func (r *Redis) HealthCheck(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
