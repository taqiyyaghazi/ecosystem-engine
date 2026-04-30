package httputil

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/apperror"
)

func HandleError(c *gin.Context, err error) {
	var invalidInput *apperror.InvalidInputError

	switch {
	case errors.Is(err, apperror.ErrNotFound):
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "resource not found"})
	case errors.Is(err, apperror.ErrInvalidUUID):
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
	case errors.As(err, &invalidInput), errors.Is(err, apperror.ErrInvalidInput):
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
	default:
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
