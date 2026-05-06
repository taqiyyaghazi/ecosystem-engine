package delivery

import (
	"net/http"

	"github.com/gin-gonic/gin"
	activityDto "github.com/taqiyyaghazi/ecosystem-engine/internal/features/activity/dto"
	activityUsecase "github.com/taqiyyaghazi/ecosystem-engine/internal/features/activity/usecase"
	discoveryDto "github.com/taqiyyaghazi/ecosystem-engine/internal/features/discovery/dto"
	discoveryUsecase "github.com/taqiyyaghazi/ecosystem-engine/internal/features/discovery/usecase"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/platform/http/httputil"
)

// DiscoveryHandler wires the discovery use case to HTTP endpoints.
type DiscoveryHandler struct {
	usecase         discoveryUsecase.DiscoveryUsecase
	activityUsecase activityUsecase.ActivityUsecase
}

// NewDiscoveryHandler constructs a DiscoveryHandler.
func NewDiscoveryHandler(uc discoveryUsecase.DiscoveryUsecase, activityUc activityUsecase.ActivityUsecase) *DiscoveryHandler {
	return &DiscoveryHandler{
		usecase:         uc,
		activityUsecase: activityUc,
	}
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

	var req discoveryDto.LocationRequest
	if !httputil.BindJSON(c, &req) {
		return
	}

	if err := h.usecase.UpdateLocation(c.Request.Context(), userID, req); err != nil {
		httputil.HandleError(c, err)
		return
	}

	// Log activity
	_ = h.activityUsecase.PublishEvent(c.Request.Context(), activityDto.ActivityEvent{
		ActorID:   userID,
		Action:    "UPDATE_LOCATION",
		Payload:   map[string]interface{}{"lat": req.Latitude, "lon": req.Longitude},
		IPAddress: c.ClientIP(),
	})

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
	req := discoveryDto.NearbyRequest{
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
