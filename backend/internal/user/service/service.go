package service

import (
	"context"

	"github.com/aatist/backend/internal/user/model"
)

// Tokens represents authentication tokens
type Tokens struct {
	AccessToken  string
	RefreshToken string
}

// AuthService defines the interface for authentication operations
type RegisterInput struct {
	Email                  string
	Password               string
	Name                   string
	IP                     string
	Role                   model.Role
	// Student/Alumni fields
	StudentID              *string
	PreferredName          *string
	School                 *string
	Faculty                *string
	Major                  *string
	// Organization fields
	OrganizationName       *string
	OrganizationBio        *string
	ContactTitle           *string
	IsAffiliatedWithSchool *bool
	OrgSize                *int
}

// GoogleOAuthProfile holds verified claims after Google ID token validation.
type GoogleOAuthProfile struct {
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
	HostedDomain  string // Google Workspace domain, if any
}

// AuthService defines the interface for authentication operations
type AuthService interface {
	// Register registers a new user
	Register(ctx context.Context, input RegisterInput) (*model.User, *Tokens, error)

	// Login authenticates a user
	Login(ctx context.Context, email, password, ip, loginType string) (*model.User, *Tokens, error)

	// RegisterOrLoginGoogle finds or creates an organization (client) user from Google sign-in.
	RegisterOrLoginGoogle(ctx context.Context, profile GoogleOAuthProfile, ip string) (*model.User, *Tokens, error)

	// SaveGoogleOAuthState stores a short-lived CSRF state for the OAuth redirect flow.
	SaveGoogleOAuthState(ctx context.Context, state string) error
	// ConsumeGoogleOAuthState validates and consumes a state value (one-time use).
	ConsumeGoogleOAuthState(ctx context.Context, state string) error

	// RefreshToken refreshes an access token using a refresh token
	RefreshToken(ctx context.Context, refreshToken string) (*Tokens, error)

	// Logout invalidates a refresh token
	Logout(ctx context.Context, refreshToken string) error

	// VerifyEmail verifies a user's email using a verification token
	VerifyEmail(ctx context.Context, token string) error

	// CheckEmailExists checks if an email is already registered
	CheckEmailExists(ctx context.Context, email string) (bool, error)

	// ChangePassword changes user's password (requires current password verification)
	ChangePassword(ctx context.Context, userID int64, currentPassword, newPassword string) error
}

// UserSummary represents a lightweight user profile for display in lists/cards
type UserSummary struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	Role      string  `json:"role"`
	School    *string `json:"school,omitempty"`
	Major     *string `json:"major,omitempty"`
}
