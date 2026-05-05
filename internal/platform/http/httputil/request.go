package httputil

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/apperror"
)

func GetUUIDParam(c *gin.Context, paramName string) (string, bool) {
	id := c.Param(paramName)
	if _, err := uuid.Parse(id); err != nil {
		HandleError(c, apperror.ErrInvalidUUID)
		return "", false
	}
	return id, true
}

func ExtractUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		HandleError(c, apperror.ErrUnauthorized)
		return "", false
	}
	id, ok := userID.(string)
	if !ok {
		HandleError(c, apperror.ErrUnauthorized)
		return "", false
	}
	return id, true
}
