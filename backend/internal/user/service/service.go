package service

import (
	"context"

	"github.com/aalto-talent-network/backend/internal/user/model"
)

// Tokens represents authentication tokens
type Tokens struct {
	AccessToken  string
	RefreshToken string
}

// AuthService defines the interface for authentication operations
type AuthService interface {
	// Register registers a new user
	Register(ctx context.Context, email, password, name, ip string) (*model.User, *Tokens, error)

	// Login authenticates a user
	Login(ctx context.Context, email, password, ip string) (*model.User, *Tokens, error)

	// RefreshToken refreshes an access token using a refresh token
	RefreshToken(ctx context.Context, refreshToken string) (*Tokens, error)

	// Logout invalidates a refresh token
	Logout(ctx context.Context, refreshToken string) error

	// VerifyEmail verifies a user's email using a verification token
	VerifyEmail(ctx context.Context, token string) error
}

