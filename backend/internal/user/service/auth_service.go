package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aatist/backend/internal/platform/auth"
	"github.com/aatist/backend/internal/platform/cache"
	"github.com/aatist/backend/internal/platform/log"
	"github.com/aatist/backend/internal/platform/metrics"
	"github.com/aatist/backend/internal/user/model"
	"github.com/aatist/backend/internal/user/repository"
	"github.com/aatist/backend/pkg/errs"
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
	passwordMinLength = 6
)

// authService implements AuthService
type authService struct {
	userRepo                 repository.UserRepository
	jwt                      *auth.JWT
	redis                    *cache.Redis
	logger                   *log.Logger
	emailVerificationSvc     *EmailVerificationService
	autoVerifiedDomains      []string // School email domains that auto-verify student/alumni accounts
	disableEmailVerification bool     // When true, skip verification requirement and treat new users as verified
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userRepo repository.UserRepository,
	jwt *auth.JWT,
	redis *cache.Redis,
	logger *log.Logger,
	emailVerificationSvc *EmailVerificationService,
	autoVerifiedDomains []string,
	disableEmailVerification bool,
) AuthService {
	if emailVerificationSvc == nil {
		emailVerificationSvc = NewEmailVerificationService(userRepo, redis, logger)
	}
	// Default to @aalto.fi if no domains provided
	if len(autoVerifiedDomains) == 0 {
		autoVerifiedDomains = []string{"@aalto.fi"}
	}
	return &authService{
		userRepo:                 userRepo,
		jwt:                      jwt,
		redis:                    redis,
		logger:                   logger,
		emailVerificationSvc:     emailVerificationSvc,
		autoVerifiedDomains:      autoVerifiedDomains,
		disableEmailVerification: disableEmailVerification,
	}
}

// isAutoVerifiedEmail checks if an email should be automatically verified
// Only student/alumni roles with school email domains are auto-verified
func (s *authService) isAutoVerifiedEmail(email string, role model.Role) bool {
	// Only auto-verify student and alumni roles
	if !role.IsStudentRole() {
		return false
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	for _, domain := range s.autoVerifiedDomains {
		normalizedDomain := strings.ToLower(strings.TrimSpace(domain))
		if strings.HasSuffix(normalizedEmail, normalizedDomain) {
			return true
		}
	}
	return false
}

// Register registers a new user
func (s *authService) Register(ctx context.Context, input RegisterInput) (*model.User, *Tokens, error) {
	// Input validation
	if err := s.validateEmail(input.Email); err != nil {
		return nil, nil, errs.NewAppError(err, 400, "invalid email format")
	}
	// Password is required for non-OAuth registration
	if input.Password == "" {
		return nil, nil, errs.NewAppError(errs.ErrInvalidInput, 400, "password is required")
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

	// Normalize email to lowercase (email should be case-insensitive)
	normalizedEmail := strings.ToLower(strings.TrimSpace(input.Email))

	// Check if email already exists (using normalized email)
	existingUser, err := s.userRepo.FindByEmail(ctx, normalizedEmail)
	if err != nil && err != errs.ErrUserNotFound {
		s.logger.Error("Failed to check email existence", zap.Error(err))
		return nil, nil, fmt.Errorf("failed to check email: %w", err)
	}
	if existingUser != nil {
		return nil, nil, errs.NewAppError(errs.ErrEmailExists, 409, "email already registered")
	}

	// Hash password (OAuth users may not have password)
	var passwordHashPtr *string
	if input.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
		if err != nil {
			s.logger.Error("Failed to hash password", zap.Error(err))
			return nil, nil, fmt.Errorf("failed to hash password: %w", err)
		}
		hashStr := string(hashed)
		passwordHashPtr = &hashStr
	}

	// Determine role (default to student)
	role := input.Role
	if !role.IsValid() {
		role = model.RoleStudent
	}

	// Role-specific registration constraints aligned with registration flow:
	// - student/alumni: require @aalto.fi email and academic profile fields
	// - organization roles: require organization name + contact title
	if role.IsStudentRole() {
		if !strings.HasSuffix(normalizedEmail, "@aalto.fi") {
			return nil, nil, errs.NewAppError(errs.ErrInvalidInput, 400, "student registration requires Aalto email (@aalto.fi)")
		}
		if input.School == nil || strings.TrimSpace(*input.School) == "" {
			return nil, nil, errs.NewAppError(errs.ErrInvalidInput, 400, "school is required for student registration")
		}
		if input.Faculty == nil || strings.TrimSpace(*input.Faculty) == "" {
			return nil, nil, errs.NewAppError(errs.ErrInvalidInput, 400, "department is required for student registration")
		}
		if input.Major == nil || strings.TrimSpace(*input.Major) == "" {
			return nil, nil, errs.NewAppError(errs.ErrInvalidInput, 400, "program is required for student registration")
		}
	}
	if role.IsOrgRole() {
		if input.OrganizationName == nil || strings.TrimSpace(*input.OrganizationName) == "" {
			return nil, nil, errs.NewAppError(errs.ErrInvalidInput, 400, "company is required for client registration")
		}
		if input.ContactTitle == nil || strings.TrimSpace(*input.ContactTitle) == "" {
			return nil, nil, errs.NewAppError(errs.ErrInvalidInput, 400, "role is required for client registration")
		}
	}

	// Check if email is from verified school domain (for role_verified)
	// All users still need email verification, but school emails get role_verified = true
	roleVerified := s.isAutoVerifiedEmail(normalizedEmail, role)
	if roleVerified {
		s.logger.Info("School email detected, setting role_verified",
			zap.String("email", normalizedEmail),
			zap.String("role", role.String()),
		)
	}

	// Create user (store email in lowercase)
	// When disableEmailVerification is true, treat new users as verified so they can login without verification
	initialVerified := s.disableEmailVerification
	user := &model.User{
		Email:               normalizedEmail,
		PasswordHash:        passwordHashPtr,
		Name:                input.Name,
		Role:                role,
		ProfileVisibility:   model.VisibilityPublic,          // Default to public
		PortfolioVisibility: model.PortfolioVisibilityPublic, // Default to public
		IsVerifiedEmail:     initialVerified,                 // Verified when email verification is disabled, otherwise requires verification
		RoleVerified:        roleVerified,                    // True if email is from verified school domain
		FailedAttempts:      0,
		// Student/Alumni fields
		StudentID: input.StudentID,
		School:    input.School,
		Faculty:   input.Faculty,
		Major:     input.Major,
		// Organization fields
		OrganizationName:       input.OrganizationName,
		OrganizationBio:        input.OrganizationBio,
		ContactTitle:           input.ContactTitle,
		IsAffiliatedWithSchool: false,
		OrgSize:                input.OrgSize,
	}
	if input.IsAffiliatedWithSchool != nil {
		user.IsAffiliatedWithSchool = *input.IsAffiliatedWithSchool
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
		zap.String("role", user.Role.String()),
		zap.Bool("email_verified", user.IsVerifiedEmail),
		zap.Bool("role_verified", user.RoleVerified),
		zap.String("ip", input.IP),
	)
	metrics.RegisterSuccessTotal.Inc()

	// Note: Email verification email will be sent asynchronously via MQ for all users
	// School email domains (e.g., @aalto.fi) get role_verified = true, but still need email verification

	return user, tokens, nil
}

// Login authenticates a user
func (s *authService) Login(ctx context.Context, email, password, ip, loginType string) (*model.User, *Tokens, error) {
	// Input validation
	if err := s.validateEmail(email); err != nil {
		return nil, nil, errs.NewAppError(err, 400, "invalid email format")
	}
	if password == "" {
		return nil, nil, errs.NewAppError(errs.ErrInvalidInput, 400, "password is required")
	}

	// Normalize email to lowercase (email should be case-insensitive)
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	// Rate limiting (use normalized email for consistency)
	if err := s.checkRateLimit(ctx, fmt.Sprintf(loginRateLimitKey, normalizedEmail, ip), loginRateLimit); err != nil {
		s.logger.Warn("Login rate limit exceeded", zap.String("email", normalizedEmail), zap.String("ip", ip))
		return nil, nil, err
	}

	// Find user (using normalized email)
	user, err := s.userRepo.FindByEmail(ctx, normalizedEmail)
	if err != nil {
		// Use same error message for security (don't reveal if email exists)
		s.handleFailedLogin(ctx, nil, ip)
		s.logger.Warn("Login failed: user not found", zap.String("email", email), zap.String("ip", ip))
		metrics.LoginFailureTotal.Inc()
		return nil, nil, errs.NewAppError(errs.ErrInvalidCredentials, 401, "invalid email or password")
	}

	// Check if email is verified (skip when email verification is disabled)
	if !s.disableEmailVerification && !user.IsVerifiedEmail {
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

	// Verify password (OAuth users may not have password)
	if user.PasswordHash == nil || *user.PasswordHash == "" {
		s.handleFailedLogin(ctx, user, ip)
		s.logger.Warn("Login failed: user has no password (OAuth only)", zap.Int64("user_id", user.ID), zap.String("ip", ip))
		metrics.LoginFailureTotal.Inc()
		return nil, nil, errs.NewAppError(errs.ErrInvalidCredentials, 401, "invalid email or password")
	}
	err = bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password))
	if err != nil {
		s.handleFailedLogin(ctx, user, ip)
		s.logger.Warn("Login failed: invalid password", zap.Int64("user_id", user.ID), zap.String("ip", ip))
		metrics.LoginFailureTotal.Inc()
		return nil, nil, errs.NewAppError(errs.ErrInvalidCredentials, 401, "invalid email or password")
	}

	// Validate login flow after password verification to avoid
	// "login success" logs when request is ultimately rejected.
	normalizedLoginType := strings.ToLower(strings.TrimSpace(loginType))
	switch normalizedLoginType {
	case "", "client", "student":
		// accepted values
	default:
		return nil, nil, errs.NewAppError(errs.ErrInvalidInput, 400, "invalid login_type, expected client or student")
	}
	if normalizedLoginType == "client" && user.Role.IsStudentRole() {
		return nil, nil, errs.NewAppError(errs.ErrInvalidInput, 403, "please use student login for student accounts")
	}
	if normalizedLoginType == "student" {
		if !user.Role.IsStudentRole() {
			return nil, nil, errs.NewAppError(errs.ErrInvalidInput, 403, "please use client login for organization accounts")
		}
		if !strings.HasSuffix(strings.ToLower(user.Email), "@aalto.fi") {
			return nil, nil, errs.NewAppError(errs.ErrInvalidInput, 403, "student login requires Aalto email (@aalto.fi)")
		}
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
	emailRegex := strings.Contains(email, "@")
	if !emailRegex {
		return fmt.Errorf("invalid email format")
	}
	if len(strings.Split(email, "@")) != 2 {
		return fmt.Errorf("invalid email format")
	}
	if strings.TrimSpace(email) != email {
		return fmt.Errorf("invalid email format")
	}
	parts := strings.Split(email, "@")
	if parts[0] == "" || parts[1] == "" || !strings.Contains(parts[1], ".") {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

func (s *authService) validatePassword(password string) error {
	if len(password) < passwordMinLength {
		return fmt.Errorf("password must be at least %d characters", passwordMinLength)
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

// CheckEmailExists checks if an email is already registered
func (s *authService) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	if err := s.validateEmail(email); err != nil {
		return false, errs.NewAppError(err, 400, "invalid email format")
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	return s.userRepo.ExistsByEmail(ctx, normalizedEmail)
}

// ChangePassword changes user's password (requires current password verification)
func (s *authService) ChangePassword(ctx context.Context, userID int64, currentPassword, newPassword string) error {
	// Validate new password
	if err := s.validatePassword(newPassword); err != nil {
		return errs.NewAppError(err, 400, "new password does not meet requirements")
	}

	// Get user to verify current password
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	// Check if user has a password (OAuth users may not have password)
	if user.PasswordHash == nil || *user.PasswordHash == "" {
		return errs.NewAppError(errs.ErrInvalidInput, 400, "user does not have a password set")
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(currentPassword)); err != nil {
		return errs.NewAppError(errs.ErrInvalidCredentials, 401, "current password is incorrect")
	}

	// Check that new password is different from current
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(newPassword)); err == nil {
		return errs.NewAppError(errs.ErrInvalidInput, 400, "new password must be different from current password")
	}

	// Hash new password
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		s.logger.Error("Failed to hash new password", zap.Error(err))
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, userID, string(newPasswordHash)); err != nil {
		s.logger.Error("Failed to update password", zap.Error(err))
		return fmt.Errorf("failed to update password: %w", err)
	}

	s.logger.Info("Password changed successfully", zap.Int64("user_id", userID))
	return nil
}
