package model

import (
	"time"
)

// OpportunityStatus represents the status of an opportunity
type OpportunityStatus string

const (
	StatusDraft     OpportunityStatus = "draft"
	StatusPublished OpportunityStatus = "published"
	StatusClosed    OpportunityStatus = "closed"
	StatusArchived  OpportunityStatus = "archived"

	// Legacy support
	StatusActive OpportunityStatus = "active"
)

// IsValid checks if status is valid
func (s OpportunityStatus) IsValid() bool {
	return s == StatusDraft || s == StatusPublished || s == StatusClosed || s == StatusArchived || s == StatusActive
}

// String returns string representation of status
func (s OpportunityStatus) String() string {
	return string(s)
}

// BudgetType represents the type of budget
type BudgetType string

const (
	BudgetTypeHourly BudgetType = "hourly"
	BudgetTypeFixed  BudgetType = "fixed"
)

// IsValid checks if budget type is valid
func (b BudgetType) IsValid() bool {
	return b == BudgetTypeHourly || b == BudgetTypeFixed
}

// String returns string representation of budget type
func (b BudgetType) String() string {
	return string(b)
}

// ApplicationStatus represents the status of an application
type ApplicationStatus string

const (
	ApplicationStatusPending  ApplicationStatus = "pending"
	ApplicationStatusReviewed ApplicationStatus = "reviewed"
	ApplicationStatusAccepted ApplicationStatus = "accepted"
	ApplicationStatusRejected ApplicationStatus = "rejected"
)

// IsValid checks if application status is valid
func (a ApplicationStatus) IsValid() bool {
	return a == ApplicationStatusPending || a == ApplicationStatusReviewed ||
		a == ApplicationStatusAccepted || a == ApplicationStatusRejected
}

// String returns string representation of application status
func (a ApplicationStatus) String() string {
	return string(a)
}

// Opportunity represents an opportunity in the system
type Opportunity struct {
	ID             int64             `db:"id" json:"id"`
	Title          string            `db:"title" json:"title"`
	Organization   string            `db:"organization" json:"organization"`
	Category       string            `db:"category" json:"category"`
	BudgetType     BudgetType        `db:"budget_type" json:"budget_type"`
	BudgetValue    *float64          `db:"budget_value" json:"budget_value"`
	Location       string            `db:"location" json:"location"`
	DurationMonths *int              `db:"duration_months" json:"duration_months,omitempty"`
	Languages      []string          `db:"languages" json:"languages"`
	StartDate      *time.Time        `db:"start_date" json:"start_date,omitempty"`
	PublishedAt    time.Time         `db:"published_at" json:"published_at"`
	Urgent         bool              `db:"urgent" json:"urgent"`
	Description    *string           `db:"description" json:"description,omitempty"`
	Tags           []string          `db:"tags" json:"tags"`
	CreatedBy      int64             `db:"created_by" json:"created_by"`
	Status         OpportunityStatus `db:"status" json:"status"`
	CreatedAt      time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time         `db:"updated_at" json:"updated_at"`
}

// Note: Favorites/saved items are now handled by user-service's saved_items table
// No need for OpportunityFavorite model

// OpportunityApplication represents an application to an opportunity
type OpportunityApplication struct {
	ID            int64             `db:"id" json:"id"`
	UserID        int64             `db:"user_id" json:"user_id"`
	OpportunityID int64             `db:"opportunity_id" json:"opportunity_id"`
	Message       *string           `db:"message" json:"message,omitempty"`
	CVURL         *string           `db:"cv_url" json:"cv_url,omitempty"`
	PortfolioURL  *string           `db:"portfolio_url" json:"portfolio_url,omitempty"`
	Status        ApplicationStatus `db:"status" json:"status"`
	CreatedAt     time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time         `db:"updated_at" json:"updated_at"`
}
