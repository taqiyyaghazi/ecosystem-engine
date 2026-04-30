package apperror

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidationFieldError represents a single field validation error.
type ValidationFieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationError wraps multiple field validation errors into a structured error.
type ValidationError struct {
	Fields []ValidationFieldError
}

func (e *ValidationError) Error() string {
	msgs := make([]string, 0, len(e.Fields))
	for _, f := range e.Fields {
		msgs = append(msgs, fmt.Sprintf("%s: %s", f.Field, f.Message))
	}
	return "validation failed: " + strings.Join(msgs, "; ")
}

func (e *ValidationError) Is(target error) bool {
	return target == ErrInvalidInput
}

// FormatValidationErrors converts validator.ValidationErrors into a structured ValidationError.
// If the error is not a validator.ValidationErrors, it returns nil.
func FormatValidationErrors(err error) *ValidationError {
	ve, ok := err.(validator.ValidationErrors)
	if !ok {
		return nil
	}

	fields := make([]ValidationFieldError, 0, len(ve))
	for _, fe := range ve {
		fields = append(fields, ValidationFieldError{
			Field:   toSnakeCase(fe.Field()),
			Message: buildMessage(fe),
		})
	}

	return &ValidationError{Fields: fields}
}

// buildMessage creates a human-readable message from a validator.FieldError.
func buildMessage(fe validator.FieldError) string {
	field := strings.ToLower(fe.Field())

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, fe.Param())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, fe.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, fe.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, fe.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, fe.Param())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only alphanumeric characters", field)
	default:
		return fmt.Sprintf("%s failed validation: %s", field, fe.Tag())
	}
}

// toSnakeCase converts a PascalCase field name to snake_case.
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteByte('_')
			}
			result.WriteRune(r + 32) // to lowercase
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
