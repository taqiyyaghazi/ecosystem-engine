package delivery

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	authDto "github.com/taqiyyaghazi/ecosystem-engine/internal/features/auth/dto"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/discovery/dto"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/discovery/usecase"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/platform/http/httputil"
)

// DiscoveryHandler wires the discovery use case to HTTP endpoints.
type DiscoveryHandler struct {
	usecase usecase.DiscoveryUsecase
}

// NewDiscoveryHandler constructs a DiscoveryHandler.
func NewDiscoveryHandler(uc usecase.DiscoveryUsecase) *DiscoveryHandler {
	return &DiscoveryHandler{usecase: uc}
}

// RegisterRoutes mounts the discovery endpoints under the provided router group.
func (h *DiscoveryHandler) RegisterRoutes(r *gin.RouterGroup) {
	discovery := r.Group("/discovery")
	{
		discovery.POST("/location", h.UpdateLocation)
		discovery.POST("/offline", h.SetOffline)
		discovery.GET("/nearby", h.GetNearby)
	}
}

// extractUserID helper securely retrieves the user ID from the Gin context session.
func extractUserID(c *gin.Context) (string, bool) {
	sessionData, exists := c.Get("session")
	if !exists {
		return "", false
	}
	sessionObj, ok := sessionData.(*authDto.SessionResponse)
	if !ok {
		return "", false
	}
	return sessionObj.UserID, true
}

// UpdateLocation handles POST /v1/discovery/location.
func (h *DiscoveryHandler) UpdateLocation(c *gin.Context) {
	userID, ok := extractUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.LocationRequest
	if !httputil.BindJSON(c, &req) {
		return
	}

	if err := h.usecase.UpdateLocation(c.Request.Context(), userID, req); err != nil {
		httputil.HandleError(c, err)
		return
	}

	httputil.NewSuccessResponse(c, http.StatusOK, "location updated successfully", nil)
}

// SetOffline handles POST /v1/discovery/offline.
func (h *DiscoveryHandler) SetOffline(c *gin.Context) {
	userID, ok := extractUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := h.usecase.SetOffline(c.Request.Context(), userID); err != nil {
		httputil.HandleError(c, err)
		return
	}

	httputil.NewSuccessResponse(c, http.StatusOK, "partner is now offline", nil)
}

// GetNearby handles GET /v1/discovery/nearby
func (h *DiscoveryHandler) GetNearby(c *gin.Context) {
	latStr := c.Query("lat")
	lonStr := c.Query("lon")

	if latStr == "" || lonStr == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "lat and lon query parameters are required"})
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "lat must be a valid float"})
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "lon must be a valid float"})
		return
	}

	radiusStr := c.DefaultQuery("radius", "5")
	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil || radius <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "radius must be a positive number"})
		return
	}

	unit := c.DefaultQuery("unit", "km")
	if unit != "m" && unit != "km" && unit != "mi" && unit != "ft" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "unit must be one of: m, km, mi, ft"})
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "limit must be a positive integer"})
		return
	}

	results, err := h.usecase.GetNearby(c.Request.Context(), lat, lon, radius, unit, limit)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}

	httputil.NewSuccessResponse(c, http.StatusOK, "nearby partners retrieved successfully", results)
}
