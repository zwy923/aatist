package repository

import (
	"context"

	"github.com/aatist/backend/internal/portfolio/model"
)

// ProjectRepository defines the interface for portfolio project operations
type ProjectRepository interface {
	// FindByUserID finds all projects for a user (owner dashboard)
	FindByUserID(ctx context.Context, userID int64) ([]*model.Project, error)

	// FindPublishedPublicByUserID finds projects visible on another user's public profile
	FindPublishedPublicByUserID(ctx context.Context, userID int64) ([]*model.Project, error)

	// FindByID finds a project by ID
	FindByID(ctx context.Context, id int64) (*model.Project, error)

	// FindAll finds all projects with pagination
	FindAll(ctx context.Context, limit, offset int) ([]*model.Project, error)

	// FindAllPublic finds projects whose owners have portfolio_visibility = 'public'
	FindAllPublic(ctx context.Context, limit, offset int) ([]*model.Project, error)

	// Create creates a new project
	Create(ctx context.Context, project *model.Project) error

	// Update updates an existing project
	Update(ctx context.Context, project *model.Project) error

	// Delete deletes a project
	Delete(ctx context.Context, id int64, userID int64) error
}






