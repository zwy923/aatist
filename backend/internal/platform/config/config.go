package config

import (
	"time"
)

// Config holds all configuration for the application
type Config struct {
	App      AppConfig
	Postgres PostgresConfig
	Redis    RedisConfig
	MQ       MQConfig
	S3       S3Config
	JWT      JWTConfig
	Email    EmailConfig
	Gateway  GatewayConfig
}

// AppConfig holds application-level configuration
type AppConfig struct {
	Name     string
	Env      string
	HTTPPort int `yaml:"http_port"`
}

// PostgresConfig holds PostgreSQL configuration
type PostgresConfig struct {
	DSN string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Addr string
	DB   int
}

// MQConfig holds message queue configuration
type MQConfig struct {
	Broker string
}

// EmailConfig holds email provider configuration
type EmailConfig struct {
	SendGridAPIKey string `yaml:"sendgrid_api_key"`
	FromEmail      string `yaml:"from_email"`
	FrontendURL    string `yaml:"frontend_url"`
}

// S3Config holds S3 configuration
type S3Config struct {
	Endpoint  string
	Bucket    string
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	UseSSL    bool   `yaml:"use_ssl"`
	PublicURL string `yaml:"public_url"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret     string
	TTLMinutes int `yaml:"ttl_minutes"`
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// GatewayConfig holds Gateway configuration
type GatewayConfig struct {
	// ServiceTimeout is the default timeout for all downstream services
	ServiceTimeout time.Duration `yaml:"service_timeout"`
	// ServiceTimeouts allows per-service timeout configuration
	// Key is service name (e.g., "user-service"), value is timeout duration
	ServiceTimeouts map[string]time.Duration `yaml:"service_timeouts"`
}
