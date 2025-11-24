package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/aalto-talent-network/backend/internal/platform/auth"
	"github.com/aalto-talent-network/backend/internal/platform/cache"
	"github.com/aalto-talent-network/backend/internal/platform/log"
	"github.com/aalto-talent-network/backend/internal/platform/metrics"
	"github.com/aalto-talent-network/backend/internal/user/model"
	"github.com/aalto-talent-network/backend/internal/user/repository"
	"github.com/aalto-talent-network/backend/pkg/errs"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	// Rate limit constants
	registerRateLimitKey = "register:%s:%s" // register:{fingerprint}:{ip}
	loginRateLimitKey    = "login:%s:%s"    // login:{email}:{ip}
	rateLimitWindow      = 1 * time.Minute
	registerRateLimit    = 3 // max 3 registrations per minute per IP
	loginRateLimit       = 5 // max 5 login attempts per minute per email+IP

	// Account lock constants
	maxFailedAttempts = 5
	lockDuration      = 30 * time.Minute

	// Refresh token key prefix
	refreshTokenKeyPrefix = "refresh_token:"

	// Password policy
	passwordMinLength = 10
)

var (
	lowercaseRegex   = regexp.MustCompile(`[a-z]`)
	uppercaseRegex   = regexp.MustCompile(`[A-Z]`)
	numberRegex      = regexp.MustCompile(`[0-9]`)
	specialCharRegex = regexp.MustCompile(`[!@#$%^&*()_\-+=\[\]{}|\\:;"'<>,.?/~]`)
)

// authService implements AuthService
type authService struct {
	userRepo             repository.UserRepository
	jwt                  *auth.JWT
	redis                *cache.Redis
	logger               *log.Logger
	emailVerificationSvc *EmailVerificationService
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userRepo repository.UserRepository,
	jwt *auth.JWT,
	redis *cache.Redis,
	logger *log.Logger,
	emailVerificationSvc *EmailVerificationService,
) AuthService {
	if emailVerificationSvc == nil {
		emailVerificationSvc = NewEmailVerificationService(userRepo, redis, logger)
	}
	return &authService{
		userRepo:             userRepo,
		jwt:                  jwt,
		redis:                redis,
		logger:               logger,
		emailVerificationSvc: emailVerificationSvc,
	}
}

// Register registers a new user
func (s *authService) Register(ctx context.Context, input RegisterInput) (*model.User, *Tokens, error) {
	// Input validation
	if err := s.validateEmail(input.Email); err != nil {
		return nil, nil, errs.NewAppError(err, 400, "invalid email format")
	}
	if err := s.validatePassword(input.Password); err != nil {
		return nil, nil, errs.NewAppError(err, 400, "password does not meet requirements")
	}
	if len(input.Name) < 1 || len(input.Name) > 100 {
		return nil, nil, errs.NewAppError(errs.ErrInvalidInput, 400, "name must be between 1 and 100 characters")
	}

	// Rate limiting (using IP as fingerprint for now)
	// IP should be passed from handler via context
	if err := s.checkRateLimit(ctx, fmt.Sprintf(registerRateLimitKey, "ip", input.IP), registerRateLimit); err != nil {
		s.logger.Warn("Registration rate limit exceeded", zap.String("ip", input.IP))
		return nil, nil, err
	}

	// Check if email already exists
	existingUser, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err != nil && err != errs.ErrUserNotFound {
		s.logger.Error("Failed to check email existence", zap.Error(err))
		return nil, nil, fmt.Errorf("failed to check email: %w", err)
	}
	if existingUser != nil {
		return nil, nil, errs.NewAppError(errs.ErrEmailExists, 409, "email already registered")
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Determine role (default to student)
	role := input.Role
	if !role.IsValid() {
		role = model.RoleStudent
	}

	// Auto-verify @aalto.fi emails
	isVerified := strings.HasSuffix(strings.ToLower(input.Email), "@aalto.fi")

	// Create user
	user := &model.User{
		Email:             input.Email,
		PasswordHash:      string(passwordHash),
		Name:              input.Name,
		Role:              role,
		StudentID:         input.StudentID,
		School:            input.School,
		Faculty:           input.Faculty,
		ProfileVisibility: model.VisibilityPublic, // Default to public
		IsVerifiedEmail:   isVerified,
		FailedAttempts:    0,
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		if err == errs.ErrEmailExists {
			return nil, nil, errs.NewAppError(err, 409, "email already registered")
		}
		s.logger.Error("Failed to create user", zap.Error(err))
		return nil, nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	accessToken, err := s.jwt.GenerateAccessToken(user.ID, user.Role.String(), user.Email)
	if err != nil {
		s.logger.Error("Failed to generate access token", zap.Error(err))
		return nil, nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		s.logger.Error("Failed to generate refresh token", zap.Error(err))
		return nil, nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token in Redis
	if err := s.storeRefreshToken(ctx, refreshToken, user.ID); err != nil {
		s.logger.Error("Failed to store refresh token", zap.Error(err))
		// Non-critical, continue
	}

	tokens := &Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	// Log and metrics
	s.logger.Info("User registered successfully",
		zap.Int64("user_id", user.ID),
		zap.String("email", user.Email),
		zap.String("ip", input.IP),
	)
	metrics.RegisterSuccessTotal.Inc()

	// Note: Email verification will be handled asynchronously via MQ
	// The MQ message will be published by the handler after successful registration

	return user, tokens, nil
}

// Login authenticates a user
func (s *authService) Login(ctx context.Context, email, password, ip string) (*model.User, *Tokens, error) {
	// Input validation
	if err := s.validateEmail(email); err != nil {
		return nil, nil, errs.NewAppError(err, 400, "invalid email format")
	}
	if password == "" {
		return nil, nil, errs.NewAppError(errs.ErrInvalidInput, 400, "password is required")
	}

	// Rate limiting
	if err := s.checkRateLimit(ctx, fmt.Sprintf(loginRateLimitKey, email, ip), loginRateLimit); err != nil {
		s.logger.Warn("Login rate limit exceeded", zap.String("email", email), zap.String("ip", ip))
		return nil, nil, err
	}

	// Find user
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		// Use same error message for security (don't reveal if email exists)
		s.handleFailedLogin(ctx, nil, ip)
		s.logger.Warn("Login failed: user not found", zap.String("email", email), zap.String("ip", ip))
		metrics.LoginFailureTotal.Inc()
		return nil, nil, errs.NewAppError(errs.ErrInvalidCredentials, 401, "invalid email or password")
	}

	// Check if email is verified
	if !user.IsVerifiedEmail {
		s.logger.Warn("Login attempt with unverified email", zap.Int64("user_id", user.ID), zap.String("ip", ip))
		metrics.LoginFailureTotal.Inc()
		return nil, nil, errs.NewAppError(
			fmt.Errorf("email not verified"),
			403,
			"please verify your email before logging in",
		).WithCode("EMAIL_NOT_VERIFIED")
	}

	// Check if account is locked
	if user.IsLocked() {
		s.logger.Warn("Login attempt on locked account", zap.Int64("user_id", user.ID), zap.String("ip", ip))
		metrics.LoginFailureTotal.Inc()
		return nil, nil, errs.NewAppError(errs.ErrAccountLocked, 423, "account is locked")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		s.handleFailedLogin(ctx, user, ip)
		s.logger.Warn("Login failed: invalid password", zap.Int64("user_id", user.ID), zap.String("ip", ip))
		metrics.LoginFailureTotal.Inc()
		return nil, nil, errs.NewAppError(errs.ErrInvalidCredentials, 401, "invalid email or password")
	}

	// Login successful - reset failed attempts and update last login
	now := time.Now()
	if err := s.userRepo.UpdateLoginInfo(ctx, user.ID, &now, 0, nil); err != nil {
		s.logger.Error("Failed to update login info", zap.Error(err))
		// Non-critical, continue
	} else {
		user.LastLoginAt = &now
		user.FailedAttempts = 0
		user.LockedUntil = nil
	}

	// Generate tokens
	accessToken, err := s.jwt.GenerateAccessToken(user.ID, user.Role.String(), user.Email)
	if err != nil {
		s.logger.Error("Failed to generate access token", zap.Error(err))
		return nil, nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		s.logger.Error("Failed to generate refresh token", zap.Error(err))
		return nil, nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token in Redis
	if err := s.storeRefreshToken(ctx, refreshToken, user.ID); err != nil {
		s.logger.Error("Failed to store refresh token", zap.Error(err))
		// Non-critical, continue
	}

	tokens := &Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	// Log and metrics
	s.logger.Info("User logged in successfully",
		zap.Int64("user_id", user.ID),
		zap.String("email", user.Email),
		zap.String("ip", ip),
	)
	metrics.LoginSuccessTotal.Inc()

	return user, tokens, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*Tokens, error) {
	// Validate refresh token
	claims, err := s.jwt.ValidateToken(refreshToken)
	if err != nil {
		return nil, errs.NewAppError(errs.ErrInvalidToken, 401, "invalid refresh token")
	}

	// Check if refresh token exists in Redis (not revoked)
	key := refreshTokenKeyPrefix + refreshToken
	exists, err := s.redis.GetClient().Exists(ctx, key).Result()
	if err != nil {
		s.logger.Error("Failed to check refresh token in Redis", zap.Error(err))
		return nil, fmt.Errorf("failed to validate refresh token: %w", err)
	}
	if exists == 0 {
		return nil, errs.NewAppError(errs.ErrInvalidToken, 401, "refresh token not found or revoked")
	}

	// Get user to verify still exists
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, errs.NewAppError(errs.ErrUserNotFound, 404, "user not found")
	}

	// Generate new tokens
	accessToken, err := s.jwt.GenerateAccessToken(user.ID, user.Role.String(), user.Email)
	if err != nil {
		s.logger.Error("Failed to generate access token", zap.Error(err))
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		s.logger.Error("Failed to generate refresh token", zap.Error(err))
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Rotate refresh token: delete old, store new
	if err := s.redis.GetClient().Del(ctx, key).Err(); err != nil {
		s.logger.Error("Failed to delete old refresh token", zap.Error(err))
	}
	if err := s.storeRefreshToken(ctx, newRefreshToken, user.ID); err != nil {
		s.logger.Error("Failed to store new refresh token", zap.Error(err))
	}

	// Log and metrics
	s.logger.Info("Token refreshed", zap.Int64("user_id", user.ID))
	metrics.JWTRefreshTotal.Inc()

	return &Tokens{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// Logout invalidates a refresh token
func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	key := refreshTokenKeyPrefix + refreshToken
	if err := s.redis.GetClient().Del(ctx, key).Err(); err != nil {
		s.logger.Error("Failed to delete refresh token", zap.Error(err))
		return fmt.Errorf("failed to logout: %w", err)
	}

	s.logger.Info("User logged out")
	return nil
}

// VerifyEmail verifies a user's email using a verification token
func (s *authService) VerifyEmail(ctx context.Context, token string) error {
	// Verify token
	tokenData, err := s.emailVerificationSvc.VerifyToken(ctx, token)
	if err != nil {
		return err
	}

	// Mark email as verified
	if err := s.emailVerificationSvc.MarkEmailAsVerified(ctx, tokenData.UserID); err != nil {
		return err
	}

	return nil
}

// Helper methods

func (s *authService) validateEmail(email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

func (s *authService) validatePassword(password string) error {
	if len(password) < passwordMinLength {
		return fmt.Errorf("password must be at least %d characters", passwordMinLength)
	}

	if !lowercaseRegex.MatchString(password) ||
		!uppercaseRegex.MatchString(password) ||
		!numberRegex.MatchString(password) ||
		!specialCharRegex.MatchString(password) {
		return fmt.Errorf("password must include uppercase, lowercase, number, and symbol characters")
	}

	return nil
}

func (s *authService) checkRateLimit(ctx context.Context, key string, limit int) error {
	client := s.redis.GetClient()
	count, err := client.Incr(ctx, key).Result()
	if err != nil {
		s.logger.Error("Failed to check rate limit", zap.Error(err))
		// Fail open - allow request if Redis is down
		return nil
	}

	if count == 1 {
		// Set expiration on first request
		client.Expire(ctx, key, rateLimitWindow)
	}

	if count > int64(limit) {
		return errs.NewAppError(errs.ErrRateLimitExceeded, 429, "rate limit exceeded")
	}

	return nil
}

func (s *authService) handleFailedLogin(ctx context.Context, user *model.User, ip string) {
	if user == nil {
		return
	}

	newAttempts := user.FailedAttempts + 1
	var lockedUntil *time.Time

	if newAttempts >= maxFailedAttempts {
		lockTime := time.Now().Add(lockDuration)
		lockedUntil = &lockTime
		s.logger.Warn("Account locked due to too many failed attempts",
			zap.Int64("user_id", user.ID),
			zap.Int("attempts", newAttempts),
			zap.String("ip", ip),
		)
		metrics.AccountLockedTotal.Inc()
	}

	if err := s.userRepo.UpdateLoginInfo(ctx, user.ID, nil, newAttempts, lockedUntil); err != nil {
		s.logger.Error("Failed to update failed attempts", zap.Error(err))
	}
}

func (s *authService) storeRefreshToken(ctx context.Context, token string, userID int64) error {
	key := refreshTokenKeyPrefix + token
	// Store for 30 days (refresh token TTL)
	return s.redis.GetClient().Set(ctx, key, userID, 30*24*time.Hour).Err()
}
