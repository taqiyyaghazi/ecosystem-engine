package httputil

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/apperror"
)

func HandleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperror.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "resource not found"})
	case errors.Is(err, apperror.ErrInvalidUUID):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
	case errors.Is(err, apperror.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
