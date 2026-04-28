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
	usecase usecase.ServiceUsecase
}

func NewServiceHandler(usecase usecase.ServiceUsecase) *ServiceHandler {
	return &ServiceHandler{
		usecase: usecase,
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
	}
}

func (h *ServiceHandler) Create(c *gin.Context) {
	var req dto.CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.HandleError(c, apperror.NewInvalidInputError(err))
		return
	}

	res, err := h.usecase.Create(c.Request.Context(), req)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}

	httputil.NewSuccessResponse(c, http.StatusCreated, "service created successfully", res)
}

func (h *ServiceHandler) GetAll(c *gin.Context) {
	res, cacheStatus, err := h.usecase.GetAll(c.Request.Context())
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

	res, cacheStatus, err := h.usecase.GetByID(c.Request.Context(), id)
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
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.HandleError(c, apperror.NewInvalidInputError(err))
		return
	}

	if err := req.Validate(); err != nil {
		httputil.HandleError(c, apperror.NewInvalidInputError(err))
		return
	}

	res, err := h.usecase.Update(c.Request.Context(), id, req)
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

	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		httputil.HandleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
