package repository

import (
	"context"

	"github.com/aalto-talent-network/backend/internal/file/model"
)

// FileRepository defines the interface for file data operations
type FileRepository interface {
	// Create creates a new file record
	Create(ctx context.Context, file *model.File) error

	// FindByID finds a file by ID
	FindByID(ctx context.Context, id int64) (*model.File, error)

	// FindByUserID finds files by user ID
	FindByUserID(ctx context.Context, userID int64) ([]*model.File, error)

	// FindByUserIDAndType finds files by user ID and type
	FindByUserIDAndType(ctx context.Context, userID int64, fileType model.FileType) ([]*model.File, error)

	// Delete deletes a file record
	Delete(ctx context.Context, id int64) error

	// DeleteByUserID deletes all files for a user
	DeleteByUserID(ctx context.Context, userID int64) error

	// FindByObjectKey finds a file by object key
	FindByObjectKey(ctx context.Context, objectKey string) (*model.File, error)

	// Update updates a file record
	Update(ctx context.Context, file *model.File) error
}

