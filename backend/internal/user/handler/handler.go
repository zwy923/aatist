package handler

import (
	"net/http"
	"time"

	"github.com/aalto-talent-network/backend/internal/platform/log"
	"github.com/aalto-talent-network/backend/internal/user/model"
	"github.com/aalto-talent-network/backend/internal/user/service"
	"github.com/aalto-talent-network/backend/pkg/errs"
	"github.com/aalto-talent-network/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService   service.AuthService
	emailVerifSvc *service.EmailVerificationService
	mq            interface {
		PublishEmailVerification(message interface{}) error
	}
	logger *log.Logger
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(
	authService service.AuthService,
	emailVerifSvc *service.EmailVerificationService,
	mq interface {
		PublishEmailVerification(message interface{}) error
	},
	logger *log.Logger,
) *AuthHandler {
	return &AuthHandler{
		authService:   authService,
		emailVerifSvc: emailVerifSvc,
		mq:            mq,
		logger:        logger,
	}
}

// RegisterHandler handles user registration
func (h *AuthHandler) RegisterHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	ip := h.getClientIP(c)
	ctx := c.Request.Context()
	user, tokens, err := h.authService.Register(ctx, req.Email, req.Password, req.Name, ip)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	// Generate verification token and publish to MQ asynchronously
	if h.emailVerifSvc != nil && h.mq != nil {
		token, err := h.emailVerifSvc.GenerateVerificationToken(ctx, user.ID, user.Email)
		if err != nil {
			h.logger.Error("Failed to generate verification token", zap.Error(err))
			// Non-critical, continue with registration
		} else {
			// Publish to MQ for async email sending
			emailMsg := model.EmailVerificationRequest{
				UserID: user.ID,
				Email:  user.Email,
				Name:   user.Name,
				Token:  token,
			}
			if err := h.mq.PublishEmailVerification(emailMsg); err != nil {
				h.logger.Error("Failed to publish email verification message", zap.Error(err))
				// Non-critical, continue with registration
			}
		}
	}

	h.respondSuccess(c, http.StatusCreated, user, tokens)
}

// LoginHandler handles user login
func (h *AuthHandler) LoginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	ip := h.getClientIP(c)
	user, tokens, err := h.authService.Login(c.Request.Context(), req.Email, req.Password, ip)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	h.respondSuccess(c, http.StatusOK, user, tokens)
}

// RefreshTokenHandler handles token refresh
func (h *AuthHandler) RefreshTokenHandler(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	tokens, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	}))
}

// LogoutHandler handles user logout
func (h *AuthHandler) LogoutHandler(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	if err := h.authService.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "logged out successfully"}))
}

// VerifyEmailHandler handles email verification
func (h *AuthHandler) VerifyEmailHandler(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	if err := h.authService.VerifyEmail(c.Request.Context(), req.Token); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "email verified successfully"}))
}

// Helper methods

func (h *AuthHandler) respondSuccess(c *gin.Context, statusCode int, user *model.User, tokens *service.Tokens) {
	userResp := UserResponse{
		ID:              user.ID,
		Email:           user.Email,
		Name:            user.Name,
		Role:            user.Role.String(),
		IsVerifiedEmail: user.IsVerifiedEmail,
		OAuthProvider:   user.OAuthProvider,
		CreatedAt:       user.CreatedAt.Format(time.RFC3339),
	}

	if user.LastLoginAt != nil {
		lastLogin := user.LastLoginAt.Format(time.RFC3339)
		userResp.LastLoginAt = &lastLogin
	}

	data := AuthResponse{
		User:         userResp,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}

	c.JSON(statusCode, response.Success(data))
}

func (h *AuthHandler) respondError(c *gin.Context, statusCode int, err error, message string) {
	appErr := errs.NewAppError(err, statusCode, message).WithCode(errs.GetErrorCode(err))
	c.JSON(statusCode, response.Error(appErr))
}

func (h *AuthHandler) handleServiceError(c *gin.Context, err error) {
	statusCode := errs.ToHTTPStatus(err)
	var message string
	var code string

	if appErr, ok := err.(*errs.AppError); ok {
		message = appErr.Message
		code = appErr.Code
		if code == "" {
			code = errs.GetErrorCode(err)
		}
	} else {
		message = err.Error()
		code = errs.GetErrorCode(err)
	}

	h.logger.Error("Service error",
		zap.Error(err),
		zap.Int("status_code", statusCode),
		zap.String("error_code", code),
	)

	appErr := errs.NewAppError(err, statusCode, message).WithCode(code)
	c.JSON(statusCode, response.Error(appErr))
}

func (h *AuthHandler) getClientIP(c *gin.Context) string {
	ip := c.GetHeader("X-Forwarded-For")
	if ip == "" {
		ip = c.GetHeader("X-Real-IP")
	}
	if ip == "" {
		ip = c.ClientIP()
	}
	return ip
}
