package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/aatist/backend/internal/platform/auth"
	"github.com/aatist/backend/internal/user/repository"
	"github.com/aatist/backend/pkg/errs"
	"github.com/aatist/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// JWTUserBindingMiddleware ensures JWT claims still match the current database user.
// Without this, after a full DB reset the same numeric user IDs are reused (starting at 1) while
// old access tokens remain valid if the signing secret is unchanged, so a stale token would
// authenticate as whoever now owns that ID.
func JWTUserBindingMiddleware(jwt *auth.JWT, users repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, response.Error(errs.NewAppError(errs.ErrInvalidToken, http.StatusUnauthorized, "authorization token required").WithCode(errs.CodeInvalidToken)))
			c.Abort()
			return
		}

		claims, err := jwt.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, response.Error(errs.NewAppError(errs.ErrInvalidToken, http.StatusUnauthorized, "invalid or expired token").WithCode(errs.CodeInvalidToken)))
			c.Abort()
			return
		}

		if strings.TrimSpace(claims.Email) == "" {
			c.JSON(http.StatusUnauthorized, response.Error(errs.NewAppError(errs.ErrInvalidToken, http.StatusUnauthorized, "invalid token: missing email claim").WithCode(errs.CodeInvalidToken)))
			c.Abort()
			return
		}

		user, err := users.FindByID(c.Request.Context(), claims.UserID)
		if err != nil {
			if errors.Is(err, errs.ErrUserNotFound) {
				c.JSON(http.StatusUnauthorized, response.Error(errs.NewAppError(errs.ErrInvalidToken, http.StatusUnauthorized, "invalid session").WithCode(errs.CodeInvalidToken)))
				c.Abort()
				return
			}
			c.JSON(http.StatusInternalServerError, response.Error(errs.NewAppError(errs.ErrInternalError, http.StatusInternalServerError, "failed to verify session").WithCode(errs.CodeInternalError)))
			c.Abort()
			return
		}

		if !strings.EqualFold(strings.TrimSpace(user.Email), strings.TrimSpace(claims.Email)) {
			c.JSON(http.StatusUnauthorized, response.Error(errs.NewAppError(errs.ErrInvalidToken, http.StatusUnauthorized, "session no longer valid").WithCode(errs.CodeInvalidToken)))
			c.Abort()
			return
		}

		c.Set("user_id", user.ID)
		c.Set("role", user.Role.String())
		c.Set("email", user.Email)
		c.Set("claims", claims)

		c.Next()
	}
}
