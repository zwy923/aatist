package repository

import (
	"context"

	"github.com/aatist/backend/internal/event/model"
)

// EventListSort represents sort options for listing events
type EventListSort string

const (
	SortNewEvents    EventListSort = "new"        // published_at desc
	SortStartingSoon EventListSort = "start_time" // start_time asc
	SortPopular      EventListSort = "popular"    // interested_count desc
)

// TimeFilter represents time-based filters
type TimeFilter string

const (
	TimeFilterToday     TimeFilter = "today"
	TimeFilterThisWeek  TimeFilter = "this_week"
	TimeFilterThisMonth TimeFilter = "this_month"
)

// EventListFilter represents filters for listing events
type EventListFilter struct {
	Search     *string
	Sort       EventListSort
	TimeFilter *TimeFilter
	Types      []string
	Schools    []string
	IsFree     *bool
	Languages  []string
	Location   *string
	Status     *string
	Page       int
	Limit      int
}

// EventRepository defines operations on events
type EventRepository interface {
	// Create creates a new event
	Create(ctx context.Context, event *model.Event) error

	// Update updates an event
	Update(ctx context.Context, event *model.Event) error

	// FindByID finds an event by ID
	FindByID(ctx context.Context, id int64) (*model.Event, error)

	// List lists events with filters, sorting, and pagination
	List(ctx context.Context, filter EventListFilter) ([]*model.Event, error)

	// Count counts events matching the filter (for pagination)
	Count(ctx context.Context, filter EventListFilter) (int64, error)

	// Delete deletes an event (soft delete by setting status to canceled)
	Delete(ctx context.Context, id int64, userID int64) error

	// ListByUserID lists events created by a user
	ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]*model.Event, error)
}

// EventInterestRepository defines operations on event interests
type EventInterestRepository interface {
	// Create creates an interest relationship
	Create(ctx context.Context, userID, eventID int64) error

	// Delete deletes an interest relationship
	Delete(ctx context.Context, userID, eventID int64) error

	// Exists checks if an interest relationship exists
	Exists(ctx context.Context, userID, eventID int64) (bool, error)

	// CountByEventID counts interests for an event
	CountByEventID(ctx context.Context, eventID int64) (int64, error)
}

// EventGoingRepository defines operations on event going
type EventGoingRepository interface {
	// Create creates a going relationship
	Create(ctx context.Context, userID, eventID int64) error

	// Delete deletes a going relationship
	Delete(ctx context.Context, userID, eventID int64) error

	// Exists checks if a going relationship exists
	Exists(ctx context.Context, userID, eventID int64) (bool, error)

	// CountByEventID counts going for an event
	CountByEventID(ctx context.Context, eventID int64) (int64, error)
}

// EventCommentRepository defines operations on event comments
type EventCommentRepository interface {
	// Create creates a new comment
	Create(ctx context.Context, comment *model.EventComment) error

	// FindByID finds a comment by ID
	FindByID(ctx context.Context, id int64) (*model.EventComment, error)

	// ListByEventID lists all comments for an event
	ListByEventID(ctx context.Context, eventID int64, limit, offset int) ([]*model.EventComment, error)

	// Delete deletes a comment
	Delete(ctx context.Context, id int64, userID int64) error

	// CountByEventID counts comments for an event
	CountByEventID(ctx context.Context, eventID int64) (int64, error)
}
