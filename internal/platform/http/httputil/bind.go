package httputil

import (
	"bytes"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/apperror"
)

// BindJSON binds the request body to the given struct and handles validation errors.
// It first checks ALL fields for type mismatches before proceeding with binding.
// Returns true if binding was successful, false if an error was returned to the client.
func BindJSON(c *gin.Context, obj interface{}) bool {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		HandleError(c, apperror.NewInvalidInputErrorWithMessage("failed to read request body"))
		return false
	}

	// Check all fields for type mismatches at once
	if ve := apperror.ValidateJSONTypes(body, obj); ve != nil {
		HandleError(c, ve)
		return false
	}

	// Restore body so ShouldBindJSON can read it for unmarshal + validation
	c.Request.Body = io.NopCloser(bytes.NewReader(body))

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

// BindQuery binds the query parameters to the given struct and handles validation errors.
// Returns true if binding was successful, false if an error was returned to the client.
func BindQuery(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindQuery(obj); err != nil {
		if ve := apperror.FormatValidationErrors(err); ve != nil {
			HandleError(c, ve)
		} else {
			HandleError(c, apperror.NewInvalidInputError(err))
		}
		return false
	}

	return true
}
