package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/aatist/backend/internal/event/model"
	eventservice "github.com/aatist/backend/internal/event/service"
	"github.com/aatist/backend/internal/platform/log"
	"github.com/aatist/backend/internal/platform/middleware"
	"github.com/aatist/backend/pkg/errs"
	"github.com/aatist/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// EventHandler handles HTTP requests for events
type EventHandler struct {
	eventService    eventservice.EventService
	interestService eventservice.EventInterestService
	goingService    eventservice.EventGoingService
	commentService  eventservice.EventCommentService
	logger          *log.Logger
}

// NewEventHandler creates a new event handler
func NewEventHandler(
	eventService eventservice.EventService,
	interestService eventservice.EventInterestService,
	goingService eventservice.EventGoingService,
	commentService eventservice.EventCommentService,
	logger *log.Logger,
) *EventHandler {
	return &EventHandler{
		eventService:    eventService,
		interestService: interestService,
		goingService:    goingService,
		commentService:  commentService,
		logger:          logger,
	}
}

func (h *EventHandler) respondError(c *gin.Context, status int, err error, message string) {
	h.logger.Error("Handler error", zap.Error(err), zap.String("message", message))
	c.JSON(status, response.Error(errs.NewAppError(err, status, message)))
}

func (h *EventHandler) handleServiceError(c *gin.Context, err error) {
	if appErr, ok := err.(*errs.AppError); ok {
		h.respondError(c, appErr.StatusCode, appErr, appErr.Message)
		return
	}
	h.respondError(c, http.StatusInternalServerError, err, "internal server error")
}

// toEventResponse converts model.Event to EventResponse
func (h *EventHandler) toEventResponse(event *model.Event) *EventResponse {
	return &EventResponse{
		ID:              event.ID,
		Title:           event.Title,
		Organizer:       event.Organizer,
		TypeTags:        event.TypeTags,
		SchoolTags:      event.SchoolTags,
		IsExternal:      event.IsExternal,
		IsFree:          event.IsFree,
		Location:        event.Location,
		Languages:       event.Languages,
		StartTime:       event.StartTime.Format(time.RFC3339),
		EndTime:         event.EndTime.Format(time.RFC3339),
		MaxParticipants: event.MaxParticipants,
		Description:     event.Description,
		CreatedBy:       event.CreatedBy,
		PublishedAt:     event.PublishedAt.Format(time.RFC3339),
		Status:          event.Status.String(),
		CoverImageURL:   event.CoverImageURL,
		CreatedAt:       event.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       event.UpdatedAt.Format(time.RFC3339),
		InterestedCount: event.InterestedCount,
		GoingCount:      event.GoingCount,
		CommentCount:    event.CommentCount,
		IsInterested:    event.IsInterested,
		IsGoing:         event.IsGoing,
	}
}

// ListEventsHandler handles GET /events
func (h *EventHandler) ListEventsHandler(c *gin.Context) {
	filter := eventservice.ListEventsFilter{}

	// Parse query parameters
	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	if sort := c.Query("sort"); sort != "" {
		filter.Sort = sort
	} else {
		filter.Sort = "new"
	}

	if timeFilter := c.Query("time"); timeFilter != "" {
		filter.TimeFilter = &timeFilter
	}

	if types := c.QueryArray("types"); len(types) > 0 {
		filter.Types = types
	}

	if schools := c.QueryArray("schools"); len(schools) > 0 {
		filter.Schools = schools
	}

	if freeStr := c.Query("free"); freeStr != "" {
		if free, err := strconv.ParseBool(freeStr); err == nil {
			filter.IsFree = &free
		}
	}

	if languages := c.QueryArray("languages"); len(languages) > 0 {
		filter.Languages = languages
	}

	if location := c.Query("location"); location != "" {
		filter.Location = &location
	}

	if status := c.Query("status"); status != "" {
		filter.Status = &status
	}

	// Pagination
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		} else {
			filter.Page = 1
		}
	} else {
		filter.Page = 1
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		} else {
			filter.Limit = 20
		}
	} else {
		filter.Limit = 20
	}

	// Get user ID if authenticated
	var userID *int64
	if id, err := middleware.GetUserID(c); err == nil {
		userID = &id
	}

	result, err := h.eventService.List(c.Request.Context(), filter, userID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	responses := make([]*EventResponse, len(result.Data))
	for i, event := range result.Data {
		responses[i] = h.toEventResponse(event)
	}

	c.JSON(http.StatusOK, response.Success(&ListEventsResponse{
		Data:  responses,
		Page:  result.Page,
		Limit: result.Limit,
		Total: result.Total,
	}))
}

// GetEventHandler handles GET /events/:id
func (h *EventHandler) GetEventHandler(c *gin.Context) {
	var req struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid event id")
		return
	}

	// Get user ID if authenticated
	var userID *int64
	if id, err := middleware.GetUserID(c); err == nil {
		userID = &id
	}

	event, err := h.eventService.GetByID(c.Request.Context(), req.ID, userID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(h.toEventResponse(event)))
}

// CreateEventHandler handles POST /events
func (h *EventHandler) CreateEventHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	var req CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	input := eventservice.CreateEventInput{
		Title:           req.Title,
		Organizer:       req.Organizer,
		TypeTags:        req.TypeTags,
		SchoolTags:      req.SchoolTags,
		IsExternal:      req.IsExternal,
		IsFree:          req.IsFree,
		Location:        req.Location,
		Languages:       req.Languages,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		MaxParticipants: req.MaxParticipants,
		Description:     req.Description,
		CoverImageURL:   req.CoverImageURL,
		CreatedBy:       userID,
	}

	event, err := h.eventService.Create(c.Request.Context(), input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.Success(h.toEventResponse(event)))
}

// UpdateEventHandler handles PATCH /events/:id
func (h *EventHandler) UpdateEventHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	var uriReq struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uriReq); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid event id")
		return
	}

	var req UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	input := eventservice.UpdateEventInput{
		Title:           req.Title,
		Organizer:       req.Organizer,
		TypeTags:        req.TypeTags,
		SchoolTags:      req.SchoolTags,
		IsExternal:      req.IsExternal,
		IsFree:          req.IsFree,
		Location:        req.Location,
		Languages:       req.Languages,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		MaxParticipants: req.MaxParticipants,
		Description:     req.Description,
		CoverImageURL:   req.CoverImageURL,
	}

	event, err := h.eventService.Update(c.Request.Context(), uriReq.ID, userID, input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(h.toEventResponse(event)))
}

// DeleteEventHandler handles DELETE /events/:id
func (h *EventHandler) DeleteEventHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	var req struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid event id")
		return
	}

	if err := h.eventService.Delete(c.Request.Context(), req.ID, userID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "event deleted successfully"}))
}

// AddInterestHandler handles POST /events/:id/interested
func (h *EventHandler) AddInterestHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	var req struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid event id")
		return
	}

	if err := h.interestService.AddInterest(c.Request.Context(), userID, req.ID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "interest added successfully"}))
}

// RemoveInterestHandler handles DELETE /events/:id/interested
func (h *EventHandler) RemoveInterestHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	var req struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid event id")
		return
	}

	if err := h.interestService.RemoveInterest(c.Request.Context(), userID, req.ID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "interest removed successfully"}))
}

// AddGoingHandler handles POST /events/:id/going
func (h *EventHandler) AddGoingHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	var req struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid event id")
		return
	}

	if err := h.goingService.AddGoing(c.Request.Context(), userID, req.ID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "going added successfully"}))
}

// RemoveGoingHandler handles DELETE /events/:id/going
func (h *EventHandler) RemoveGoingHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	var req struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid event id")
		return
	}

	if err := h.goingService.RemoveGoing(c.Request.Context(), userID, req.ID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "going removed successfully"}))
}

// CreateCommentHandler handles POST /events/:id/comments
func (h *EventHandler) CreateCommentHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	var uriReq struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uriReq); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid event id")
		return
	}

	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	input := eventservice.CreateCommentInput{
		EventID: uriReq.ID,
		Content: req.Content,
	}

	comment, err := h.commentService.CreateComment(c.Request.Context(), userID, input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	resp := &CommentResponse{
		ID:        comment.ID,
		EventID:   comment.EventID,
		UserID:    comment.UserID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusCreated, response.Success(resp))
}

// ListCommentsHandler handles GET /events/:id/comments
func (h *EventHandler) ListCommentsHandler(c *gin.Context) {
	var uriReq struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uriReq); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid event id")
		return
	}

	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	comments, err := h.commentService.ListComments(c.Request.Context(), uriReq.ID, page, limit)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	responses := make([]*CommentResponse, len(comments))
	for i, comment := range comments {
		responses[i] = &CommentResponse{
			ID:        comment.ID,
			EventID:   comment.EventID,
			UserID:    comment.UserID,
			Content:   comment.Content,
			CreatedAt: comment.CreatedAt.Format(time.RFC3339),
		}
	}

	c.JSON(http.StatusOK, response.Success(responses))
}

// DeleteCommentHandler handles DELETE /events/:id/comments/:commentId
func (h *EventHandler) DeleteCommentHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	var req struct {
		CommentID int64 `uri:"commentId" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid comment id")
		return
	}

	if err := h.commentService.DeleteComment(c.Request.Context(), req.CommentID, userID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "comment deleted successfully"}))
}
