package service

import (
	"context"

	"github.com/aalto-talent-network/backend/internal/user/model"
	"github.com/aalto-talent-network/backend/internal/user/repository"
	"github.com/aalto-talent-network/backend/pkg/errs"
)

type ProjectService interface {
	GetUserProjects(ctx context.Context, userID int64) ([]*model.PortfolioProject, error)
	GetProject(ctx context.Context, id int64) (*model.PortfolioProject, error)
	CreateProject(ctx context.Context, userID int64, project *model.PortfolioProject) error
	UpdateProject(ctx context.Context, userID int64, project *model.PortfolioProject) error
	DeleteProject(ctx context.Context, userID int64, id int64) error
}

type projectService struct {
	projectRepo repository.ProjectRepository
}

func NewProjectService(projectRepo repository.ProjectRepository) ProjectService {
	return &projectService{
		projectRepo: projectRepo,
	}
}

func (s *projectService) GetUserProjects(ctx context.Context, userID int64) ([]*model.PortfolioProject, error) {
	return s.projectRepo.FindByUserID(ctx, userID)
}

func (s *projectService) GetProject(ctx context.Context, id int64) (*model.PortfolioProject, error) {
	return s.projectRepo.FindByID(ctx, id)
}

func (s *projectService) CreateProject(ctx context.Context, userID int64, project *model.PortfolioProject) error {
	if project.Title == "" {
		return errs.NewAppError(errs.ErrInvalidInput, 400, "title is required")
	}
	project.UserID = userID
	return s.projectRepo.Create(ctx, project)
}

func (s *projectService) UpdateProject(ctx context.Context, userID int64, project *model.PortfolioProject) error {
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
