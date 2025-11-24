package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// Project represents a portfolio project in user profile.
type Project struct {
	Title       string   `json:"title"`
	ClientName  string   `json:"client_name,omitempty"`
	Description string   `json:"description"`
	Tags        []string `json:"tags,omitempty"`
	Year        *int     `json:"year,omitempty"`
}

// Projects is a slice of Project with custom DB marshaling.
type Projects []Project

func (p Projects) Value() (driver.Value, error) {
	if len(p) == 0 {
		return []byte("[]"), nil
	}
	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (p *Projects) Scan(value interface{}) error {
	if value == nil {
		*p = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("invalid type for Projects: %T", value)
	}

	if len(bytes) == 0 {
		*p = nil
		return nil
	}

	var temp []Project
	if err := json.Unmarshal(bytes, &temp); err != nil {
		return err
	}
	*p = temp
	return nil
}

// StringArray is a JSONB-backed string slice.
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
