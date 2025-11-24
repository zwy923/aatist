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

// Skill represents a skill with proficiency level.
type Skill struct {
	Name  string `json:"name"`
	Level string `json:"level"` // Expert / Advanced / Intermediate
}

// Skills is a JSONB-backed slice of Skill.
type Skills []Skill

func (s Skills) Value() (driver.Value, error) {
	if len(s) == 0 {
		return []byte("[]"), nil
	}
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (s *Skills) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("invalid type for Skills: %T", value)
	}

	if len(bytes) == 0 {
		*s = nil
		return nil
	}

	var temp []Skill
	if err := json.Unmarshal(bytes, &temp); err != nil {
		return err
	}
	*s = temp
	return nil
}

// WeeklyAvailability represents availability for a specific week.
type WeeklyAvailability struct {
	Week   int    `json:"week"`
	Year   int    `json:"year"`
	Status string `json:"status"` // open / busy / limited
}

// WeeklyAvailabilityArray is a JSONB-backed slice of WeeklyAvailability.
type WeeklyAvailabilityArray []WeeklyAvailability

func (wa WeeklyAvailabilityArray) Value() (driver.Value, error) {
	if len(wa) == 0 {
		return []byte("[]"), nil
	}
	b, err := json.Marshal(wa)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (wa *WeeklyAvailabilityArray) Scan(value interface{}) error {
	if value == nil {
		*wa = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("invalid type for WeeklyAvailabilityArray: %T", value)
	}

	if len(bytes) == 0 {
		*wa = nil
		return nil
	}

	var temp []WeeklyAvailability
	if err := json.Unmarshal(bytes, &temp); err != nil {
		return err
	}
	*wa = temp
	return nil
}

// StringArray is a JSONB-backed string slice (kept for backward compatibility).
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
