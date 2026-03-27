package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// StringArray is a JSONB-backed slice of strings.
type StringArray []string

func (sa StringArray) Value() (driver.Value, error) {
	if len(sa) == 0 {
		return []byte("[]"), nil
	}
	b, err := json.Marshal(sa)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (sa *StringArray) Scan(value interface{}) error {
	if value == nil {
		*sa = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("invalid type for StringArray: %T", value)
	}

	if len(bytes) == 0 {
		*sa = nil
		return nil
	}

	var temp []string
	if err := json.Unmarshal(bytes, &temp); err != nil {
		return err
	}
	*sa = temp
	return nil
}

// PortfolioCoCreator is stored in projects.co_creators JSONB.
type PortfolioCoCreator struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// CoCreatorsArray is JSONB-backed []PortfolioCoCreator.
type CoCreatorsArray []PortfolioCoCreator

func (c CoCreatorsArray) Value() (driver.Value, error) {
	if len(c) == 0 {
		return []byte("[]"), nil
	}
	b, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (c *CoCreatorsArray) Scan(value interface{}) error {
	if value == nil {
		*c = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("invalid type for CoCreatorsArray: %T", value)
	}
	if len(bytes) == 0 {
		*c = nil
		return nil
	}
	var temp []PortfolioCoCreator
	if err := json.Unmarshal(bytes, &temp); err != nil {
		return err
	}
	*c = temp
	return nil
}

// Project represents a project in the projects table.
type Project struct {
	ID               int64           `db:"id" json:"id"`
	UserID           int64           `db:"user_id" json:"user_id"`
	Title            string          `db:"title" json:"title"`
	ShortCaption     *string         `db:"short_caption" json:"short_caption,omitempty"`
	ServiceCategory  *string         `db:"service_category" json:"service_category,omitempty"`
	ClientName       *string         `db:"client_name" json:"client_name,omitempty"`
	Description      *string         `db:"description" json:"description,omitempty"`
	Year             *int            `db:"year" json:"year,omitempty"`
	Tags             StringArray     `db:"tags" json:"tags,omitempty"`
	MediaURLs        StringArray     `db:"media_urls" json:"media_urls,omitempty"`
	RelatedServices  StringArray     `db:"related_services" json:"related_services,omitempty"`
	CoCreators       CoCreatorsArray `db:"co_creators" json:"co_creators,omitempty"`
	CoverImageURL    *string         `db:"cover_image_url" json:"cover_image_url,omitempty"`
	ProjectLink      *string         `db:"project_link" json:"project_link,omitempty"`
	IsPublished      bool            `db:"is_published" json:"is_published"`
	IsPublic         bool            `db:"is_public" json:"is_public"`
	CreatedAt        time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time       `db:"updated_at" json:"updated_at"`
}
