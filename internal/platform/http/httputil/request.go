package httputil

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/apperror"
)

func GetUUIDParam(c *gin.Context, paramName string) (string, bool) {
	id := c.Param(paramName)
	if _, err := uuid.Parse(id); err != nil {
		HandleError(c, apperror.NewInvalidInputErrorWithMessage("invalid id format"))
		return "", false
	}
	return id, true
}
