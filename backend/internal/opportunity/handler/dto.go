package handler

import (
	"time"
)

// CreateOpportunityRequest represents the request body for creating an opportunity
type CreateOpportunityRequest struct {
	Title          string     `json:"title" binding:"required"`
	Organization   string     `json:"organization"`
	Position       string     `json:"position"`
	Category       string     `json:"category" binding:"required"`
	BudgetType     string     `json:"budgetType" binding:"required,oneof=hourly fixed"`
	BudgetValue    *float64   `json:"budgetValue" binding:"omitempty,gte=0"`
	Location       string     `json:"location" binding:"required"`
	DurationMonths *int       `json:"durationMonths"`
	Languages      []string   `json:"languages"`
	StartDate      *time.Time `json:"startDate"`
	Tags           []string   `json:"tags"`
	Urgent         bool       `json:"urgent"`
	Description    *string    `json:"description"`
}

// UpdateOpportunityRequest represents the request body for updating an opportunity
type UpdateOpportunityRequest struct {
	Title          *string    `json:"title"`
	Organization   *string    `json:"organization"`
	Position       *string    `json:"position"`
	Category       *string    `json:"category"`
	BudgetType     *string    `json:"budgetType" binding:"omitempty,oneof=hourly fixed"`
	BudgetValue    *float64   `json:"budgetValue" binding:"omitempty,gte=0"`
	Location       *string    `json:"location"`
	DurationMonths *int       `json:"durationMonths"`
	Languages      []string   `json:"languages"`
	StartDate      *time.Time `json:"startDate"`
	Tags           []string   `json:"tags"`
	Urgent         *bool      `json:"urgent"`
	Description    *string    `json:"description"`
}

// CreateApplicationRequest represents the request body for creating an application
type CreateApplicationRequest struct {
	Message      *string `json:"message"`
	CVURL        *string `json:"cv_url"`
	PortfolioURL *string `json:"portfolio_url"`
}

// OpportunityResponse represents an opportunity in API responses
type OpportunityResponse struct {
	ID             int64    `json:"id"`
	Title          string   `json:"title"`
	Organization   string   `json:"organization"`
	Position       string   `json:"position"`
	CreatorName    string   `json:"creator_name"`
	Category       string   `json:"category"`
	BudgetType     string   `json:"budget_type"`
	BudgetValue    *float64 `json:"budget_value,omitempty"`
	Location       string   `json:"location"`
	DurationMonths *int     `json:"duration_months,omitempty"`
	Languages      []string `json:"languages"`
	StartDate      *string  `json:"start_date,omitempty"`
	PublishedAt    string   `json:"published_at"`
	Urgent         bool     `json:"urgent"`
	Description    *string  `json:"description,omitempty"`
	Tags           []string `json:"tags"`
	CreatedBy      int64    `json:"created_by"`
	Status         string   `json:"status"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
	IsFavorite     *bool    `json:"is_favorite,omitempty"` // Only included when user is authenticated
}

// ApplicationResponse represents an application in API responses
type ApplicationResponse struct {
	ID            int64   `json:"id"`
	UserID        int64   `json:"user_id"`
	OpportunityID int64   `json:"opportunity_id"`
	Message       *string `json:"message,omitempty"`
	CVURL         *string `json:"cv_url,omitempty"`
	PortfolioURL  *string `json:"portfolio_url,omitempty"`
	Status        string  `json:"status"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// ListOpportunitiesResponse represents the response for listing opportunities
type ListOpportunitiesResponse struct {
	Data  []*OpportunityResponse `json:"data"`
	Page  int                    `json:"page"`
	Limit int                    `json:"limit"`
	Total int64                  `json:"total"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message"`
}
