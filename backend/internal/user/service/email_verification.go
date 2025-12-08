package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aatist/backend/internal/platform/cache"
	"github.com/aatist/backend/internal/platform/log"
	"github.com/aatist/backend/internal/user/repository"
	"github.com/aatist/backend/pkg/errs"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	// Email verification token constants
	emailVerifyTokenPrefix = "email_verify_token:"
	emailVerifyTokenTTL    = 30 * time.Minute // 30 minutes
)

// EmailVerificationToken represents the data stored in Redis for email verification
type EmailVerificationToken struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
}

// EmailVerificationService handles email verification logic
type EmailVerificationService struct {
	userRepo repository.UserRepository
	redis    *cache.Redis
	logger   *log.Logger
}

// NewEmailVerificationService creates a new email verification service
func NewEmailVerificationService(
	userRepo repository.UserRepository,
	redis *cache.Redis,
	logger *log.Logger,
) *EmailVerificationService {
	return &EmailVerificationService{
		userRepo: userRepo,
		redis:    redis,
		logger:   logger,
	}
}

// GenerateVerificationToken generates a new email verification token
func (s *EmailVerificationService) GenerateVerificationToken(ctx context.Context, userID int64, email string) (string, error) {
	// Generate random UUID token
	token := uuid.New().String()

	// Create token data
	tokenData := EmailVerificationToken{
		UserID: userID,
		Email:  email,
	}

	// Serialize to JSON
	data, err := json.Marshal(tokenData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token data: %w", err)
	}

	// Store in Redis with TTL
	key := emailVerifyTokenPrefix + token
	if err := s.redis.GetClient().Set(ctx, key, data, emailVerifyTokenTTL).Err(); err != nil {
		return "", fmt.Errorf("failed to store verification token: %w", err)
	}

	s.logger.Info("Generated email verification token",
		zap.Int64("user_id", userID),
		zap.String("email", email),
	)

	return token, nil
}

// VerifyToken verifies an email verification token
func (s *EmailVerificationService) VerifyToken(ctx context.Context, token string) (*EmailVerificationToken, error) {
	key := emailVerifyTokenPrefix + token

	// Get token from Redis
	data, err := s.redis.GetClient().Get(ctx, key).Result()
	if err != nil {
		return nil, errs.NewAppError(errs.ErrInvalidToken, 400, "invalid or expired verification token").WithCode("INVALID_VERIFICATION_TOKEN")
	}

	// Parse token data
	var tokenData EmailVerificationToken
	if err := json.Unmarshal([]byte(data), &tokenData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token data: %w", err)
	}

	// Verify user exists and email matches
	user, err := s.userRepo.FindByID(ctx, tokenData.UserID)
	if err != nil {
		return nil, errs.NewAppError(errs.ErrUserNotFound, 404, "user not found").WithCode(errs.CodeUserNotFound)
	}

	if user.Email != tokenData.Email {
		return nil, errs.NewAppError(errs.ErrInvalidToken, 400, "email mismatch").WithCode("EMAIL_MISMATCH")
	}

	// Delete token (one-time use)
	if err := s.redis.GetClient().Del(ctx, key).Err(); err != nil {
		s.logger.Warn("Failed to delete verification token", zap.Error(err))
		// Non-critical, continue
	}

	return &tokenData, nil
}

// MarkEmailAsVerified marks a user's email as verified
func (s *EmailVerificationService) MarkEmailAsVerified(ctx context.Context, userID int64) error {
	if err := s.userRepo.SetEmailVerified(ctx, userID); err != nil {
		return fmt.Errorf("failed to mark email as verified: %w", err)
	}

	s.logger.Info("Email verified",
		zap.Int64("user_id", userID),
	)

	return nil
}
