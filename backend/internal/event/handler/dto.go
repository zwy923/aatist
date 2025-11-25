package handler

import (
	"time"
)

// CreateEventRequest represents the request body for creating an event
type CreateEventRequest struct {
	Title           string     `json:"title" binding:"required"`
	Organizer       string     `json:"organizer" binding:"required"`
	TypeTags        []string   `json:"type_tags" binding:"required"`
	SchoolTags      []string   `json:"school_tags" binding:"required"`
	IsExternal      bool       `json:"is_external"`
	IsFree          bool       `json:"is_free"`
	Location        string     `json:"location" binding:"required"`
	Languages       []string   `json:"languages" binding:"required"`
	StartTime       time.Time  `json:"start_time" binding:"required"`
	EndTime         time.Time  `json:"end_time" binding:"required"`
	MaxParticipants *int       `json:"max_participants"`
	Description     *string    `json:"description"`
	CoverImageURL   *string    `json:"cover_image_url"`
}

// UpdateEventRequest represents the request body for updating an event
type UpdateEventRequest struct {
	Title           *string    `json:"title"`
	Organizer       *string    `json:"organizer"`
	TypeTags        []string   `json:"type_tags"`
	SchoolTags      []string   `json:"school_tags"`
	IsExternal      *bool      `json:"is_external"`
	IsFree          *bool      `json:"is_free"`
	Location        *string    `json:"location"`
	Languages       []string   `json:"languages"`
	StartTime       *time.Time `json:"start_time"`
	EndTime         *time.Time `json:"end_time"`
	MaxParticipants *int       `json:"max_participants"`
	Description     *string    `json:"description"`
	CoverImageURL   *string    `json:"cover_image_url"`
}

// CreateCommentRequest represents the request body for creating a comment
type CreateCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

// EventResponse represents an event in API responses
type EventResponse struct {
	ID              int64     `json:"id"`
	Title           string    `json:"title"`
	Organizer       string    `json:"organizer"`
	TypeTags        []string  `json:"type_tags"`
	SchoolTags      []string  `json:"school_tags"`
	IsExternal      bool      `json:"is_external"`
	IsFree          bool      `json:"is_free"`
	Location        string    `json:"location"`
	Languages       []string  `json:"languages"`
	StartTime       string    `json:"start_time"`
	EndTime         string    `json:"end_time"`
	MaxParticipants *int      `json:"max_participants,omitempty"`
	Description     *string   `json:"description,omitempty"`
	CreatedBy       int64     `json:"created_by"`
	PublishedAt     string    `json:"published_at"`
	Status          string    `json:"status"`
	CoverImageURL   *string   `json:"cover_image_url,omitempty"`
	CreatedAt       string    `json:"created_at"`
	UpdatedAt       string    `json:"updated_at"`
	InterestedCount *int64    `json:"interested_count,omitempty"`
	GoingCount      *int64    `json:"going_count,omitempty"`
	CommentCount    *int64    `json:"comment_count,omitempty"`
	IsInterested    *bool     `json:"is_interested,omitempty"`
	IsGoing         *bool     `json:"is_going,omitempty"`
}

// CommentResponse represents a comment in API responses
type CommentResponse struct {
	ID        int64  `json:"id"`
	EventID   int64  `json:"event_id"`
	UserID    int64  `json:"user_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// ListEventsResponse represents the response for listing events
type ListEventsResponse struct {
	Data  []*EventResponse `json:"data"`
	Page  int              `json:"page"`
	Limit int              `json:"limit"`
	Total int64            `json:"total"`
}

