package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Load loads configuration from YAML file and environment variables
func Load(configPath string) (*Config, error) {
	cfg := &Config{}

	// Load from YAML file if exists
	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		if err == nil {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("failed to parse config file: %w", err)
			}
		}
	}

	// Override with environment variables
	if dsn := os.Getenv("POSTGRES_DSN"); dsn != "" {
		cfg.Postgres.DSN = dsn
	}
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		cfg.Redis.Addr = addr
	} else if cfg.Redis.Addr == "" {
		// Default to localhost if not set in config or env
		cfg.Redis.Addr = "localhost:6379"
	}
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		cfg.JWT.Secret = secret
	}
	if port := os.Getenv("HTTP_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.App.HTTPPort)
	}
	if env := os.Getenv("APP_ENV"); env != "" {
		cfg.App.Env = env
	}
	if broker := os.Getenv("MQ_BROKER"); broker != "" {
		cfg.MQ.Broker = broker
	}
	if confirmTimeout := os.Getenv("MQ_CONFIRM_TIMEOUT"); confirmTimeout != "" {
		if d, err := time.ParseDuration(confirmTimeout); err == nil {
			cfg.MQ.PublishConfirmTimeout = d
		}
	}
	if endpoint := os.Getenv("S3_ENDPOINT"); endpoint != "" {
		cfg.S3.Endpoint = endpoint
	}
	if bucket := os.Getenv("S3_BUCKET"); bucket != "" {
		cfg.S3.Bucket = bucket
	}
	if accessKey := os.Getenv("S3_ACCESS_KEY"); accessKey != "" {
		cfg.S3.AccessKey = accessKey
	}
	if secretKey := os.Getenv("S3_SECRET_KEY"); secretKey != "" {
		cfg.S3.SecretKey = secretKey
	}
	if publicURL := os.Getenv("S3_PUBLIC_URL"); publicURL != "" {
		cfg.S3.PublicURL = publicURL
	}
	if useSSL := os.Getenv("S3_USE_SSL"); useSSL != "" {
		cfg.S3.UseSSL = strings.EqualFold(useSSL, "true") || useSSL == "1"
	}
	if apiKey := os.Getenv("SENDGRID_API_KEY"); apiKey != "" {
		cfg.Email.SendGridAPIKey = apiKey
	}
	if from := os.Getenv("SENDGRID_FROM_EMAIL"); from != "" {
		cfg.Email.FromEmail = from
	}
	if frontend := os.Getenv("FRONTEND_URL"); frontend != "" {
		cfg.Email.FrontendURL = frontend
	}

	// Set default values
	if cfg.App.HTTPPort == 0 {
		cfg.App.HTTPPort = 8080
	}
	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = "CHANGE_ME"
	}
	if cfg.JWT.TTLMinutes == 0 {
		cfg.JWT.TTLMinutes = 60
	}
	if cfg.App.Env == "" {
		cfg.App.Env = "dev"
	}
	if cfg.Email.FrontendURL == "" {
		cfg.Email.FrontendURL = "http://localhost:5173"
	}
	if cfg.MQ.PublishConfirmTimeout == 0 {
		cfg.MQ.PublishConfirmTimeout = 5 * time.Second
	}

	// Calculate JWT TTL durations
	cfg.JWT.AccessTTL = 15 * time.Minute
	cfg.JWT.RefreshTTL = 30 * 24 * time.Hour // 30 days

	// Set default Gateway service timeout (10 seconds)
	if cfg.Gateway.ServiceTimeout == 0 {
		cfg.Gateway.ServiceTimeout = 10 * time.Second
	}
	if cfg.Gateway.ServiceTimeouts == nil {
		cfg.Gateway.ServiceTimeouts = make(map[string]time.Duration)
	}

	return cfg, nil
}
