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

// Project represents a project in the projects table.
type Project struct {
	ID              int64       `db:"id" json:"id"`
	UserID          int64       `db:"user_id" json:"user_id"`
	Title           string      `db:"title" json:"title"`
	ServiceCategory *string     `db:"service_category" json:"service_category,omitempty"`
	ClientName      *string     `db:"client_name" json:"client_name,omitempty"`
	Description     *string     `db:"description" json:"description,omitempty"`
	Year            *int        `db:"year" json:"year,omitempty"`
	Tags            StringArray `db:"tags" json:"tags,omitempty"`
	CoverImageURL   *string     `db:"cover_image_url" json:"cover_image_url,omitempty"`
	ProjectLink     *string     `db:"project_link" json:"project_link,omitempty"`
	CreatedAt       time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time   `db:"updated_at" json:"updated_at"`
}
