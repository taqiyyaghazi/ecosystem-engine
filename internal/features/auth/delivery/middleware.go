package delivery

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/apperror"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/auth/usecase"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/platform/http/httputil"
)

// RequireAuth is the session-validation middleware.
// It reads the session_id from the Secure HttpOnly Cookie or the
// X-Session-ID header (for non-browser clients), validates the
// session's existence in Redis, applies sliding expiration, and
// injects the session data into the Gin context under the "session"
// key. Returns 401 Unauthorized if the session is missing or expired.
func RequireAuth(uc usecase.AuthUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, _ := c.Cookie(sessionCookieName)
		if sessionID == "" {
			sessionID = c.GetHeader("X-Session-ID")
		}

		if sessionID == "" {
			httputil.HandleError(c, apperror.ErrUnauthorized)
			return
		}

		sessionData, err := uc.Me(c.Request.Context(), sessionID)
		if err != nil {
			slog.Error("session validation failed", "error", err, "session_id", sessionID)
			httputil.HandleError(c, apperror.ErrUnauthorized)
			return
		}

		if err := uc.RefreshSession(c.Request.Context(), sessionID); err != nil {
			slog.Warn("failed to refresh session TTL", "error", err, "session_id", sessionID)
		}

		c.Set("session_id", sessionID)
		c.Set("session", sessionData)
		c.Set("user_id", sessionData.UserID)

		c.Next()
	}
}
