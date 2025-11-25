package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aalto-talent-network/backend/internal/platform/cache"
	"github.com/aalto-talent-network/backend/internal/platform/log"
	"github.com/aalto-talent-network/backend/internal/user/repository"
	"github.com/aalto-talent-network/backend/pkg/errs"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	// Password reset token constants
	passwordResetTokenPrefix = "password_reset_token:"
	passwordResetTokenTTL    = 30 * time.Minute // 30 minutes
	passwordResetRateLimitKey = "rate:password_reset:%s"
	passwordResetRateLimit    = 3 // max 3 requests per hour per email
	passwordResetRateLimitWindow = 1 * time.Hour
)

// PasswordResetToken represents the data stored in Redis for password reset
type PasswordResetToken struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
}

// PasswordResetService handles password reset logic
type PasswordResetService struct {
	userRepo repository.UserRepository
	redis    *cache.Redis
	logger   *log.Logger
}

// NewPasswordResetService creates a new password reset service
func NewPasswordResetService(
	userRepo repository.UserRepository,
	redis *cache.Redis,
	logger *log.Logger,
) *PasswordResetService {
	return &PasswordResetService{
		userRepo: userRepo,
		redis:    redis,
		logger:   logger,
	}
}

// GenerateResetToken generates a new password reset token for the given email
// Returns the token and user info if email exists, or nil if email not found (for security, don't reveal if email exists)
func (s *PasswordResetService) GenerateResetToken(ctx context.Context, email string) (token string, userID int64, userName string, err error) {
	// Normalize email
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	// Rate limiting per email
	if err := s.checkRateLimit(ctx, normalizedEmail); err != nil {
		return "", 0, "", err
	}

	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, normalizedEmail)
	if err != nil {
		if err == errs.ErrUserNotFound {
			// Don't reveal if email exists - return success but with empty token
			// The handler will still return success message for security
			s.logger.Info("Password reset requested for non-existent email",
				zap.String("email", normalizedEmail),
			)
			return "", 0, "", nil
		}
		return "", 0, "", fmt.Errorf("failed to find user: %w", err)
	}

	// Generate random UUID token
	token = uuid.New().String()

	// Create token data
	tokenData := PasswordResetToken{
		UserID: user.ID,
		Email:  normalizedEmail,
	}

	// Serialize to JSON
	data, err := json.Marshal(tokenData)
	if err != nil {
		return "", 0, "", fmt.Errorf("failed to marshal token data: %w", err)
	}

	// Store in Redis with TTL
	key := passwordResetTokenPrefix + token
	if err := s.redis.GetClient().Set(ctx, key, data, passwordResetTokenTTL).Err(); err != nil {
		return "", 0, "", fmt.Errorf("failed to store reset token: %w", err)
	}

	s.logger.Info("Generated password reset token",
		zap.Int64("user_id", user.ID),
		zap.String("email", normalizedEmail),
	)

	return token, user.ID, user.Name, nil
}

// VerifyTokenAndReset verifies the reset token and updates the password
func (s *PasswordResetService) VerifyTokenAndReset(ctx context.Context, token, newPassword string) error {
	// Validate new password
	if err := s.validatePassword(newPassword); err != nil {
		return errs.NewAppError(err, 400, "password does not meet requirements")
	}

	key := passwordResetTokenPrefix + token

	// Get token from Redis
	data, err := s.redis.GetClient().Get(ctx, key).Result()
	if err != nil {
		return errs.NewAppError(errs.ErrInvalidToken, 400, "invalid or expired reset token").WithCode("INVALID_RESET_TOKEN")
	}

	// Parse token data
	var tokenData PasswordResetToken
	if err := json.Unmarshal([]byte(data), &tokenData); err != nil {
		return fmt.Errorf("failed to unmarshal token data: %w", err)
	}

	// Verify user exists
	user, err := s.userRepo.FindByID(ctx, tokenData.UserID)
	if err != nil {
		return errs.NewAppError(errs.ErrUserNotFound, 404, "user not found").WithCode(errs.CodeUserNotFound)
	}

	// Verify email matches
	if !strings.EqualFold(user.Email, tokenData.Email) {
		return errs.NewAppError(errs.ErrInvalidToken, 400, "email mismatch").WithCode("EMAIL_MISMATCH")
	}

	// Hash new password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, tokenData.UserID, string(passwordHash)); err != nil {
		s.logger.Error("Failed to update password", zap.Error(err))
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Delete token (one-time use)
	if err := s.redis.GetClient().Del(ctx, key).Err(); err != nil {
		s.logger.Warn("Failed to delete reset token", zap.Error(err))
		// Non-critical, continue
	}

	s.logger.Info("Password reset successful",
		zap.Int64("user_id", tokenData.UserID),
	)

	return nil
}

// checkRateLimit checks if the email has exceeded the rate limit for password reset requests
func (s *PasswordResetService) checkRateLimit(ctx context.Context, email string) error {
	if s.redis == nil {
		return nil
	}

	key := fmt.Sprintf(passwordResetRateLimitKey, email)
	client := s.redis.GetClient()

	count, err := client.Incr(ctx, key).Result()
	if err != nil {
		s.logger.Warn("Rate limit check failed", zap.Error(err))
		return nil // Fail open
	}

	if count == 1 {
		client.Expire(ctx, key, passwordResetRateLimitWindow)
	}

	if count > int64(passwordResetRateLimit) {
		return errs.NewAppError(errs.ErrRateLimitExceeded, 429, "too many password reset requests, please try again later")
	}

	return nil
}

// validatePassword validates password strength
func (s *PasswordResetService) validatePassword(password string) error {
	if len(password) < 10 {
		return fmt.Errorf("password must be at least 10 characters")
	}

	hasLower := false
	hasUpper := false
	hasNumber := false
	hasSpecial := false

	for _, c := range password {
		switch {
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= '0' && c <= '9':
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*()_-+=[]{}|\\:;\"'<>,.?/~`", c):
			hasSpecial = true
		}
	}

	if !hasLower || !hasUpper || !hasNumber || !hasSpecial {
		return fmt.Errorf("password must include uppercase, lowercase, number, and symbol characters")
	}

	return nil
}

