package handler

import (
	"net/http"
	"strings"

	"github.com/aatist/backend/internal/community/model"
	"github.com/aatist/backend/internal/community/repository"
	"github.com/aatist/backend/internal/community/service"
	"github.com/aatist/backend/internal/platform/log"
	"github.com/aatist/backend/internal/platform/middleware"
	"github.com/aatist/backend/pkg/errs"
	"github.com/aatist/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CommunityHandler wires HTTP endpoints to services.
type CommunityHandler struct {
	postSvc    service.PostService
	commentSvc service.CommentService
	likeSvc    service.LikeService
	logger     *log.Logger
}

func NewCommunityHandler(postSvc service.PostService, commentSvc service.CommentService, likeSvc service.LikeService, logger *log.Logger) *CommunityHandler {
	return &CommunityHandler{
		postSvc:    postSvc,
		commentSvc: commentSvc,
		likeSvc:    likeSvc,
		logger:     logger,
	}
}

func (h *CommunityHandler) respondError(c *gin.Context, status int, err error, message string) {
	h.logger.Warn("community handler error", zap.Int("status", status), zap.String("path", c.Request.URL.Path), zap.Error(err))
	c.JSON(status, response.Error(errs.NewAppError(err, status, message)))
}

func (h *CommunityHandler) handleServiceError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	if err == errs.ErrNotFound || err == errs.ErrUserNotFound {
		h.respondError(c, http.StatusNotFound, err, "resource not found")
		return
	}
	if err == errs.ErrUnauthorized {
		h.respondError(c, http.StatusUnauthorized, err, "unauthorized")
		return
	}
	h.respondError(c, http.StatusInternalServerError, err, "internal server error")
}

// GetPostsHandler handles GET /community/posts
func (h *CommunityHandler) GetPostsHandler(c *gin.Context) {
	var query listPostsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid query parameters")
		return
	}

	limit := query.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	if query.Search != "" {
		filter := repository.PostSearchFilter{
			Query:  query.Search,
			Limit:  limit,
			Offset: offset,
		}
		var (
			posts []*model.DiscussionPost
			err   error
		)
		if strings.ToLower(query.Sort) == "trending" {
			posts, err = h.postSvc.SearchPostsTrending(c.Request.Context(), filter)
		} else {
			posts, err = h.postSvc.SearchPosts(c.Request.Context(), filter)
		}
		if err != nil {
			h.handleServiceError(c, err)
			return
		}
		c.JSON(http.StatusOK, response.Success(posts))
		return
	}

	filter := repository.PostListFilter{
		Limit:  limit,
		Offset: offset,
	}
	if cat, ok := parseCategory(query.Category); ok {
		filter.Category = &cat
	}
	switch strings.ToLower(query.Sort) {
	case "oldest":
		filter.Sort = repository.PostListSortOldest
	default:
		filter.Sort = repository.PostListSortNewest
	}

	posts, err := h.postSvc.ListPosts(c.Request.Context(), filter)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(posts))
}

// GetTrendingPostsHandler handles GET /community/posts/trending
func (h *CommunityHandler) GetTrendingPostsHandler(c *gin.Context) {
	var query struct {
		Limit int `form:"limit"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid query parameters")
		return
	}

	posts, err := h.postSvc.GetTrendingPosts(c.Request.Context(), query.Limit)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(posts))
}

// GetPostDetailHandler handles GET /community/posts/:id
func (h *CommunityHandler) GetPostDetailHandler(c *gin.Context) {
	var uri struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid post id")
		return
	}

	post, err := h.postSvc.GetPost(c.Request.Context(), uri.ID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(post))
}

// CreatePostHandler handles POST /community/posts
func (h *CommunityHandler) CreatePostHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	var req postMutationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	post := &model.DiscussionPost{
		UserID:   userID,
		Title:    req.Title,
		Content:  req.Content,
		Category: normalizeCategory(req.Category),
		Tags:     model.StringArray(req.Tags),
	}

	if err := h.postSvc.CreatePost(c.Request.Context(), post); err != nil {
		h.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.Success(post))
}

// UpdatePostHandler handles PUT /community/posts/:id
func (h *CommunityHandler) UpdatePostHandler(c *gin.Context) {
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

	var req postMutationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	post := &model.DiscussionPost{
		ID:       uri.ID,
		UserID:   userID,
		Title:    req.Title,
		Content:  req.Content,
		Category: normalizeCategory(req.Category),
		Tags:     model.StringArray(req.Tags),
	}

	if err := h.postSvc.UpdatePost(c.Request.Context(), post); err != nil {
		h.handleServiceError(c, err)
		return
	}
	updated, err := h.postSvc.GetPost(c.Request.Context(), uri.ID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(updated))
}

// DeletePostHandler handles DELETE /community/posts/:id
func (h *CommunityHandler) DeletePostHandler(c *gin.Context) {
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
	if err := h.postSvc.DeletePost(c.Request.Context(), uri.ID, userID); err != nil {
		h.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(gin.H{"message": "post deleted"}))
}

// LikePostHandler handles POST /community/posts/:id/like
func (h *CommunityHandler) LikePostHandler(c *gin.Context) {
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
func (h *CommunityHandler) UnlikePostHandler(c *gin.Context) {
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

// ListCommentsHandler handles GET /community/posts/:id/comments
func (h *CommunityHandler) ListCommentsHandler(c *gin.Context) {
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
func (h *CommunityHandler) CreateCommentHandler(c *gin.Context) {
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
func (h *CommunityHandler) UpdateCommentHandler(c *gin.Context) {
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
func (h *CommunityHandler) DeleteCommentHandler(c *gin.Context) {
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

// GetUserPostsHandler handles GET /community/users/:id/posts
func (h *CommunityHandler) GetUserPostsHandler(c *gin.Context) {
	var uri struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid user id")
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

	posts, err := h.postSvc.ListUserPosts(c.Request.Context(), uri.ID, query.Limit, query.Offset)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(posts))
}

// GetMyPostsHandler handles GET /community/users/me/posts
func (h *CommunityHandler) GetMyPostsHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
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

	posts, err := h.postSvc.ListUserPosts(c.Request.Context(), userID, query.Limit, query.Offset)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(posts))
}
