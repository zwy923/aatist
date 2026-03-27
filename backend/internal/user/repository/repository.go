package repository

import (
	"context"
	"time"

	"github.com/aatist/backend/internal/user/model"
)

// UserRepository defines the interface for user data operations
type ProfileUpdate struct {
	UserID int64
	Fields map[string]interface{}
}

type UserSearchFilter struct {
	Query         string
	Faculty       string
	School        string
	Major         string
	Role          string
	Limit         int
	Offset        int
	ExcludeUserID int64 // Exclude current user from talent search results
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

	// UpdateBannerURL updates the profile banner/cover image URL independently
	UpdateBannerURL(ctx context.Context, userID int64, bannerURL string) (*model.User, error)

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

	// Metadata search
	SearchSkills(ctx context.Context, query string, limit int) ([]model.SkillMetadata, error)
	SearchCourses(ctx context.Context, query string, limit int) ([]model.CourseMetadata, error)
	SearchTags(ctx context.Context, tagType string, query string, limit int) ([]model.TagMetadata, error)
	// SearchUsers searches for users based on filter
	SearchUsers(ctx context.Context, filter UserSearchFilter) ([]*model.User, error)
}

// SavedItemRepository defines the interface for saved items operations
type SavedItemRepository interface {
	// FindByUserID finds all saved items for a user
	FindByUserID(ctx context.Context, userID int64) ([]*model.SavedItem, error)

	// FindByUserIDAndType finds saved items by user and type
	FindByUserIDAndType(ctx context.Context, userID int64, itemType model.SavedItemType) ([]*model.SavedItem, error)

	// Create creates a new saved item
	Create(ctx context.Context, item *model.SavedItem) error

	// Delete deletes a saved item by target item ID and type
	Delete(ctx context.Context, userID int64, itemID int64, itemType model.SavedItemType) error

	// DeleteByID deletes a saved item by its unique ID
	DeleteByID(ctx context.Context, userID int64, id int64) error

	// Exists checks if a saved item exists
	Exists(ctx context.Context, userID int64, itemID int64, itemType model.SavedItemType) (bool, error)
}
