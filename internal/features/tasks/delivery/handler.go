package delivery

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/tasks/dto"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/tasks/usecase"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/platform/http/httputil"
)

type TaskHandler struct {
	usecase usecase.TaskUsecase
}

func NewTaskHandler(u usecase.TaskUsecase) *TaskHandler {
	return &TaskHandler{
		usecase: u,
	}
}

func (h *TaskHandler) Route(r *gin.RouterGroup) {
	tasks := r.Group("/tasks")
	{
		tasks.POST("/invoice", h.SimulateInvoice)
	}
}

func (h *TaskHandler) SimulateInvoice(c *gin.Context) {
	var req dto.CreateInvoiceTaskRequest
	if !httputil.BindJSON(c, &req) {
		return
	}

	if err := h.usecase.DispatchInvoiceTask(c.Request.Context(), &req); err != nil {
		httputil.HandleError(c, err)
		return
	}

	httputil.NewSuccessResponse(c, http.StatusAccepted, "invoice generation task dispatched", nil)
}
