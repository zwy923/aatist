package handler

import (
	"net/http"

	"github.com/aatist/backend/internal/community/model"
	"github.com/aatist/backend/internal/community/service"
	"github.com/aatist/backend/internal/platform/log"
	"github.com/aatist/backend/internal/platform/middleware"
	"github.com/aatist/backend/pkg/errs"
	"github.com/aatist/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// CommentHandler handles HTTP requests related to comments
type CommentHandler struct {
	HandlerBase
	commentSvc service.CommentService
}

// NewCommentHandler creates a new comment handler
func NewCommentHandler(commentSvc service.CommentService, logger *log.Logger) *CommentHandler {
	return &CommentHandler{
		HandlerBase: HandlerBase{logger: logger},
		commentSvc: commentSvc,
	}
}

// ListCommentsHandler handles GET /community/posts/:id/comments
func (h *CommentHandler) ListCommentsHandler(c *gin.Context) {
	var uri struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid post id")
		return
	}
	var query struct {
		Limit  int `form:"limit"`
		Offset int `form:"offset"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid query parameters")
		return
	}
	if query.Limit <= 0 || query.Limit > 100 {
		query.Limit = 50
	}
	if query.Offset < 0 {
		query.Offset = 0
	}

	comments, err := h.commentSvc.ListComments(c.Request.Context(), uri.ID, query.Limit, query.Offset)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(comments))
}

// CreateCommentHandler handles POST /community/posts/:id/comments
func (h *CommentHandler) CreateCommentHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}
	var uri struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid post id")
		return
	}
	var req commentMutationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	comment := &model.Comment{
		PostID:   uri.ID,
		UserID:   userID,
		ParentID: req.ParentID,
		Content:  req.Content,
	}

	if err := h.commentSvc.CreateComment(c.Request.Context(), comment); err != nil {
		h.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.Success(comment))
}

// UpdateCommentHandler handles PUT /community/comments/:id
func (h *CommentHandler) UpdateCommentHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}
	var uri struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid comment id")
		return
	}
	var req commentMutationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	comment := &model.Comment{
		ID:      uri.ID,
		Content: req.Content,
	}
	if err := h.commentSvc.UpdateComment(c.Request.Context(), comment, userID); err != nil {
		h.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(gin.H{"message": "comment updated"}))
}

// DeleteCommentHandler handles DELETE /community/comments/:id
func (h *CommentHandler) DeleteCommentHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}
	var uri struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid comment id")
		return
	}
	if err := h.commentSvc.DeleteComment(c.Request.Context(), uri.ID, userID); err != nil {
		h.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(gin.H{"message": "comment deleted"}))
}

