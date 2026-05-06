package delivery

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/apperror"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/services/dto"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/services/usecase"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/platform/http/httputil"
)

type ServiceHandler struct {
	serviceUsecase usecase.ServiceUsecase
	partnerUsecase usecase.PartnerUsecase
}

func NewServiceHandler(serviceUsecase usecase.ServiceUsecase, partnerUsecase usecase.PartnerUsecase) *ServiceHandler {
	return &ServiceHandler{
		serviceUsecase: serviceUsecase,
		partnerUsecase: partnerUsecase,
	}
}

func (h *ServiceHandler) RegisterRoutes(r *gin.RouterGroup) {
	services := r.Group("/services")
	{
		services.POST("", h.Create)
		services.GET("", h.GetAll)
		services.GET("/:id", h.GetByID)
		services.PUT("/:id", h.Update)
		services.DELETE("/:id", h.Delete)
		services.POST("/:id/partners", h.JoinAsPartner)
	}
}

func (h *ServiceHandler) Create(c *gin.Context) {
	var req dto.CreateServiceRequest
	if !httputil.BindJSON(c, &req) {
		return
	}

	res, err := h.serviceUsecase.Create(c.Request.Context(), req)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}

	httputil.NewSuccessResponse(c, http.StatusCreated, "service created successfully", res)
}

func (h *ServiceHandler) GetAll(c *gin.Context) {
	res, cacheStatus, err := h.serviceUsecase.GetAll(c.Request.Context())
	if err != nil {
		httputil.HandleError(c, err)
		return
	}

	httputil.SetCacheHeader(c, cacheStatus)

	httputil.NewSuccessResponse(c, http.StatusOK, "services retrieved successfully", res)
}

func (h *ServiceHandler) GetByID(c *gin.Context) {
	id, ok := httputil.GetUUIDParam(c, "id")
	if !ok {
		return
	}

	res, cacheStatus, err := h.serviceUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}

	httputil.SetCacheHeader(c, cacheStatus)

	httputil.NewSuccessResponse(c, http.StatusOK, "service retrieved successfully", res)
}

func (h *ServiceHandler) Update(c *gin.Context) {
	id, ok := httputil.GetUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateServiceRequest
	if !httputil.BindJSON(c, &req) {
		return
	}

	if err := req.Validate(); err != nil {
		httputil.HandleError(c, apperror.NewInvalidInputError(err))
		return
	}

	res, err := h.serviceUsecase.Update(c.Request.Context(), id, req)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}

	httputil.NewSuccessResponse(c, http.StatusOK, "service updated successfully", res)
}

func (h *ServiceHandler) Delete(c *gin.Context) {
	id, ok := httputil.GetUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.serviceUsecase.Delete(c.Request.Context(), id); err != nil {
		httputil.HandleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ServiceHandler) JoinAsPartner(c *gin.Context) {
	serviceID, ok := httputil.GetUUIDParam(c, "id")
	if !ok {
		return
	}

	userID, ok := httputil.ExtractUserID(c)
	if !ok {
		return
	}

	res, err := h.partnerUsecase.JoinAsPartner(c.Request.Context(), userID, serviceID)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}

	httputil.NewSuccessResponse(c, http.StatusCreated, "joined as partner successfully", res)
}
