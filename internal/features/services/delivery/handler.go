package delivery

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		httputil.HandleError(c, apperror.ErrInvalidInput)
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
	res, err := h.usecase.GetAll(c.Request.Context())
	if err != nil {
		httputil.HandleError(c, err)
		return
	}

	httputil.NewSuccessResponse(c, http.StatusOK, "services retrieved successfully", res)
}

func (h *ServiceHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		httputil.HandleError(c, apperror.ErrInvalidUUID)
		return
	}

	res, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}

	httputil.NewSuccessResponse(c, http.StatusOK, "service retrieved successfully", res)
}

func (h *ServiceHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		httputil.HandleError(c, apperror.ErrInvalidUUID)
		return
	}

	var req dto.UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.HandleError(c, apperror.ErrInvalidInput)
		return
	}

	if err := req.Validate(); err != nil {
		httputil.HandleError(c, apperror.ErrInvalidInput)
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
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		httputil.HandleError(c, apperror.ErrInvalidUUID)
		return
	}

	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		httputil.HandleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
