package httputil

import (
	"github.com/gin-gonic/gin"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/apperror"
)

// BindJSON binds the request body to the given struct and handles validation errors.
// Returns true if binding was successful, false if an error was returned to the client.
func BindJSON(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		if ve := apperror.FormatValidationErrors(err); ve != nil {
			HandleError(c, ve)
		} else {
			HandleError(c, apperror.NewInvalidInputError(err))
		}
		return false
	}
	return true
}
