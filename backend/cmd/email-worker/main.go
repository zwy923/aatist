package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aatist/backend/internal/platform/app"
	"github.com/aatist/backend/internal/platform/log"
	"github.com/aatist/backend/internal/platform/mq"
	"github.com/aatist/backend/internal/user/model"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"go.uber.org/zap"
)

// EmailService handles email sending (placeholder - implement with actual email provider)
type EmailService struct {
	apiKey      string
	fromEmail   string
	frontendURL string
	logger      *zap.Logger
}

func NewEmailService(apiKey, fromEmail, frontendURL string, logger *zap.Logger) *EmailService {
	return &EmailService{
		apiKey:      apiKey,
		fromEmail:   fromEmail,
		frontendURL: frontendURL,
		logger:      logger,
	}
}

// SendVerificationEmail sends an email verification email
func (e *EmailService) SendVerificationEmail(email, name, token string) error {
	verificationLink := fmt.Sprintf("%s/verify?token=%s", e.frontendURL, token)

	subject := "Verify your Aatist account"
	content := fmt.Sprintf(`
Hello %s,

Click the link below to verify your email:

%s

If you didn’t request this, please ignore this email.
`, name, verificationLink)

	from := mail.NewEmail("Aatist", e.fromEmail)
	to := mail.NewEmail(name, email)

	msg := mail.NewSingleEmail(from, subject, to, content, content)
	client := sendgrid.NewSendClient(e.apiKey)

	resp, err := client.Send(msg)
	if err != nil {
		e.logger.Error("SendGrid error", zap.Error(err))
		return err
	}

	if resp.StatusCode >= 400 {
		e.logger.Error("SendGrid API returned error",
			zap.Int("status", resp.StatusCode),
			zap.String("body", resp.Body),
		)
		return fmt.Errorf("SendGrid returned status %d", resp.StatusCode)
	}

	e.logger.Info("Email sent", zap.String("to", email))
	return nil
}

func main() {
	// Load configuration
	cfg, err := app.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger, err := app.InitLogger(cfg.App.Env)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting email worker")

	// Initialize RabbitMQ with retry
	rabbitMQ := mustConnectRabbitMQ(cfg.MQ.Broker, cfg.MQ.PublishConfirmTimeout, logger)
	defer rabbitMQ.Close()

	// Initialize email service
	if cfg.Email.SendGridAPIKey == "" || cfg.Email.FromEmail == "" {
		logger.Fatal("SendGrid credentials are not configured")
	}

	emailService := NewEmailService(
		cfg.Email.SendGridAPIKey,
		cfg.Email.FromEmail,
		cfg.Email.FrontendURL,
		logger.Logger,
	)

	// Start consuming messages
	err = rabbitMQ.ConsumeEmailVerification(func(body []byte) error {
		var msg model.EmailVerificationRequest
		if err := json.Unmarshal(body, &msg); err != nil {
			return fmt.Errorf("failed to unmarshal message: %w", err)
		}

		if msg.Email == "" || msg.Token == "" {
			return fmt.Errorf("invalid message: missing email or token")
		}

		// Send verification email
		if err := emailService.SendVerificationEmail(msg.Email, msg.Name, msg.Token); err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}

		logger.Info("Email verification sent",
			zap.Int64("user_id", msg.UserID),
			zap.String("email", msg.Email),
		)

		return nil
	})

	if err != nil {
		logger.Fatal("Failed to start consuming messages", zap.Error(err))
	}

	logger.Info("Email worker started, waiting for messages...")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down email worker...")
	time.Sleep(1 * time.Second)
	logger.Info("Email worker stopped")
}

func mustConnectRabbitMQ(broker string, confirmTimeout time.Duration, logger *log.Logger) *mq.RabbitMQ {
	if broker == "" {
		logger.Fatal("MQ broker not configured")
	}

	const maxAttempts = 5
	var (
		client *mq.RabbitMQ
		err    error
	)
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		client, err = mq.NewRabbitMQ(broker, confirmTimeout, logger)
		if err == nil {
			logger.Info("Connected to RabbitMQ", zap.String("broker", broker))
			return client
		}

		logger.Warn("Failed to connect to RabbitMQ",
			zap.Int("attempt", attempt),
			zap.Error(err),
		)
		sleep := time.Duration(attempt*2) * time.Second
		time.Sleep(sleep)
	}

	logger.Fatal("Exceeded retries connecting to RabbitMQ", zap.Error(err))
	return nil
}
