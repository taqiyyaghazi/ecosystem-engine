package delivery

import (
	"net/http"

	"github.com/gin-gonic/gin"
	activityDto "github.com/taqiyyaghazi/ecosystem-engine/internal/features/activity/dto"
	activityUsecase "github.com/taqiyyaghazi/ecosystem-engine/internal/features/activity/usecase"
	authDto "github.com/taqiyyaghazi/ecosystem-engine/internal/features/auth/dto"
	authEntity "github.com/taqiyyaghazi/ecosystem-engine/internal/features/auth/entity"
	authUsecase "github.com/taqiyyaghazi/ecosystem-engine/internal/features/auth/usecase"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/platform/http/httputil"
)

const sessionCookieName = "session_id"

type AuthHandler struct {
	usecase         authUsecase.AuthUsecase
	activityUsecase activityUsecase.ActivityUsecase
}

func NewAuthHandler(usecase authUsecase.AuthUsecase, activityUsecase activityUsecase.ActivityUsecase) *AuthHandler {
	return &AuthHandler{
		usecase:         usecase,
		activityUsecase: activityUsecase,
	}
}

func (h *AuthHandler) RegisterRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.GET("/me", RequireAuth(h.usecase), h.Me)
		auth.POST("/logout", RequireAuth(h.usecase), h.Logout)
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req authDto.RegisterRequest
	if !httputil.BindJSON(c, &req) {
		return
	}

	if err := h.usecase.Register(c.Request.Context(), req); err != nil {
		httputil.HandleError(c, err)
		return
	}

	httputil.NewSuccessResponse(c, http.StatusCreated, "user registered successfully", nil)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req authDto.LoginRequest
	if !httputil.BindJSON(c, &req) {
		return
	}

	userAgent := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()

	sessionID, err := h.usecase.Login(c.Request.Context(), req, userAgent, ipAddress)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		MaxAge:   int(authEntity.SessionTTL.Seconds()),
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	// Log activity
	if sessionData, err := h.usecase.Me(c.Request.Context(), sessionID); err == nil {
		_ = h.activityUsecase.PublishEvent(c.Request.Context(), activityDto.ActivityEvent{
			ActorID:   sessionData.UserID,
			Action:    "LOGIN",
			Payload:   map[string]interface{}{"user_agent": userAgent},
			IPAddress: ipAddress,
		})
	}

	httputil.NewSuccessResponse(c, http.StatusOK, "login successful", gin.H{"session_id": sessionID})
}

func (h *AuthHandler) Me(c *gin.Context) {
	sessionID := c.GetString("session_id")

	res, err := h.usecase.Me(c.Request.Context(), sessionID)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}

	httputil.NewSuccessResponse(c, http.StatusOK, "session data retrieved successfully", res)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID := c.GetString("session_id")

	if err := h.usecase.Logout(c.Request.Context(), sessionID); err != nil {
		httputil.HandleError(c, err)
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	httputil.NewSuccessResponse(c, http.StatusOK, "logout successful", nil)
}
