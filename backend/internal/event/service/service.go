package service

import (
	"context"
	"fmt"
	"time"

	"github.com/aatist/backend/internal/event/model"
	"github.com/aatist/backend/internal/event/repository"
	"github.com/aatist/backend/pkg/errs"
)

// EventService defines the interface for event operations
type EventService interface {
	// Create creates a new event
	Create(ctx context.Context, input CreateEventInput) (*model.Event, error)

	// Update updates an event (only by creator)
	Update(ctx context.Context, eventID, userID int64, input UpdateEventInput) (*model.Event, error)

	// GetByID gets an event by ID
	GetByID(ctx context.Context, id int64, userID *int64) (*model.Event, error)

	// List lists events with filters, sorting, and pagination
	List(ctx context.Context, filter ListEventsFilter, userID *int64) (*ListEventsResult, error)

	// Delete deletes an event (soft delete)
	Delete(ctx context.Context, eventID, userID int64) error

	// ListByUserID lists events created by a user
	ListByUserID(ctx context.Context, userID int64, page, limit int) ([]*model.Event, error)
}

// EventInterestService defines the interface for interest operations
type EventInterestService interface {
	// AddInterest adds an interest relationship
	AddInterest(ctx context.Context, userID, eventID int64) error

	// RemoveInterest removes an interest relationship
	RemoveInterest(ctx context.Context, userID, eventID int64) error

	// IsInterested checks if a user is interested in an event
	IsInterested(ctx context.Context, userID, eventID int64) (bool, error)
}

// EventGoingService defines the interface for going operations
type EventGoingService interface {
	// AddGoing adds a going relationship
	AddGoing(ctx context.Context, userID, eventID int64) error

	// RemoveGoing removes a going relationship
	RemoveGoing(ctx context.Context, userID, eventID int64) error

	// IsGoing checks if a user is going to an event
	IsGoing(ctx context.Context, userID, eventID int64) (bool, error)
}

// EventCommentService defines the interface for comment operations
type EventCommentService interface {
	// CreateComment creates a new comment
	CreateComment(ctx context.Context, userID int64, input CreateCommentInput) (*model.EventComment, error)

	// GetComment gets a comment by ID
	GetComment(ctx context.Context, id int64) (*model.EventComment, error)

	// ListComments lists all comments for an event
	ListComments(ctx context.Context, eventID int64, page, limit int) ([]*model.EventComment, error)

	// DeleteComment deletes a comment
	DeleteComment(ctx context.Context, id, userID int64) error
}

// CreateEventInput represents input for creating an event
type CreateEventInput struct {
	Title           string
	Organizer       string
	TypeTags        []string
	SchoolTags      []string
	IsExternal      bool
	IsFree          bool
	Location        string
	Languages       []string
	StartTime       time.Time
	EndTime         time.Time
	MaxParticipants *int
	Description     *string
	CoverImageURL   *string
	CreatedBy       int64
}

// UpdateEventInput represents input for updating an event
type UpdateEventInput struct {
	Title           *string
	Organizer       *string
	TypeTags        []string
	SchoolTags      []string
	IsExternal      *bool
	IsFree          *bool
	Location        *string
	Languages       []string
	StartTime       *time.Time
	EndTime         *time.Time
	MaxParticipants *int
	Description     *string
	CoverImageURL   *string
}

// ListEventsFilter represents filters for listing events
type ListEventsFilter struct {
	Search     *string
	Sort       string
	TimeFilter *string
	Types      []string
	Schools    []string
	IsFree     *bool
	Languages  []string
	Location   *string
	Status     *string
	Page       int
	Limit      int
}

// ListEventsResult represents the result of listing events
type ListEventsResult struct {
	Data  []*model.Event `json:"data"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
	Total int64          `json:"total"`
}

// CreateCommentInput represents input for creating a comment
type CreateCommentInput struct {
	EventID int64
	Content string
}

// eventService implements EventService
type eventService struct {
	eventRepo    repository.EventRepository
	interestRepo repository.EventInterestRepository
	goingRepo    repository.EventGoingRepository
	commentRepo  repository.EventCommentRepository
}

// NewEventService creates a new event service
func NewEventService(
	eventRepo repository.EventRepository,
	interestRepo repository.EventInterestRepository,
	goingRepo repository.EventGoingRepository,
	commentRepo repository.EventCommentRepository,
) EventService {
	return &eventService{
		eventRepo:    eventRepo,
		interestRepo: interestRepo,
		goingRepo:    goingRepo,
		commentRepo:  commentRepo,
	}
}

func (s *eventService) Create(ctx context.Context, input CreateEventInput) (*model.Event, error) {
	if input.EndTime.Before(input.StartTime) {
		return nil, errs.NewAppError(errs.ErrInvalidInput, 400, "end_time must be after start_time").WithCode(errs.CodeInvalidInput)
	}

	event := &model.Event{
		Title:           input.Title,
		Organizer:       input.Organizer,
		TypeTags:        input.TypeTags,
		SchoolTags:      input.SchoolTags,
		IsExternal:      input.IsExternal,
		IsFree:          input.IsFree,
		Location:        input.Location,
		Languages:       input.Languages,
		StartTime:       input.StartTime,
		EndTime:         input.EndTime,
		MaxParticipants: input.MaxParticipants,
		Description:     input.Description,
		CoverImageURL:   input.CoverImageURL,
		CreatedBy:       input.CreatedBy,
		Status:          model.EventStatusActive,
	}

	if err := s.eventRepo.Create(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	return event, nil
}

func (s *eventService) Update(ctx context.Context, eventID, userID int64, input UpdateEventInput) (*model.Event, error) {
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	if event.CreatedBy != userID {
		return nil, errs.NewAppError(errs.ErrForbidden, 403, "only creator can update event").WithCode(errs.CodeForbidden)
	}

	// Update fields if provided
	if input.Title != nil {
		event.Title = *input.Title
	}
	if input.Organizer != nil {
		event.Organizer = *input.Organizer
	}
	if input.TypeTags != nil {
		event.TypeTags = input.TypeTags
	}
	if input.SchoolTags != nil {
		event.SchoolTags = input.SchoolTags
	}
	if input.IsExternal != nil {
		event.IsExternal = *input.IsExternal
	}
	if input.IsFree != nil {
		event.IsFree = *input.IsFree
	}
	if input.Location != nil {
		event.Location = *input.Location
	}
	if input.Languages != nil {
		event.Languages = input.Languages
	}
	if input.StartTime != nil {
		event.StartTime = *input.StartTime
	}
	if input.EndTime != nil {
		event.EndTime = *input.EndTime
	}
	if input.MaxParticipants != nil {
		event.MaxParticipants = input.MaxParticipants
	}
	if input.Description != nil {
		event.Description = input.Description
	}
	if input.CoverImageURL != nil {
		event.CoverImageURL = input.CoverImageURL
	}

	if event.EndTime.Before(event.StartTime) {
		return nil, errs.NewAppError(errs.ErrInvalidInput, 400, "end_time must be after start_time").WithCode(errs.CodeInvalidInput)
	}

	if err := s.eventRepo.Update(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to update event: %w", err)
	}

	return event, nil
}

func (s *eventService) GetByID(ctx context.Context, id int64, userID *int64) (*model.Event, error) {
	event, err := s.eventRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get counts
	interestedCount, _ := s.interestRepo.CountByEventID(ctx, id)
	goingCount, _ := s.goingRepo.CountByEventID(ctx, id)
	commentCount, _ := s.commentRepo.CountByEventID(ctx, id)

	event.InterestedCount = &interestedCount
	event.GoingCount = &goingCount
	event.CommentCount = &commentCount

	// Check user's status if authenticated
	if userID != nil {
		isInterested, _ := s.interestRepo.Exists(ctx, *userID, id)
		isGoing, _ := s.goingRepo.Exists(ctx, *userID, id)
		event.IsInterested = &isInterested
		event.IsGoing = &isGoing
	}

	return event, nil
}

func (s *eventService) List(ctx context.Context, filter ListEventsFilter, userID *int64) (*ListEventsResult, error) {
	// Convert sort string to repository sort type
	var sort repository.EventListSort
	switch filter.Sort {
	case "start_time":
		sort = repository.SortStartingSoon
	case "popular":
		sort = repository.SortPopular
	default:
		sort = repository.SortNewEvents
	}

	// Convert time filter
	var timeFilter *repository.TimeFilter
	if filter.TimeFilter != nil {
		tf := repository.TimeFilter(*filter.TimeFilter)
		timeFilter = &tf
	}

	repoFilter := repository.EventListFilter{
		Search:     filter.Search,
		Sort:       sort,
		TimeFilter: timeFilter,
		Types:      filter.Types,
		Schools:    filter.Schools,
		IsFree:     filter.IsFree,
		Languages:  filter.Languages,
		Location:   filter.Location,
		Status:     filter.Status,
		Page:       filter.Page,
		Limit:      filter.Limit,
	}

	events, err := s.eventRepo.List(ctx, repoFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	// Check user's status for each event if authenticated
	if userID != nil {
		for _, event := range events {
			isInterested, _ := s.interestRepo.Exists(ctx, *userID, event.ID)
			isGoing, _ := s.goingRepo.Exists(ctx, *userID, event.ID)
			event.IsInterested = &isInterested
			event.IsGoing = &isGoing
		}
	}

	total, err := s.eventRepo.Count(ctx, repoFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to count events: %w", err)
	}

	return &ListEventsResult{
		Data:  events,
		Page:  filter.Page,
		Limit: filter.Limit,
		Total: total,
	}, nil
}

func (s *eventService) Delete(ctx context.Context, eventID, userID int64) error {
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return err
	}

	if event.CreatedBy != userID {
		return errs.NewAppError(errs.ErrForbidden, 403, "only creator can delete event").WithCode(errs.CodeForbidden)
	}

	return s.eventRepo.Delete(ctx, eventID, userID)
}

func (s *eventService) ListByUserID(ctx context.Context, userID int64, page, limit int) ([]*model.Event, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	return s.eventRepo.ListByUserID(ctx, userID, limit, offset)
}

// eventInterestService implements EventInterestService
type eventInterestService struct {
	interestRepo repository.EventInterestRepository
}

// NewEventInterestService creates a new event interest service
func NewEventInterestService(interestRepo repository.EventInterestRepository) EventInterestService {
	return &eventInterestService{interestRepo: interestRepo}
}

func (s *eventInterestService) AddInterest(ctx context.Context, userID, eventID int64) error {
	if err := s.interestRepo.Create(ctx, userID, eventID); err != nil {
		return fmt.Errorf("failed to add interest: %w", err)
	}
	return nil
}

func (s *eventInterestService) RemoveInterest(ctx context.Context, userID, eventID int64) error {
	if err := s.interestRepo.Delete(ctx, userID, eventID); err != nil {
		if err == errs.ErrNotFound {
			return nil // Not interested, no error
		}
		return fmt.Errorf("failed to remove interest: %w", err)
	}
	return nil
}

func (s *eventInterestService) IsInterested(ctx context.Context, userID, eventID int64) (bool, error) {
	return s.interestRepo.Exists(ctx, userID, eventID)
}

// eventGoingService implements EventGoingService
type eventGoingService struct {
	goingRepo repository.EventGoingRepository
}

// NewEventGoingService creates a new event going service
func NewEventGoingService(goingRepo repository.EventGoingRepository) EventGoingService {
	return &eventGoingService{goingRepo: goingRepo}
}

func (s *eventGoingService) AddGoing(ctx context.Context, userID, eventID int64) error {
	if err := s.goingRepo.Create(ctx, userID, eventID); err != nil {
		return fmt.Errorf("failed to add going: %w", err)
	}
	return nil
}

func (s *eventGoingService) RemoveGoing(ctx context.Context, userID, eventID int64) error {
	if err := s.goingRepo.Delete(ctx, userID, eventID); err != nil {
		if err == errs.ErrNotFound {
			return nil // Not going, no error
		}
		return fmt.Errorf("failed to remove going: %w", err)
	}
	return nil
}

func (s *eventGoingService) IsGoing(ctx context.Context, userID, eventID int64) (bool, error) {
	return s.goingRepo.Exists(ctx, userID, eventID)
}

// eventCommentService implements EventCommentService
type eventCommentService struct {
	commentRepo repository.EventCommentRepository
}

// NewEventCommentService creates a new event comment service
func NewEventCommentService(commentRepo repository.EventCommentRepository) EventCommentService {
	return &eventCommentService{commentRepo: commentRepo}
}

func (s *eventCommentService) CreateComment(ctx context.Context, userID int64, input CreateCommentInput) (*model.EventComment, error) {
	comment := &model.EventComment{
		EventID: input.EventID,
		UserID:  userID,
		Content: input.Content,
	}

	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	return comment, nil
}

func (s *eventCommentService) GetComment(ctx context.Context, id int64) (*model.EventComment, error) {
	return s.commentRepo.FindByID(ctx, id)
}

func (s *eventCommentService) ListComments(ctx context.Context, eventID int64, page, limit int) ([]*model.EventComment, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	return s.commentRepo.ListByEventID(ctx, eventID, limit, offset)
}

func (s *eventCommentService) DeleteComment(ctx context.Context, id, userID int64) error {
	return s.commentRepo.Delete(ctx, id, userID)
}
