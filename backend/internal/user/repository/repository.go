package repository

import (
	"context"
	"time"

	"github.com/aalto-talent-network/backend/internal/user/model"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// FindByEmail finds a user by email
	FindByEmail(ctx context.Context, email string) (*model.User, error)

	// FindByID finds a user by ID
	FindByID(ctx context.Context, id int64) (*model.User, error)

	// CreateUser creates a new user
	CreateUser(ctx context.Context, user *model.User) error

	// UpdateLoginInfo updates login-related information
	UpdateLoginInfo(ctx context.Context, userID int64, lastLogin *time.Time, failedAttempts int, lockedUntil *time.Time) error

	// SetEmailVerified sets email verification status
	SetEmailVerified(ctx context.Context, userID int64) error

	// SetFailedAttempts sets failed login attempts count
	SetFailedAttempts(ctx context.Context, userID int64, attempts int) error

	// LockAccount locks an account until the specified time
	LockAccount(ctx context.Context, userID int64, until *time.Time) error
}

