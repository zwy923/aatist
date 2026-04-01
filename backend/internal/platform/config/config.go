package config

import (
	"time"
)

// Config holds all configuration for the application
type Config struct {
	App         AppConfig
	Postgres    PostgresConfig
	Redis       RedisConfig
	MQ          MQConfig
	S3          S3Config
	JWT         JWTConfig
	Email       EmailConfig
	Gateway     GatewayConfig
	GoogleOAuth GoogleOAuthConfig
}

// GoogleOAuthConfig configures Sign in with Google (OAuth2 authorization code flow).
// RedirectURI must exactly match a URI allowed in Google Cloud Console.
type GoogleOAuthConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectURI  string `yaml:"redirect_uri"`
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
	Broker                string
	PublishConfirmTimeout time.Duration `yaml:"publish_confirm_timeout"`
}

// EmailConfig holds email provider configuration
type EmailConfig struct {
	SendGridAPIKey           string   `yaml:"sendgrid_api_key"`
	FromEmail                string   `yaml:"from_email"`
	FrontendURL              string   `yaml:"frontend_url"`
	AutoVerifiedDomains      []string `yaml:"auto_verified_domains"`       // School email domains that auto-verify student/alumni accounts
	DisableEmailVerification bool     `yaml:"disable_email_verification"` // When true, skip sending verification email and allow login without verification
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
