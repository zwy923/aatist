package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/aalto-talent-network/backend/internal/platform/log"
	"github.com/aalto-talent-network/backend/pkg/errs"
	"github.com/aalto-talent-network/backend/pkg/response"
	"go.uber.org/zap"
)

// RecoveryMiddleware recovers from panics and returns a 500 error
func RecoveryMiddleware(logger *log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				logger.Error("Panic recovered",
					zap.Any("error", err),
					RequestIDLogField(c),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)

				// Return 500 error
				c.JSON(http.StatusInternalServerError, response.Error(
					errs.NewAppError(
						errs.ErrInternalError,
						http.StatusInternalServerError,
						"internal server error",
					).WithCode(errs.CodeInternalError),
				))

				c.Abort()
			}
		}()

		c.Next()
	}
}

