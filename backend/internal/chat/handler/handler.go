package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/aatist/backend/internal/chat/service"
	"github.com/aatist/backend/internal/platform/log"
	"github.com/aatist/backend/internal/platform/middleware"
	"github.com/aatist/backend/pkg/errs"
	"github.com/aatist/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// ChatHandler handles chat HTTP requests
type ChatHandler struct {
	chatSvc service.ChatService
	logger  *log.Logger
}

// NewChatHandler creates a new chat handler
func NewChatHandler(chatSvc service.ChatService, logger *log.Logger) *ChatHandler {
	return &ChatHandler{chatSvc: chatSvc, logger: logger}
}

// CreateMessageRequest body for internal create message
type CreateMessageRequest struct {
	ConversationID string `json:"conversation_id"`
	FromUserID     int64  `json:"from_user_id"`
	Content        string `json:"content"`
	CreatedAt      string `json:"created_at"`
}

// CreateMessageHandler internal only: persist a message (called by gateway after WS receive)
func (h *ChatHandler) CreateMessageHandler(c *gin.Context) {
	var req CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errs.NewAppError(errs.ErrInvalidInput, http.StatusBadRequest, "invalid body").WithCode(errs.CodeInvalidInput)))
		return
	}
	if req.ConversationID == "" || req.FromUserID == 0 || req.Content == "" {
		c.JSON(http.StatusBadRequest, response.Error(errs.NewAppError(errs.ErrInvalidInput, http.StatusBadRequest, "conversation_id, from_user_id, content required").WithCode(errs.CodeInvalidInput)))
		return
	}
	msg, err := h.chatSvc.CreateMessage(c.Request.Context(), req.ConversationID, req.FromUserID, req.Content, req.CreatedAt)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.Success(msg))
}

// StartConversationRequest body for starting a new conversation
type StartConversationRequest struct {
	OtherUserID int64 `json:"other_user_id"`
}

// StartConversationHandler returns conversation_id for messaging another user (no DB write until first message)
func (h *ChatHandler) StartConversationHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Error(errs.ErrUnauthorized))
		return
	}
	var req StartConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errs.NewAppError(errs.ErrInvalidInput, http.StatusBadRequest, "invalid body").WithCode(errs.CodeInvalidInput)))
		return
	}
	if req.OtherUserID <= 0 {
		c.JSON(http.StatusBadRequest, response.Error(errs.NewAppError(errs.ErrInvalidInput, http.StatusBadRequest, "other_user_id required").WithCode(errs.CodeInvalidInput)))
		return
	}
	if req.OtherUserID == userID {
		c.JSON(http.StatusBadRequest, response.Error(errs.NewAppError(errs.ErrInvalidInput, http.StatusBadRequest, "cannot message yourself").WithCode(errs.CodeInvalidInput)))
		return
	}
	convID := conversationID(userID, req.OtherUserID)
	c.JSON(http.StatusOK, response.Success(gin.H{"conversation_id": convID}))
}

// conversationID generates "smallID_largeID" format
func conversationID(a, b int64) string {
	if a > b {
		a, b = b, a
	}
	return strconv.FormatInt(a, 10) + "_" + strconv.FormatInt(b, 10)
}

// GetConversationsHandler returns conversation list for the current user
func (h *ChatHandler) GetConversationsHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Error(errs.ErrUnauthorized))
		return
	}
	limit := 50
	if l := c.Query("limit"); l != "" {
		if n, e := strconv.Atoi(l); e == nil && n > 0 {
			if n > 100 {
				n = 100
			}
			limit = n
		}
	}
	list, err := h.chatSvc.ListConversations(c.Request.Context(), userID, limit)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(gin.H{"conversations": list}))
}

// GetMessagesHandler returns message history for a conversation
func (h *ChatHandler) GetMessagesHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Error(errs.ErrUnauthorized))
		return
	}
	conversationID := strings.TrimSpace(c.Param("id"))
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, response.Error(errs.NewAppError(errs.ErrInvalidInput, http.StatusBadRequest, "conversation id required").WithCode(errs.CodeInvalidInput)))
		return
	}
	limit, offset := 50, 0
	if l := c.Query("limit"); l != "" {
		if n, e := strconv.Atoi(l); e == nil && n > 0 {
			if n > 100 {
				n = 100
			}
			limit = n
		}
	}
	if o := c.Query("offset"); o != "" {
		if n, e := strconv.Atoi(o); e == nil && n >= 0 {
			offset = n
		}
	}
	list, err := h.chatSvc.ListMessages(c.Request.Context(), conversationID, userID, limit, offset)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(gin.H{"messages": list}))
}

// MarkConversationAsReadHandler marks a conversation as read for the current user
func (h *ChatHandler) MarkConversationAsReadHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Error(errs.ErrUnauthorized))
		return
	}
	conversationID := strings.TrimSpace(c.Param("id"))
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, response.Error(errs.NewAppError(errs.ErrInvalidInput, http.StatusBadRequest, "conversation id required").WithCode(errs.CodeInvalidInput)))
		return
	}
	if err := h.chatSvc.MarkConversationAsRead(c.Request.Context(), userID, conversationID); err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(gin.H{"message": "marked as read"}))
}

// DeleteConversationHandler handles DELETE /conversations/:id - deletes all messages in the conversation
func (h *ChatHandler) DeleteConversationHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Error(errs.ErrUnauthorized))
		return
	}
	conversationID := strings.TrimSpace(c.Param("id"))
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, response.Error(errs.NewAppError(errs.ErrInvalidInput, http.StatusBadRequest, "conversation id required").WithCode(errs.CodeInvalidInput)))
		return
	}
	if err := h.chatSvc.DeleteConversation(c.Request.Context(), conversationID, userID); err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(gin.H{"message": "conversation deleted"}))
}

func (h *ChatHandler) respondServiceError(c *gin.Context, err error) {
	statusCode := errs.ToHTTPStatus(err)
	c.JSON(statusCode, response.Error(err))
}
