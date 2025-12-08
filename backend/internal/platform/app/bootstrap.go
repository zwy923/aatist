package app

import (
	"fmt"
	"os"
	"time"

	"github.com/aatist/backend/internal/platform/cache"
	"github.com/aatist/backend/internal/platform/config"
	"github.com/aatist/backend/internal/platform/db"
	"github.com/aatist/backend/internal/platform/log"
	"github.com/aatist/backend/internal/platform/mq"
	"go.uber.org/zap"
)

// LoadConfig loads configuration from file or environment
func LoadConfig() (*config.Config, error) {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "configs/config.yaml"
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return cfg, nil
}

// InitLogger initializes and returns a logger
func InitLogger(env string) (*log.Logger, error) {
	logger, err := log.NewLogger(env)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	return logger, nil
}

// InitPostgres initializes PostgreSQL connection
func InitPostgres(dsn string, logger *log.Logger) (*db.Postgres, error) {
	postgres, err := db.NewPostgres(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize PostgreSQL: %w", err)
	}

	logger.Info("Connected to PostgreSQL")
	return postgres, nil
}

// RunMigrations runs database migrations
func RunMigrations(postgres *db.Postgres, logger *log.Logger) error {
	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}

	if err := db.RunMigrations(postgres.GetSQLDB(), migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info("Database migrations completed")
	return nil
}

// InitRedis initializes Redis connection (optional)
// Returns nil if Redis is not configured or initialization fails
func InitRedis(addr string, dbNum int, logger *log.Logger) (*cache.Redis, error) {
	if addr == "" {
		return nil, nil // Redis is optional
	}

	redis, err := cache.NewRedis(addr, dbNum)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis: %w", err)
	}

	logger.Info("Connected to Redis", zap.String("addr", addr))
	return redis, nil
}

// InitRabbitMQ initializes RabbitMQ connection (optional)
// Returns nil if RabbitMQ is not configured or initialization fails
func InitRabbitMQ(broker string, publishConfirmTimeout time.Duration, logger *log.Logger) (*mq.RabbitMQ, error) {
	if broker == "" {
		return nil, nil // RabbitMQ is optional
	}

	rabbitMQ, err := mq.NewRabbitMQ(broker, publishConfirmTimeout, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize RabbitMQ: %w", err)
	}

	logger.Info("Connected to RabbitMQ")
	return rabbitMQ, nil
}
