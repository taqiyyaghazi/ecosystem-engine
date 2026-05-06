package dto

type TaskPayload struct {
	TaskID string `json:"task_id"`
	Type   string `json:"type"`
}

type CreateInvoiceTaskRequest struct {
	// Any metadata needed for invoice generation
	// For testing, we might just pass a dummy string
	Description string `json:"description" binding:"required"`
}
