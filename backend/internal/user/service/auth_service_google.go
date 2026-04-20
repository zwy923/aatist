package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aatist/backend/internal/platform/metrics"
	"github.com/aatist/backend/internal/user/model"
	"github.com/aatist/backend/pkg/errs"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	googleOAuthProvider   = "google"
	oauthGoogleStateKeyPF = "oauth_gstate:"
)

func (s *authService) SaveGoogleOAuthState(ctx context.Context, state string) error {
	if s.redis == nil {
		return fmt.Errorf("redis is required for Google OAuth")
	}
	if strings.TrimSpace(state) == "" {
		return errs.NewAppError(errs.ErrInvalidInput, 400, "invalid oauth state")
	}
	key := oauthGoogleStateKeyPF + state
	return s.redis.GetClient().Set(ctx, key, "1", 10*time.Minute).Err()
}

func (s *authService) ConsumeGoogleOAuthState(ctx context.Context, state string) error {
	if s.redis == nil {
		return fmt.Errorf("redis is required for Google OAuth")
	}
	if strings.TrimSpace(state) == "" {
		return errs.NewAppError(errs.ErrInvalidToken, 400, "missing oauth state")
	}
	key := oauthGoogleStateKeyPF + state
	_, err := s.redis.GetClient().GetDel(ctx, key).Result()
	if err == redis.Nil {
		return errs.NewAppError(errs.ErrInvalidToken, 400, "invalid or expired oauth state")
	}
	if err != nil {
		return fmt.Errorf("oauth state: %w", err)
	}
	return nil
}

// RegisterOrLoginGoogle creates or logs in an organization (client) account via Google.
func (s *authService) RegisterOrLoginGoogle(ctx context.Context, profile GoogleOAuthProfile, ip string) (*model.User, *Tokens, error) {
	subject := strings.TrimSpace(profile.Subject)
	emailNorm := strings.ToLower(strings.TrimSpace(profile.Email))
	if subject == "" || emailNorm == "" {
		return nil, nil, errs.NewAppError(errs.ErrInvalidInput, 400, "invalid Google account (missing subject or email)")
	}
	if err := s.validateEmail(emailNorm); err != nil {
		return nil, nil, errs.NewAppError(err, 400, "invalid email format")
	}
	if !profile.EmailVerified {
		return nil, nil, errs.NewAppError(errs.ErrInvalidInput, 403, "Google email is not verified")
	}

	existingOAuth, err := s.userRepo.FindByOAuth(ctx, googleOAuthProvider, subject)
	if err == nil {
		return s.completeGoogleSession(ctx, existingOAuth, ip)
	}
	if err != errs.ErrUserNotFound {
		return nil, nil, fmt.Errorf("find oauth user: %w", err)
	}

	byEmail, err := s.userRepo.FindByEmail(ctx, emailNorm)
	if err == nil {
		if byEmail.Role.IsStudentRole() {
			return nil, nil, errs.NewAppError(
				errs.ErrInvalidInput,
				409,
				"this email is registered as a student; use student sign-in",
			).WithCode(errs.CodeOAuthEmailConflict)
		}
		if byEmail.PasswordHash != nil && strings.TrimSpace(*byEmail.PasswordHash) != "" {
			return nil, nil, errs.NewAppError(
				errs.ErrEmailExists,
				409,
				"an account with this email already exists; sign in with password",
			).WithCode(errs.CodeOAuthEmailConflict)
		}
		if byEmail.OAuthSubject != nil && *byEmail.OAuthSubject != subject {
			return nil, nil, errs.NewAppError(
				errs.ErrInvalidInput,
				409,
				"this email is linked to another sign-in method",
			).WithCode(errs.CodeOAuthEmailConflict)
		}
		return nil, nil, errs.NewAppError(
			errs.ErrEmailExists,
			409,
			"email already registered",
		).WithCode(errs.CodeOAuthEmailConflict)
	}
	if err != errs.ErrUserNotFound {
		return nil, nil, fmt.Errorf("find by email: %w", err)
	}

	if err := s.checkRateLimit(ctx, fmt.Sprintf(registerRateLimitKey, "ip", ip), registerRateLimit); err != nil {
		s.logger.Warn("Google OAuth registration rate limit exceeded", zap.String("ip", ip))
		return nil, nil, err
	}

	orgName := "Independent"
	if hd := strings.TrimSpace(profile.HostedDomain); hd != "" {
		orgName = hd
	}
	name := strings.TrimSpace(profile.Name)
	if name == "" {
		name = strings.Split(emailNorm, "@")[0]
	}

	prov := googleOAuthProvider
	user := &model.User{
		Email:                  emailNorm,
		PasswordHash:           nil,
		Name:                   name,
		Role:                   model.RoleOrgTeam,
		ProfileVisibility:      model.VisibilityPublic,
		PortfolioVisibility:    model.PortfolioVisibilityPublic,
		IsVerifiedEmail:        true,
		RoleVerified:           false,
		FailedAttempts:         0,
		OAuthProvider:          &prov,
		OAuthSubject:           &subject,
		OrganizationName:       &orgName,
		IsAffiliatedWithSchool: false,
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		if err == errs.ErrEmailExists {
			return nil, nil, errs.NewAppError(err, 409, "email already registered").WithCode(errs.CodeOAuthEmailConflict)
		}
		s.logger.Error("Failed to create Google OAuth user", zap.Error(err))
		return nil, nil, fmt.Errorf("failed to create user: %w", err)
	}

	tokens, err := s.issueTokens(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	s.logger.Info("User registered via Google",
		zap.Int64("user_id", user.ID),
		zap.String("email", user.Email),
		zap.String("ip", ip),
	)
	metrics.RegisterSuccessTotal.Inc()
	return user, tokens, nil
}

func (s *authService) completeGoogleSession(ctx context.Context, user *model.User, ip string) (*model.User, *Tokens, error) {
	if !s.disableEmailVerification && !user.IsVerifiedEmail {
		return nil, nil, errs.NewAppError(
			fmt.Errorf("email not verified"),
			403,
			"please verify your email before logging in",
		).WithCode("EMAIL_NOT_VERIFIED")
	}
	if user.IsLocked() {
		return nil, nil, errs.NewAppError(errs.ErrAccountLocked, 423, "account is locked")
	}

	now := time.Now()
	if err := s.userRepo.UpdateLoginInfo(ctx, user.ID, &now, 0, nil); err != nil {
		s.logger.Error("Failed to update login info", zap.Error(err))
	} else {
		user.LastLoginAt = &now
		user.FailedAttempts = 0
		user.LockedUntil = nil
	}

	tokens, err := s.issueTokens(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	s.logger.Info("User logged in via Google",
		zap.Int64("user_id", user.ID),
		zap.String("email", user.Email),
		zap.String("ip", ip),
	)
	metrics.LoginSuccessTotal.Inc()
	return user, tokens, nil
}

func (s *authService) issueTokens(ctx context.Context, user *model.User) (*Tokens, error) {
	accessToken, err := s.jwt.GenerateAccessToken(user.ID, user.Role.String(), user.Email)
	if err != nil {
		s.logger.Error("Failed to generate access token", zap.Error(err))
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}
	refreshToken, err := s.jwt.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		s.logger.Error("Failed to generate refresh token", zap.Error(err))
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	if err := s.storeRefreshToken(ctx, refreshToken, user.ID); err != nil {
		s.logger.Error("Failed to store refresh token", zap.Error(err))
	}
	return &Tokens{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}
