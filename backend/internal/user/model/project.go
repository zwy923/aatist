package model

import (
	"time"
)

// PortfolioProject represents a project in the projects table.
type PortfolioProject struct {
	ID            int64       `db:"id" json:"id"`
	UserID        int64       `db:"user_id" json:"user_id"`
	Title         string      `db:"title" json:"title"`
	Description   *string     `db:"description" json:"description,omitempty"`
	Year          *int        `db:"year" json:"year,omitempty"`
	Tags          StringArray `db:"tags" json:"tags,omitempty"`
	CoverImageURL *string     `db:"cover_image_url" json:"cover_image_url,omitempty"`
	ProjectLink   *string     `db:"project_link" json:"project_link,omitempty"`
	CreatedAt     time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time   `db:"updated_at" json:"updated_at"`
}
