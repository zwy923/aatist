package repository

import (
	"context"
	"time"

	"github.com/aatist/backend/internal/opportunity/model"
)

// OpportunityListSort represents sort options for listing opportunities
type OpportunityListSort string

const (
	SortPublishedAt OpportunityListSort = "published_at"
	SortStartDate   OpportunityListSort = "start_date"
	SortBudget      OpportunityListSort = "budget"
)

// OpportunityListFilter represents filters for listing opportunities
type OpportunityListFilter struct {
	Category      *string
	Location      *string
	Search        *string // q: title, description, category, tags (ILIKE)
	BudgetMin     *float64
	BudgetMax     *float64
	StartDateFrom *time.Time
	StartDateTo   *time.Time
	Languages     []string
	Urgent        *bool
	Status        *string
	Sort          OpportunityListSort
	Order         string // "asc" or "desc"
	Page          int
	Limit         int
}

// OpportunityRepository defines operations on opportunities
type OpportunityRepository interface {
	// Create creates a new opportunity
	Create(ctx context.Context, opp *model.Opportunity) error

	// Update updates an opportunity
	Update(ctx context.Context, opp *model.Opportunity) error

	// FindByID finds an opportunity by ID
	FindByID(ctx context.Context, id int64) (*model.Opportunity, error)

	// List lists opportunities with filters, sorting, and pagination
	List(ctx context.Context, filter OpportunityListFilter) ([]*model.Opportunity, error)

	// Count counts opportunities matching the filter (for pagination)
	Count(ctx context.Context, filter OpportunityListFilter) (int64, error)

	// ListDistinctLocations returns distinct non-empty locations for active opportunities
	ListDistinctLocations(ctx context.Context) ([]string, error)

	// Delete deletes an opportunity (soft delete by setting status to closed)
	Delete(ctx context.Context, id int64, userID int64) error

	// ListByUserID lists opportunities created by a user with status filter
	ListByUserID(ctx context.Context, userID int64, status *string, limit, offset int) ([]*model.Opportunity, error)

	// UpdateStatus updates the status of an opportunity
	UpdateStatus(ctx context.Context, id int64, userID int64, status model.OpportunityStatus) error

	// GetStats returns statistics for an opportunity
	GetStats(ctx context.Context, id int64, userID int64) (*OpportunityStats, error)
}

// OpportunityStats represents statistics for an opportunity
type OpportunityStats struct {
	ApplicationCount int64 `json:"application_count"`
	// Add more stats if needed (e.g., view_count)
}

// Note: Favorites/saved items are now handled by user-service's saved_items table
// No need for OpportunityFavoriteRepository

// OpportunityApplicationRepository defines operations on opportunity applications
type OpportunityApplicationRepository interface {
	// Create creates a new application
	Create(ctx context.Context, app *model.OpportunityApplication) error

	// Update updates an application
	Update(ctx context.Context, app *model.OpportunityApplication) error

	// FindByID finds an application by ID
	FindByID(ctx context.Context, id int64) (*model.OpportunityApplication, error)

	// FindByUserAndOpportunity finds an application by user and opportunity
	FindByUserAndOpportunity(ctx context.Context, userID, opportunityID int64) (*model.OpportunityApplication, error)

	// ListByUserID lists all applications by a user with status filter
	ListByUserID(ctx context.Context, userID int64, status *string, limit, offset int) ([]*model.OpportunityApplication, error)

	// ListByOpportunityID lists all applications for an opportunity
	ListByOpportunityID(ctx context.Context, opportunityID int64, limit, offset int) ([]*model.OpportunityApplication, error)

	// Exists checks if an application exists
	Exists(ctx context.Context, userID, opportunityID int64) (bool, error)
}
