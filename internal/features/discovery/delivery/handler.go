package delivery

import (
	"net/http"

	"github.com/gin-gonic/gin"
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

// UpdateLocation handles POST /v1/discovery/location.
func (h *DiscoveryHandler) UpdateLocation(c *gin.Context) {
	userID, ok := httputil.ExtractUserID(c)
	if !ok {
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
	userID, ok := httputil.ExtractUserID(c)
	if !ok {
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
	req := dto.NearbyRequest{
		Radius: 5,
		Unit:   "km",
		Limit:  20,
	}

	if !httputil.BindQuery(c, &req) {
		return
	}

	results, err := h.usecase.GetNearby(c.Request.Context(), req.Lat, req.Lon, req.Radius, req.Unit, req.Limit)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}

	httputil.NewSuccessResponse(c, http.StatusOK, "nearby partners retrieved successfully", results)
}
