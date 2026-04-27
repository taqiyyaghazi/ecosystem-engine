package httputil

import "github.com/gin-gonic/gin"

type ApiResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message"`
}

func NewSuccessResponse(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, ApiResponse{
		Data:    data,
		Message: message,
	})
}
