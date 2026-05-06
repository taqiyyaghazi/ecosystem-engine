package delivery

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/leaderboard/dto"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/leaderboard/usecase"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/platform/http/httputil"
)

type LeaderboardHandler struct {
	leaderboardUseCase usecase.LeaderboardUseCase
}

func NewLeaderboardHandler(leaderboardUseCase usecase.LeaderboardUseCase) *LeaderboardHandler {
	return &LeaderboardHandler{leaderboardUseCase: leaderboardUseCase}
}

func (h *LeaderboardHandler) RegisterRoutes(public, protected *gin.RouterGroup) {
	leaderboardPublic := public.Group("/leaderboard")
	{
		leaderboardPublic.GET("", h.GetLeaderboard)
	}

	leaderboardProtected := protected.Group("/leaderboard")
	{
		leaderboardProtected.POST("/points", h.AddPoints)
		leaderboardProtected.GET("/me", h.GetPartnerRank)
	}
}

func (h *LeaderboardHandler) AddPoints(c *gin.Context) {
	var req dto.PointRequest
	if !httputil.BindJSON(c, &req) {
		return
	}

	err := h.leaderboardUseCase.AddPoints(c.Request.Context(), req)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}

	httputil.NewSuccessResponse(c, http.StatusOK, "points added successfully", nil)
}

func (h *LeaderboardHandler) GetLeaderboard(c *gin.Context) {
	req := dto.LeaderboardQuery{
		Limit: 10,
	}

	if !httputil.BindQuery(c, &req) {
		return
	}

	resp, err := h.leaderboardUseCase.GetLeaderboard(c.Request.Context(), req.ServiceID, req.Limit)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}

	httputil.NewSuccessResponse(c, http.StatusOK, "leaderboard retrieved successfully", resp)
}

func (h *LeaderboardHandler) GetPartnerRank(c *gin.Context) {
	userID, ok := httputil.ExtractUserID(c)
	if !ok {
		return
	}

	var query dto.PartnerRankQuery
	if !httputil.BindQuery(c, &query) {
		return
	}

	resp, err := h.leaderboardUseCase.GetPartnerRank(c.Request.Context(), userID, query.ServiceID)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}

	httputil.NewSuccessResponse(c, http.StatusOK, "partner rank retrieved successfully", resp)
}
