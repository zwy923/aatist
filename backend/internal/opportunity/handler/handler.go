package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/aatist/backend/internal/opportunity/model"
	oppservice "github.com/aatist/backend/internal/opportunity/service"
	"github.com/aatist/backend/internal/platform/log"
	"github.com/aatist/backend/internal/platform/middleware"
	userservice "github.com/aatist/backend/internal/user/service"
	"github.com/aatist/backend/pkg/errs"
	"github.com/aatist/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// OpportunityHandler handles HTTP requests for opportunities
type OpportunityHandler struct {
	oppService         oppservice.OpportunityService
	savedItemClient    userservice.SavedItemClient
	applicationService oppservice.ApplicationService
	logger             *log.Logger
}

// NewOpportunityHandler creates a new opportunity handler
func NewOpportunityHandler(
	oppService oppservice.OpportunityService,
	savedItemClient userservice.SavedItemClient,
	applicationService oppservice.ApplicationService,
	logger *log.Logger,
) *OpportunityHandler {
	return &OpportunityHandler{
		oppService:         oppService,
		savedItemClient:    savedItemClient,
		applicationService: applicationService,
		logger:             logger,
	}
}

func (h *OpportunityHandler) respondError(c *gin.Context, status int, err error, message string) {
	h.logger.Error("Handler error", zap.Error(err), zap.String("message", message))
	c.JSON(status, response.Error(errs.NewAppError(err, status, message)))
}

func (h *OpportunityHandler) handleServiceError(c *gin.Context, err error) {
	if appErr, ok := err.(*errs.AppError); ok {
		h.respondError(c, appErr.StatusCode, appErr, appErr.Message)
		return
	}
	h.respondError(c, http.StatusInternalServerError, err, "internal server error")
}

// parsePaginationParams parses page and limit query parameters with defaults
func (h *OpportunityHandler) parsePaginationParams(c *gin.Context) (page, limit int) {
	page = 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit = 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	return page, limit
}

// parseDateParam parses a date string from query parameter
// Returns nil if the parameter is empty or invalid (with error logged)
func (h *OpportunityHandler) parseDateParam(c *gin.Context, paramName string) *time.Time {
	dateStr := c.Query(paramName)
	if dateStr == "" {
		return nil
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		h.logger.Warn("Invalid date parameter",
			zap.String("param", paramName),
			zap.String("value", dateStr),
			zap.Error(err),
		)
		return nil
	}

	return &date
}

// checkOwnership checks if the user is the owner of the opportunity
// Returns the opportunity if found, or an error if not found or not owner
func (h *OpportunityHandler) checkOwnership(c *gin.Context, opportunityID, userID int64) (*model.Opportunity, error) {
	opp, err := h.oppService.GetByID(c.Request.Context(), opportunityID)
	if err != nil {
		return nil, err
	}

	if opp.CreatedBy != userID {
		return nil, errs.NewAppError(errs.ErrForbidden, http.StatusForbidden, "only creator can perform this action").WithCode(errs.CodeForbidden)
	}

	return opp, nil
}

// toOpportunityResponse converts model.Opportunity to OpportunityResponse
func (h *OpportunityHandler) toOpportunityResponse(opp *model.Opportunity, isFavorite *bool) *OpportunityResponse {
	resp := &OpportunityResponse{
		ID:             opp.ID,
		Title:          opp.Title,
		Organization:   opp.Organization,
		Category:       opp.Category,
		BudgetType:     opp.BudgetType.String(),
		BudgetValue:    opp.BudgetValue,
		Location:       opp.Location,
		DurationMonths: opp.DurationMonths,
		Languages:      opp.Languages,
		Urgent:         opp.Urgent,
		Description:    opp.Description,
		Tags:           opp.Tags,
		CreatedBy:      opp.CreatedBy,
		Status:         opp.Status.String(),
		PublishedAt:    opp.PublishedAt.Format(time.RFC3339),
		CreatedAt:      opp.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      opp.UpdatedAt.Format(time.RFC3339),
		IsFavorite:     isFavorite, // Keep field name for backward compatibility, but it represents saved status
	}

	if opp.StartDate != nil {
		dateStr := opp.StartDate.Format("2006-01-02")
		resp.StartDate = &dateStr
	}

	return resp
}

// ListOpportunitiesHandler handles GET /opportunities
func (h *OpportunityHandler) ListOpportunitiesHandler(c *gin.Context) {
	filter := oppservice.ListOpportunitiesFilter{}

	// Parse query parameters
	if category := c.Query("category"); category != "" {
		filter.Category = &category
	}
	if location := c.Query("location"); location != "" {
		filter.Location = &location
	}
	if budgetMinStr := c.Query("budget_min"); budgetMinStr != "" {
		if budgetMin, err := strconv.ParseFloat(budgetMinStr, 64); err == nil {
			filter.BudgetMin = &budgetMin
		}
	}
	if budgetMaxStr := c.Query("budget_max"); budgetMaxStr != "" {
		if budgetMax, err := strconv.ParseFloat(budgetMaxStr, 64); err == nil {
			filter.BudgetMax = &budgetMax
		}
	}
	filter.StartDateFrom = h.parseDateParam(c, "start_date_from")
	filter.StartDateTo = h.parseDateParam(c, "start_date_to")
	if languages := c.QueryArray("language"); len(languages) > 0 {
		filter.Languages = languages
	}
	if urgentStr := c.Query("urgent"); urgentStr != "" {
		if urgent, err := strconv.ParseBool(urgentStr); err == nil {
			filter.Urgent = &urgent
		}
	}
	if status := c.Query("status"); status != "" {
		filter.Status = &status
	}

	// Sorting
	if sort := c.Query("sort"); sort != "" {
		filter.Sort = sort
	} else {
		filter.Sort = "published_at"
	}
	if order := c.Query("order"); order != "" {
		filter.Order = order
	} else {
		filter.Order = "desc"
	}

	// Pagination
	filter.Page, filter.Limit = h.parsePaginationParams(c)

	result, err := h.oppService.List(c.Request.Context(), filter)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	// Check saved status for authenticated users
	var userID *int64
	if id, err := middleware.GetUserID(c); err == nil {
		userID = &id
	}

	responses := make([]*OpportunityResponse, len(result.Data))
	for i, opp := range result.Data {
		var isSaved *bool
		if userID != nil {
			if saved, err := h.savedItemClient.IsOpportunitySaved(c.Request.Context(), *userID, opp.ID); err == nil {
				isSaved = &saved
			}
		}
		responses[i] = h.toOpportunityResponse(opp, isSaved)
	}

	c.JSON(http.StatusOK, response.Success(&ListOpportunitiesResponse{
		Data:  responses,
		Page:  result.Page,
		Limit: result.Limit,
		Total: result.Total,
	}))
}

// GetOpportunityHandler handles GET /opportunities/:id
func (h *OpportunityHandler) GetOpportunityHandler(c *gin.Context) {
	var req struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid opportunity id")
		return
	}

	opp, err := h.oppService.GetByID(c.Request.Context(), req.ID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	// Check saved status for authenticated users
	var isSaved *bool
	if userID, err := middleware.GetUserID(c); err == nil {
		if saved, err := h.savedItemClient.IsOpportunitySaved(c.Request.Context(), userID, opp.ID); err == nil {
			isSaved = &saved
		}
	}

	c.JSON(http.StatusOK, response.Success(h.toOpportunityResponse(opp, isSaved)))
}

// CreateOpportunityHandler handles POST /opportunities
// Only clients (org_person, org_team) can create opportunities; students cannot
func (h *OpportunityHandler) CreateOpportunityHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	role, err := middleware.GetRole(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}
	// Students and alumni cannot create opportunities; only clients (org roles) can
	if role == "student" || role == "alumni" {
		h.respondError(c, http.StatusForbidden, errs.ErrForbidden, "only clients can post opportunities; students can upload services instead")
		return
	}

	var req CreateOpportunityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	input := oppservice.CreateOpportunityInput{
		Title:          req.Title,
		Organization:   req.Organization,
		Category:       req.Category,
		BudgetType:     req.BudgetType,
		BudgetValue:    req.BudgetValue,
		Location:       req.Location,
		DurationMonths: req.DurationMonths,
		Languages:      req.Languages,
		StartDate:      req.StartDate,
		Urgent:         req.Urgent,
		Description:    req.Description,
		Tags:           req.Tags,
		CreatedBy:      userID,
	}

	opp, err := h.oppService.Create(c.Request.Context(), input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.Success(h.toOpportunityResponse(opp, nil)))
}

// UpdateOpportunityHandler handles PATCH /opportunities/:id
func (h *OpportunityHandler) UpdateOpportunityHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	var uriReq struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uriReq); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid opportunity id")
		return
	}

	// Check ownership before proceeding
	_, err = h.checkOwnership(c, uriReq.ID, userID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	var req UpdateOpportunityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	input := oppservice.UpdateOpportunityInput{
		Title:          req.Title,
		Organization:   req.Organization,
		Category:       req.Category,
		BudgetType:     req.BudgetType,
		BudgetValue:    req.BudgetValue,
		Location:       req.Location,
		DurationMonths: req.DurationMonths,
		Languages:      req.Languages,
		StartDate:      req.StartDate,
		Urgent:         req.Urgent,
		Description:    req.Description,
		Tags:           req.Tags,
	}

	opp, err := h.oppService.Update(c.Request.Context(), uriReq.ID, userID, input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(h.toOpportunityResponse(opp, nil)))
}

// DeleteOpportunityHandler handles DELETE /opportunities/:id
func (h *OpportunityHandler) DeleteOpportunityHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	var req struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid opportunity id")
		return
	}

	// Check ownership before proceeding
	_, err = h.checkOwnership(c, req.ID, userID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	if err := h.oppService.Delete(c.Request.Context(), req.ID, userID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(&MessageResponse{Message: "opportunity deleted successfully"}))
}

// SaveOpportunityHandler handles POST /opportunities/:id/favorite
// This uses the user-service's saved items API
func (h *OpportunityHandler) SaveOpportunityHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	var req struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid opportunity id")
		return
	}

	if err := h.savedItemClient.SaveOpportunity(c.Request.Context(), userID, req.ID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(&MessageResponse{Message: "opportunity saved successfully"}))
}

// UnsaveOpportunityHandler handles DELETE /opportunities/:id/favorite
// This uses the user-service's saved items API
func (h *OpportunityHandler) UnsaveOpportunityHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	var req struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid opportunity id")
		return
	}

	if err := h.savedItemClient.UnsaveOpportunity(c.Request.Context(), userID, req.ID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(&MessageResponse{Message: "opportunity unsaved successfully"}))
}

// CreateApplicationHandler handles POST /opportunities/:id/apply
func (h *OpportunityHandler) CreateApplicationHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	var uriReq struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uriReq); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid opportunity id")
		return
	}

	var req CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	input := oppservice.CreateApplicationInput{
		OpportunityID: uriReq.ID,
		Message:       req.Message,
		CVURL:         req.CVURL,
		PortfolioURL:  req.PortfolioURL,
	}

	app, err := h.applicationService.CreateApplication(c.Request.Context(), userID, input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	resp := &ApplicationResponse{
		ID:            app.ID,
		UserID:        app.UserID,
		OpportunityID: app.OpportunityID,
		Message:       app.Message,
		CVURL:         app.CVURL,
		PortfolioURL:  app.PortfolioURL,
		Status:        app.Status.String(),
		CreatedAt:     app.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     app.UpdatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusCreated, response.Success(resp))
}

// ListMyOpportunitiesHandler handles GET /opportunities/me
func (h *OpportunityHandler) ListMyOpportunitiesHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	status := c.Query("status")
	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	opportunities, err := h.oppService.ListByUserID(c.Request.Context(), userID, statusPtr, page, limit)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(opportunities))
}

// UpdateOpportunityStatusHandler handles PATCH /opportunities/:id/status
func (h *OpportunityHandler) UpdateOpportunityStatusHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid opportunity id")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	err = h.oppService.UpdateStatus(c.Request.Context(), id, userID, model.OpportunityStatus(req.Status))
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

// GetOpportunityStatsHandler handles GET /opportunities/:id/stats
func (h *OpportunityHandler) GetOpportunityStatsHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid opportunity id")
		return
	}

	stats, err := h.oppService.GetStats(c.Request.Context(), id, userID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(stats))
}

// ListMyApplicationsHandler handles GET /users/me/applications
func (h *OpportunityHandler) ListMyApplicationsHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	status := c.Query("status")
	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	applications, err := h.applicationService.ListByUserID(c.Request.Context(), userID, statusPtr, page, limit)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	responses := make([]*ApplicationResponse, len(applications))
	for i, app := range applications {
		responses[i] = &ApplicationResponse{
			ID:            app.ID,
			UserID:        app.UserID,
			OpportunityID: app.OpportunityID,
			Message:       app.Message,
			CVURL:         app.CVURL,
			PortfolioURL:  app.PortfolioURL,
			Status:        app.Status.String(),
			CreatedAt:     app.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     app.UpdatedAt.Format(time.RFC3339),
		}
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"applications": responses,
		"items":        responses,
	}))
}

// ListOpportunityApplicationsHandler handles GET /opportunities/:id/applications
func (h *OpportunityHandler) ListOpportunityApplicationsHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	var uriReq struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uriReq); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid opportunity id")
		return
	}

	page, limit := h.parsePaginationParams(c)

	applications, err := h.applicationService.ListByOpportunityID(c.Request.Context(), uriReq.ID, userID, page, limit)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	responses := make([]*ApplicationResponse, len(applications))
	for i, app := range applications {
		responses[i] = &ApplicationResponse{
			ID:            app.ID,
			UserID:        app.UserID,
			OpportunityID: app.OpportunityID,
			Message:       app.Message,
			CVURL:         app.CVURL,
			PortfolioURL:  app.PortfolioURL,
			Status:        app.Status.String(),
			CreatedAt:     app.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     app.UpdatedAt.Format(time.RFC3339),
		}
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"applications": responses,
		"items":        responses,
	}))
}
