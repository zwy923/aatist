package handler

import (
	"net/http"

	"github.com/aatist/backend/internal/community/service"
	"github.com/aatist/backend/internal/platform/log"
	"github.com/aatist/backend/internal/platform/middleware"
	"github.com/aatist/backend/pkg/errs"
	"github.com/aatist/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// LikeHandler handles HTTP requests related to likes
type LikeHandler struct {
	HandlerBase
	likeSvc service.LikeService
}

// NewLikeHandler creates a new like handler
func NewLikeHandler(likeSvc service.LikeService, logger *log.Logger) *LikeHandler {
	return &LikeHandler{
		HandlerBase: HandlerBase{logger: logger},
		likeSvc:     likeSvc,
	}
}

// LikePostHandler handles POST /community/posts/:id/like
func (h *LikeHandler) LikePostHandler(c *gin.Context) {
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
	liked, err := h.likeSvc.LikePost(c.Request.Context(), uri.ID, userID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(gin.H{"liked": liked}))
}

// UnlikePostHandler handles DELETE /community/posts/:id/like
func (h *LikeHandler) UnlikePostHandler(c *gin.Context) {
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
	if err := h.likeSvc.UnlikePost(c.Request.Context(), uri.ID, userID); err != nil {
		h.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(gin.H{"liked": false}))
}
