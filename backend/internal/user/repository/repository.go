package repository

import (
	"context"
	"time"

	"github.com/aalto-talent-network/backend/internal/user/model"
)

// UserRepository defines the interface for user data operations
type ProfileUpdate struct {
	UserID int64
	Fields map[string]interface{}
}

type UserRepository interface {
	// FindByEmail finds a user by email
	FindByEmail(ctx context.Context, email string) (*model.User, error)

	// FindByID finds a user by ID
	FindByID(ctx context.Context, id int64) (*model.User, error)

	// ExistsByEmail checks if an email is already registered
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// UpdateProfile updates profile-related fields
	UpdateProfile(ctx context.Context, update ProfileUpdate) (*model.User, error)

	// UpdateAvatarURL updates the avatar URL independently
	UpdateAvatarURL(ctx context.Context, userID int64, avatarURL string) (*model.User, error)

	// UpdatePassword updates user's password hash
	UpdatePassword(ctx context.Context, userID int64, passwordHash string) error

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

// SavedItemRepository defines the interface for saved items operations
type SavedItemRepository interface {
	// FindByUserID finds all saved items for a user
	FindByUserID(ctx context.Context, userID int64) ([]*model.SavedItem, error)

	// FindByUserIDAndType finds saved items by user and type
	FindByUserIDAndType(ctx context.Context, userID int64, itemType model.SavedItemType) ([]*model.SavedItem, error)

	// Create creates a new saved item
	Create(ctx context.Context, item *model.SavedItem) error

	// Delete deletes a saved item
	Delete(ctx context.Context, userID int64, itemID int64, itemType model.SavedItemType) error

	// Exists checks if a saved item exists
	Exists(ctx context.Context, userID int64, itemID int64, itemType model.SavedItemType) (bool, error)
}
