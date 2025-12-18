package handler

import (
	"github.com/aatist/backend/internal/platform/log"
	"github.com/aatist/backend/pkg/errs"
	"github.com/aatist/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HandlerBase provides common error handling methods for all handlers
type HandlerBase struct {
	logger *log.Logger
}

func (h *HandlerBase) respondError(c *gin.Context, status int, err error, message string) {
	h.logger.Warn("community handler error", zap.Int("status", status), zap.String("path", c.Request.URL.Path), zap.Error(err))

	// Use errs package to get error code
	code := errs.GetErrorCode(err)
	appErr := errs.NewAppError(err, status, message).WithCode(code)
	c.JSON(status, response.Error(appErr))
}

func (h *HandlerBase) handleServiceError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	// Use errs.ToHTTPStatus to get the appropriate HTTP status code
	statusCode := errs.ToHTTPStatus(err)
	code := errs.GetErrorCode(err)

	// If it's already an AppError, use its message, otherwise use error string
	var message string
	if appErr, ok := err.(*errs.AppError); ok {
		message = appErr.Message
		if message == "" {
			message = err.Error()
		}
	} else {
		message = err.Error()
	}

	appErr := errs.NewAppError(err, statusCode, message).WithCode(code)
	h.logger.Warn("service error", zap.Int("status", statusCode), zap.String("code", code), zap.Error(err))
	c.JSON(statusCode, response.Error(appErr))
}
