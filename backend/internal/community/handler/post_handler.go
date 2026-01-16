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
)

// PostHandler handles HTTP requests related to discussion posts
type PostHandler struct {
	HandlerBase
	postSvc service.PostService
}

// NewPostHandler creates a new post handler
func NewPostHandler(postSvc service.PostService, logger *log.Logger) *PostHandler {
	return &PostHandler{
		HandlerBase: HandlerBase{logger: logger},
		postSvc:     postSvc,
	}
}

// GetPostsHandler handles GET /community/posts
func (h *PostHandler) GetPostsHandler(c *gin.Context) {
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

		userID, _ := middleware.GetUserID(c)
		h.postSvc.EnrichPostsWithLikes(c.Request.Context(), posts, userID)

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

	userID, _ := middleware.GetUserID(c)
	h.postSvc.EnrichPostsWithLikes(c.Request.Context(), posts, userID)

	c.JSON(http.StatusOK, response.Success(posts))
}

// GetTrendingPostsHandler handles GET /community/posts/trending
func (h *PostHandler) GetTrendingPostsHandler(c *gin.Context) {
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

	userID, _ := middleware.GetUserID(c)
	h.postSvc.EnrichPostsWithLikes(c.Request.Context(), posts, userID)

	c.JSON(http.StatusOK, response.Success(posts))
}

// GetPostDetailHandler handles GET /community/posts/:id
func (h *PostHandler) GetPostDetailHandler(c *gin.Context) {
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

	userID, _ := middleware.GetUserID(c)
	h.postSvc.EnrichPostWithLike(c.Request.Context(), post, userID)

	c.JSON(http.StatusOK, response.Success(post))
}

// CreatePostHandler handles POST /community/posts
func (h *PostHandler) CreatePostHandler(c *gin.Context) {
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
func (h *PostHandler) UpdatePostHandler(c *gin.Context) {
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
func (h *PostHandler) DeletePostHandler(c *gin.Context) {
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
	role, _ := middleware.GetRole(c)
	isAdmin := role == "admin" || role == "org_team" // Assuming org_team or admin can moderate

	if err := h.postSvc.DeletePost(c.Request.Context(), uri.ID, userID, isAdmin); err != nil {
		h.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(gin.H{"message": "post deleted"}))
}

// GetUserPostsHandler handles GET /community/users/:id/posts
func (h *PostHandler) GetUserPostsHandler(c *gin.Context) {
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

	userID, _ := middleware.GetUserID(c)
	h.postSvc.EnrichPostsWithLikes(c.Request.Context(), posts, userID)

	c.JSON(http.StatusOK, response.Success(posts))
}

// GetMyPostsHandler handles GET /community/users/me/posts
func (h *PostHandler) GetMyPostsHandler(c *gin.Context) {
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

	h.postSvc.EnrichPostsWithLikes(c.Request.Context(), posts, userID)

	c.JSON(http.StatusOK, response.Success(posts))
}
