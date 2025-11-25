package model

import (
	"time"
)

// EventStatus represents the status of an event
type EventStatus string

const (
	EventStatusActive   EventStatus = "active"
	EventStatusCanceled EventStatus = "canceled"
)

// IsValid checks if status is valid
func (s EventStatus) IsValid() bool {
	return s == EventStatusActive || s == EventStatusCanceled
}

// String returns string representation of status
func (s EventStatus) String() string {
	return string(s)
}

// Event represents an event in the system
type Event struct {
	ID              int64      `db:"id" json:"id"`
	Title           string     `db:"title" json:"title"`
	Organizer       string     `db:"organizer" json:"organizer"`
	TypeTags        []string   `db:"type_tags" json:"type_tags"`
	SchoolTags      []string   `db:"school_tags" json:"school_tags"`
	IsExternal      bool       `db:"is_external" json:"is_external"`
	IsFree          bool       `db:"is_free" json:"is_free"`
	Location        string     `db:"location" json:"location"`
	Languages       []string   `db:"languages" json:"languages"`
	StartTime       time.Time  `db:"start_time" json:"start_time"`
	EndTime         time.Time  `db:"end_time" json:"end_time"`
	MaxParticipants *int       `db:"max_participants" json:"max_participants,omitempty"`
	Description     *string    `db:"description" json:"description,omitempty"`
	CreatedBy       int64      `db:"created_by" json:"created_by"`
	PublishedAt     time.Time  `db:"published_at" json:"published_at"`
	Status          EventStatus `db:"status" json:"status"`
	CoverImageURL   *string    `db:"cover_image_url" json:"cover_image_url,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
	
	// Computed fields (not in DB, populated by queries)
	InterestedCount *int64 `json:"interested_count,omitempty"`
	GoingCount      *int64 `json:"going_count,omitempty"`
	CommentCount    *int64 `json:"comment_count,omitempty"`
	IsInterested    *bool  `json:"is_interested,omitempty"` // For authenticated users
	IsGoing         *bool  `json:"is_going,omitempty"`      // For authenticated users
}

// EventInterest represents a user's interest in an event
type EventInterest struct {
	UserID    int64     `db:"user_id" json:"user_id"`
	EventID   int64     `db:"event_id" json:"event_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// EventGoing represents a user's going status for an event
type EventGoing struct {
	UserID    int64     `db:"user_id" json:"user_id"`
	EventID   int64     `db:"event_id" json:"event_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// EventComment represents a comment on an event
type EventComment struct {
	ID        int64     `db:"id" json:"id"`
	EventID   int64     `db:"event_id" json:"event_id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	Content   string    `db:"content" json:"content"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

