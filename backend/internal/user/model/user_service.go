package model

import "time"

// UserService represents a user's service offering
type UserService struct {
	ID                int64     `db:"id" json:"id"`
	UserID            int64     `db:"user_id" json:"user_id"`
	Category          string    `db:"category" json:"category"`
	ExperienceSummary string    `db:"experience_summary" json:"experience_summary"` // legacy, kept for compatibility
	Title             string    `db:"title" json:"title,omitempty"`
	Description       string    `db:"description" json:"description,omitempty"`
	ShortDescription  string    `db:"short_description" json:"short_description,omitempty"`
	PriceType         string    `db:"price_type" json:"price_type,omitempty"` // hourly, project, negotiable
	PriceMin          *int     `db:"price_min" json:"price_min,omitempty"`
	PriceMax          *int     `db:"price_max" json:"price_max,omitempty"`
	MediaURLs         StringArray `db:"media_urls" json:"media_urls,omitempty"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at"`
}
