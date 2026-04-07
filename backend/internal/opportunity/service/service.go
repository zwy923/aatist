package service

import (
	"context"
	"fmt"
	"time"

	"github.com/aatist/backend/internal/opportunity/model"
	"github.com/aatist/backend/internal/opportunity/repository"
	"github.com/aatist/backend/pkg/errs"
)

// OpportunityService defines the interface for opportunity operations
type OpportunityService interface {
	// Create creates a new opportunity
	Create(ctx context.Context, input CreateOpportunityInput) (*model.Opportunity, error)

	// Update updates an opportunity (only by creator)
	Update(ctx context.Context, opportunityID, userID int64, input UpdateOpportunityInput) (*model.Opportunity, error)

	// GetByID gets an opportunity by ID
	GetByID(ctx context.Context, id int64) (*model.Opportunity, error)

	// List lists opportunities with filters, sorting, and pagination
	List(ctx context.Context, filter ListOpportunitiesFilter) (*ListOpportunitiesResult, error)

	// ListDistinctLocations returns distinct locations for active opportunities (for filters UI)
	ListDistinctLocations(ctx context.Context) ([]string, error)

	// Delete deletes an opportunity (soft delete)
	Delete(ctx context.Context, opportunityID, userID int64) error

	// ListByUserID lists opportunities created by a user with status filter
	ListByUserID(ctx context.Context, userID int64, status *string, page, limit int) ([]*model.Opportunity, error)

	// UpdateStatus updates the status of an opportunity
	UpdateStatus(ctx context.Context, opportunityID, userID int64, status model.OpportunityStatus) error

	// GetStats returns statistics for an opportunity
	GetStats(ctx context.Context, opportunityID, userID int64) (*repository.OpportunityStats, error)
}

// SavedItemClient is now used instead of FavoriteService
// It's defined in user/service/saved_item_client.go and calls user-service's saved items API

// ApplicationService defines the interface for application operations
type ApplicationService interface {
	// CreateApplication creates a new application
	CreateApplication(ctx context.Context, userID int64, input CreateApplicationInput) (*model.OpportunityApplication, error)

	// GetApplication gets an application by ID
	GetApplication(ctx context.Context, id int64) (*model.OpportunityApplication, error)

	// ListByUserID lists all applications by a user with status filter
	ListByUserID(ctx context.Context, userID int64, status *string, page, limit int) ([]*model.OpportunityApplication, error)

	// ListByOpportunityID lists all applications for an opportunity (for opportunity creator)
	ListByOpportunityID(ctx context.Context, opportunityID, userID int64, page, limit int) ([]*model.OpportunityApplication, error)
}

// CreateOpportunityInput represents input for creating an opportunity
type CreateOpportunityInput struct {
	Title          string
	Organization   string
	Category       string
	BudgetType     string
	BudgetValue    *float64
	Location       string
	DurationMonths *int
	Languages      []string
	StartDate      *time.Time
	Urgent         bool
	Description    *string
	Tags           []string
	CreatedBy      int64
}

// UpdateOpportunityInput represents input for updating an opportunity
type UpdateOpportunityInput struct {
	Title          *string
	Organization   *string
	Category       *string
	BudgetType     *string
	BudgetValue    *float64
	Location       *string
	DurationMonths *int
	Languages      []string
	StartDate      *time.Time
	Urgent         *bool
	Description    *string
	Tags           []string
}

// ListOpportunitiesFilter represents filters for listing opportunities
type ListOpportunitiesFilter struct {
	Category      *string
	Location      *string
	Search        *string // q: title, description, category, tags
	BudgetMin     *float64
	BudgetMax     *float64
	StartDateFrom *time.Time
	StartDateTo   *time.Time
	Languages     []string
	Urgent        *bool
	Status        *string
	Sort          string
	Order         string
	Page          int
	Limit         int
}

// ListOpportunitiesResult represents the result of listing opportunities
type ListOpportunitiesResult struct {
	Data  []*model.Opportunity `json:"data"`
	Page  int                  `json:"page"`
	Limit int                  `json:"limit"`
	Total int64                `json:"total"`
}

// CreateApplicationInput represents input for creating an application
type CreateApplicationInput struct {
	OpportunityID int64
	Message       *string
	CVURL         *string
	PortfolioURL  *string
}

// opportunityService implements OpportunityService
type opportunityService struct {
	oppRepo repository.OpportunityRepository
}

// NewOpportunityService creates a new opportunity service
func NewOpportunityService(oppRepo repository.OpportunityRepository) OpportunityService {
	return &opportunityService{oppRepo: oppRepo}
}

func (s *opportunityService) Create(ctx context.Context, input CreateOpportunityInput) (*model.Opportunity, error) {
	budgetType := model.BudgetType(input.BudgetType)
	if !budgetType.IsValid() {
		return nil, errs.NewAppError(errs.ErrInvalidInput, 400, "invalid budget_type").WithCode(errs.CodeInvalidInput)
	}
	if input.BudgetValue != nil && *input.BudgetValue < 0 {
		return nil, errs.NewAppError(errs.ErrInvalidInput, 400, "budget must not be negative").WithCode(errs.CodeInvalidInput)
	}

	opp := &model.Opportunity{
		Title:          input.Title,
		Organization:   input.Organization,
		Category:       input.Category,
		BudgetType:     budgetType,
		BudgetValue:    input.BudgetValue,
		Location:       input.Location,
		DurationMonths: input.DurationMonths,
		Languages:      input.Languages,
		StartDate:      input.StartDate,
		Urgent:         input.Urgent,
		Description:    input.Description,
		Tags:           input.Tags,
		CreatedBy:      input.CreatedBy,
		Status:         model.StatusActive,
	}

	if err := s.oppRepo.Create(ctx, opp); err != nil {
		return nil, fmt.Errorf("failed to create opportunity: %w", err)
	}

	return opp, nil
}

func (s *opportunityService) Update(ctx context.Context, opportunityID, userID int64, input UpdateOpportunityInput) (*model.Opportunity, error) {
	opp, err := s.oppRepo.FindByID(ctx, opportunityID)
	if err != nil {
		return nil, err
	}

	if opp.CreatedBy != userID {
		return nil, errs.NewAppError(errs.ErrForbidden, 403, "only creator can update opportunity").WithCode(errs.CodeForbidden)
	}

	// Update fields if provided
	if input.Title != nil {
		opp.Title = *input.Title
	}
	if input.Organization != nil {
		opp.Organization = *input.Organization
	}
	if input.Category != nil {
		opp.Category = *input.Category
	}
	if input.BudgetType != nil {
		budgetType := model.BudgetType(*input.BudgetType)
		if !budgetType.IsValid() {
			return nil, errs.NewAppError(errs.ErrInvalidInput, 400, "invalid budget_type").WithCode(errs.CodeInvalidInput)
		}
		opp.BudgetType = budgetType
	}
	if input.BudgetValue != nil {
		if *input.BudgetValue < 0 {
			return nil, errs.NewAppError(errs.ErrInvalidInput, 400, "budget must not be negative").WithCode(errs.CodeInvalidInput)
		}
		opp.BudgetValue = input.BudgetValue
	}
	if input.Location != nil {
		opp.Location = *input.Location
	}
	if input.DurationMonths != nil {
		opp.DurationMonths = input.DurationMonths
	}
	if input.Languages != nil {
		opp.Languages = input.Languages
	}
	if input.StartDate != nil {
		opp.StartDate = input.StartDate
	}
	if input.Urgent != nil {
		opp.Urgent = *input.Urgent
	}
	if input.Description != nil {
		opp.Description = input.Description
	}
	if input.Tags != nil {
		opp.Tags = input.Tags
	}

	if err := s.oppRepo.Update(ctx, opp); err != nil {
		return nil, fmt.Errorf("failed to update opportunity: %w", err)
	}

	return opp, nil
}

func (s *opportunityService) GetByID(ctx context.Context, id int64) (*model.Opportunity, error) {
	return s.oppRepo.FindByID(ctx, id)
}

func (s *opportunityService) List(ctx context.Context, filter ListOpportunitiesFilter) (*ListOpportunitiesResult, error) {
	// Convert sort string to repository sort type
	var sort repository.OpportunityListSort
	switch filter.Sort {
	case "start_date":
		sort = repository.SortStartDate
	case "budget":
		sort = repository.SortBudget
	case "latest", "published_at":
		sort = repository.SortPublishedAt
	default:
		sort = repository.SortPublishedAt
	}

	repoFilter := repository.OpportunityListFilter{
		Category:      filter.Category,
		Location:      filter.Location,
		Search:        filter.Search,
		BudgetMin:     filter.BudgetMin,
		BudgetMax:     filter.BudgetMax,
		StartDateFrom: filter.StartDateFrom,
		StartDateTo:   filter.StartDateTo,
		Languages:     filter.Languages,
		Urgent:        filter.Urgent,
		Status:        filter.Status,
		Sort:          sort,
		Order:         filter.Order,
		Page:          filter.Page,
		Limit:         filter.Limit,
	}

	opportunities, err := s.oppRepo.List(ctx, repoFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to list opportunities: %w", err)
	}

	total, err := s.oppRepo.Count(ctx, repoFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to count opportunities: %w", err)
	}

	return &ListOpportunitiesResult{
		Data:  opportunities,
		Page:  filter.Page,
		Limit: filter.Limit,
		Total: total,
	}, nil
}

func (s *opportunityService) ListDistinctLocations(ctx context.Context) ([]string, error) {
	return s.oppRepo.ListDistinctLocations(ctx)
}

func (s *opportunityService) Delete(ctx context.Context, opportunityID, userID int64) error {
	opp, err := s.oppRepo.FindByID(ctx, opportunityID)
	if err != nil {
		return err
	}

	if opp.CreatedBy != userID {
		return errs.NewAppError(errs.ErrForbidden, 403, "only creator can delete opportunity").WithCode(errs.CodeForbidden)
	}

	return s.oppRepo.Delete(ctx, opportunityID, userID)
}

func (s *opportunityService) ListByUserID(ctx context.Context, userID int64, status *string, page, limit int) ([]*model.Opportunity, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	return s.oppRepo.ListByUserID(ctx, userID, status, limit, offset)
}

func (s *opportunityService) UpdateStatus(ctx context.Context, opportunityID, userID int64, status model.OpportunityStatus) error {
	if !status.IsValid() {
		return errs.NewAppError(errs.ErrInvalidInput, 400, "invalid status").WithCode(errs.CodeInvalidInput)
	}
	return s.oppRepo.UpdateStatus(ctx, opportunityID, userID, status)
}

func (s *opportunityService) GetStats(ctx context.Context, opportunityID, userID int64) (*repository.OpportunityStats, error) {
	return s.oppRepo.GetStats(ctx, opportunityID, userID)
}

// applicationService implements ApplicationService
type applicationService struct {
	appRepo repository.OpportunityApplicationRepository
	oppRepo repository.OpportunityRepository
}

// NewApplicationService creates a new application service
func NewApplicationService(appRepo repository.OpportunityApplicationRepository, oppRepo repository.OpportunityRepository) ApplicationService {
	return &applicationService{
		appRepo: appRepo,
		oppRepo: oppRepo,
	}
}

func (s *applicationService) CreateApplication(ctx context.Context, userID int64, input CreateApplicationInput) (*model.OpportunityApplication, error) {
	// Check if opportunity exists
	_, err := s.oppRepo.FindByID(ctx, input.OpportunityID)
	if err != nil {
		return nil, err
	}

	// Check if already applied
	exists, err := s.appRepo.Exists(ctx, userID, input.OpportunityID)
	if err != nil {
		return nil, fmt.Errorf("failed to check application existence: %w", err)
	}
	if exists {
		return nil, errs.NewAppError(errs.ErrInvalidInput, 400, "already applied to this opportunity").WithCode(errs.CodeInvalidInput)
	}

	app := &model.OpportunityApplication{
		UserID:        userID,
		OpportunityID: input.OpportunityID,
		Message:       input.Message,
		CVURL:         input.CVURL,
		PortfolioURL:  input.PortfolioURL,
		Status:        model.ApplicationStatusPending,
	}

	if err := s.appRepo.Create(ctx, app); err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	return app, nil
}

func (s *applicationService) GetApplication(ctx context.Context, id int64) (*model.OpportunityApplication, error) {
	return s.appRepo.FindByID(ctx, id)
}

func (s *applicationService) ListByUserID(ctx context.Context, userID int64, status *string, page, limit int) ([]*model.OpportunityApplication, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	return s.appRepo.ListByUserID(ctx, userID, status, limit, offset)
}

func (s *applicationService) ListByOpportunityID(ctx context.Context, opportunityID, userID int64, page, limit int) ([]*model.OpportunityApplication, error) {
	// Verify user is the creator of the opportunity
	opp, err := s.oppRepo.FindByID(ctx, opportunityID)
	if err != nil {
		return nil, err
	}

	if opp.CreatedBy != userID {
		return nil, errs.NewAppError(errs.ErrForbidden, 403, "only creator can view applications").WithCode(errs.CodeForbidden)
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	return s.appRepo.ListByOpportunityID(ctx, opportunityID, limit, offset)
}
