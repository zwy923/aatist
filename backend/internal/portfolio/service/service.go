package service

import (
	"context"

	"github.com/aatist/backend/internal/portfolio/model"
	"github.com/aatist/backend/internal/portfolio/repository"
	"github.com/aatist/backend/pkg/errs"
)

// ProjectService defines the interface for project operations
type ProjectService interface {
	GetUserProjects(ctx context.Context, userID int64) ([]*model.Project, error)
	// GetUserPortfolioForProfile returns projects for a profile; owner sees all, others only published+public.
	GetUserPortfolioForProfile(ctx context.Context, profileUserID int64, viewerUserID *int64) ([]*model.Project, error)
	GetProject(ctx context.Context, id int64) (*model.Project, error)
	GetPublicProjects(ctx context.Context, limit, offset int) ([]*model.Project, error)
	CreateProject(ctx context.Context, userID int64, project *model.Project) error
	UpdateProject(ctx context.Context, userID int64, project *model.Project) error
	DeleteProject(ctx context.Context, userID int64, id int64) error
}

// UserServiceClient defines interface for checking user profile visibility
// This will be called via HTTP to user-service
type UserServiceClient interface {
	CheckProfileVisibility(ctx context.Context, userID int64, viewerEmail *string) (bool, error)
}

type projectService struct {
	projectRepo repository.ProjectRepository
	userClient  UserServiceClient
}

func NewProjectService(projectRepo repository.ProjectRepository, userClient UserServiceClient) ProjectService {
	return &projectService{
		projectRepo: projectRepo,
		userClient:  userClient,
	}
}

func (s *projectService) GetUserProjects(ctx context.Context, userID int64) ([]*model.Project, error) {
	return s.projectRepo.FindByUserID(ctx, userID)
}

func (s *projectService) GetUserPortfolioForProfile(ctx context.Context, profileUserID int64, viewerUserID *int64) ([]*model.Project, error) {
	if viewerUserID != nil && *viewerUserID == profileUserID {
		return s.projectRepo.FindByUserID(ctx, profileUserID)
	}
	return s.projectRepo.FindPublishedPublicByUserID(ctx, profileUserID)
}

func (s *projectService) GetProject(ctx context.Context, id int64) (*model.Project, error) {
	return s.projectRepo.FindByID(ctx, id)
}

func (s *projectService) GetPublicProjects(ctx context.Context, limit, offset int) ([]*model.Project, error) {
	return s.projectRepo.FindAllPublic(ctx, limit, offset)
}

func (s *projectService) CreateProject(ctx context.Context, userID int64, project *model.Project) error {
	if project.Title == "" {
		return errs.NewAppError(errs.ErrInvalidInput, 400, "title is required")
	}
	project.UserID = userID
	return s.projectRepo.Create(ctx, project)
}

func (s *projectService) UpdateProject(ctx context.Context, userID int64, project *model.Project) error {
	if project.Title == "" {
		return errs.NewAppError(errs.ErrInvalidInput, 400, "title is required")
	}

	// Verify ownership
	existing, err := s.projectRepo.FindByID(ctx, project.ID)
	if err != nil {
		return err
	}
	if existing.UserID != userID {
		return errs.NewAppError(errs.ErrUnauthorized, 403, "not authorized to update this project")
	}

	project.UserID = userID
	return s.projectRepo.Update(ctx, project)
}

func (s *projectService) DeleteProject(ctx context.Context, userID int64, id int64) error {
	// Verify ownership
	existing, err := s.projectRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if existing.UserID != userID {
		return errs.NewAppError(errs.ErrUnauthorized, 403, "not authorized to delete this project")
	}

	return s.projectRepo.Delete(ctx, id, userID)
}
