package handler

import (
	"net/http"

	"github.com/aatist/backend/internal/platform/log"
	"github.com/aatist/backend/internal/platform/middleware"
	"github.com/aatist/backend/internal/portfolio/model"
	"github.com/aatist/backend/internal/portfolio/service"
	"github.com/aatist/backend/pkg/errs"
	"github.com/aatist/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PortfolioHandler struct {
	projectSvc service.ProjectService
	logger     *log.Logger
}

func NewPortfolioHandler(projectSvc service.ProjectService, logger *log.Logger) *PortfolioHandler {
	return &PortfolioHandler{
		projectSvc: projectSvc,
		logger:     logger,
	}
}

func (h *PortfolioHandler) respondError(c *gin.Context, status int, err error, message string) {
	h.logger.Error("Handler error", zap.Error(err), zap.String("message", message))
	c.JSON(status, response.Error(errs.NewAppError(err, status, message)))
}

func (h *PortfolioHandler) handleServiceError(c *gin.Context, err error) {
	if appErr, ok := err.(*errs.AppError); ok {
		h.respondError(c, appErr.StatusCode, appErr, appErr.Message)
		return
	}
	h.respondError(c, http.StatusInternalServerError, err, "internal server error")
}

// GetProjectDetailHandler returns a single project by ID (public)
func (h *PortfolioHandler) GetProjectDetailHandler(c *gin.Context) {
	var req struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid project id")
		return
	}

	project, err := h.projectSvc.GetProject(c.Request.Context(), req.ID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	// Note: Profile visibility check should be done via user-service client
	// For now, we'll allow access - visibility check can be added later if needed

	c.JSON(http.StatusOK, response.Success(project))
}

// GetPublicProjectsHandler returns all projects (public)
func (h *PortfolioHandler) GetPublicProjectsHandler(c *gin.Context) {
	var req struct {
		Limit  int `form:"limit"`
		Offset int `form:"offset"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid query parameters")
		return
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	projects, err := h.projectSvc.GetPublicProjects(c.Request.Context(), req.Limit, req.Offset)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"projects": projects,
		"items":    projects,
	}))
}

// GetUserPortfolioHandler returns all projects for a user (public)
func (h *PortfolioHandler) GetUserPortfolioHandler(c *gin.Context) {
	var req struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid user id")
		return
	}

	projects, err := h.projectSvc.GetUserProjects(c.Request.Context(), req.ID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"projects": projects,
		"items":    projects,
	}))
}

// GetMyPortfolioHandler returns all projects for the authenticated user
func (h *PortfolioHandler) GetMyPortfolioHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	projects, err := h.projectSvc.GetUserProjects(c.Request.Context(), userID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"projects": projects,
		"items":    projects,
	}))
}

// CreateProjectHandler creates a new project
func (h *PortfolioHandler) CreateProjectHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Title         string   `json:"title" binding:"required"`
		ClientName    *string  `json:"client_name"`
		Description   *string  `json:"description"`
		Year          *int     `json:"year"`
		Tags          []string `json:"tags"`
		CoverImageURL *string  `json:"cover_image_url"`
		ProjectLink   *string  `json:"project_link"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	project := &model.Project{
		Title:         req.Title,
		ClientName:    req.ClientName,
		Description:   req.Description,
		Year:          req.Year,
		Tags:          model.StringArray(req.Tags),
		CoverImageURL: req.CoverImageURL,
		ProjectLink:   req.ProjectLink,
	}

	if err := h.projectSvc.CreateProject(c.Request.Context(), userID, project); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.Success(project))
}

// UpdateProjectHandler updates an existing project
func (h *PortfolioHandler) UpdateProjectHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	var uriReq struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uriReq); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid project id")
		return
	}

	var req struct {
		Title         string   `json:"title" binding:"required"`
		ClientName    *string  `json:"client_name"`
		Description   *string  `json:"description"`
		Year          *int     `json:"year"`
		Tags          []string `json:"tags"`
		CoverImageURL *string  `json:"cover_image_url"`
		ProjectLink   *string  `json:"project_link"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	project := &model.Project{
		ID:            uriReq.ID,
		Title:         req.Title,
		ClientName:    req.ClientName,
		Description:   req.Description,
		Year:          req.Year,
		Tags:          model.StringArray(req.Tags),
		CoverImageURL: req.CoverImageURL,
		ProjectLink:   req.ProjectLink,
	}

	if err := h.projectSvc.UpdateProject(c.Request.Context(), userID, project); err != nil {
		h.handleServiceError(c, err)
		return
	}

	// Fetch updated project
	updated, err := h.projectSvc.GetProject(c.Request.Context(), uriReq.ID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(updated))
}

// DeleteProjectHandler deletes a project
func (h *PortfolioHandler) DeleteProjectHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	var req struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid project id")
		return
	}

	if err := h.projectSvc.DeleteProject(c.Request.Context(), userID, req.ID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "project deleted successfully"}))
}
