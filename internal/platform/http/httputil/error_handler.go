package httputil

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/apperror"
)

func HandleError(c *gin.Context, err error) {
	var (
		validationErr *apperror.ValidationError
		invalidInput  *apperror.InvalidInputError
	)

	switch {
	case errors.Is(err, apperror.ErrUnauthorized):
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	case errors.Is(err, apperror.ErrNotFound):
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "resource not found"})
	case errors.Is(err, apperror.ErrInvalidUUID):
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
	case errors.As(err, &validationErr):
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "validation failed",
			"details": validationErr.Fields,
		})
	case errors.As(err, &invalidInput):
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": invalidInput.Error()})
	case errors.Is(err, apperror.ErrInvalidInput):
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
	default:
		slog.Error("Internal server error", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
