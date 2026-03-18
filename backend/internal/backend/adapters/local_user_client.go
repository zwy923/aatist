package adapters

import (
	"context"

	"github.com/aatist/backend/internal/portfolio/service"
	"github.com/aatist/backend/internal/user/repository"
)

// LocalUserServiceClient implements portfolio's UserServiceClient via direct repository access.
// Used when user and portfolio run in the same process (modular monolith).
type LocalUserServiceClient struct {
	userRepo repository.UserRepository
}

// NewLocalUserServiceClient creates a new local user service client
func NewLocalUserServiceClient(userRepo repository.UserRepository) service.UserServiceClient {
	return &LocalUserServiceClient{userRepo: userRepo}
}

// CheckProfileVisibility checks if a user's profile is visible to the viewer
func (c *LocalUserServiceClient) CheckProfileVisibility(ctx context.Context, userID int64, viewerEmail *string) (bool, error) {
	user, err := c.userRepo.FindByID(ctx, userID)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, nil
	}
	return user.ProfileVisibility.CanView(viewerEmail), nil
}
