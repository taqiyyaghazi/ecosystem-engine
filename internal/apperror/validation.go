package apperror

import (
	"encoding/json"
	"fmt"
	"reflect"
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

// ValidateJSONTypes decodes body into a generic map and checks each field's
// actual JSON type against the expected Go struct type using reflection.
// Returns a ValidationError listing ALL type mismatches, or nil if types are correct.
func ValidateJSONTypes(body []byte, target interface{}) *ValidationError {
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil // syntax errors are handled elsewhere
	}

	t := reflect.TypeOf(target)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}

	var fields []ValidationFieldError
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)

		jsonTag := sf.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		jsonName := strings.Split(jsonTag, ",")[0]

		val, exists := raw[jsonName]
		if !exists || val == nil {
			continue // missing/null fields are handled by validator tags (e.g. required)
		}

		if !isTypeCompatible(sf.Type, val) {
			fields = append(fields, ValidationFieldError{
				Field:   jsonName,
				Message: fmt.Sprintf("%s must be of type %s", jsonName, friendlyTypeName(sf.Type)),
			})
		}
	}

	if len(fields) == 0 {
		return nil
	}
	return &ValidationError{Fields: fields}
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
	case "latitude":
		return fmt.Sprintf("%s must be a valid latitude (-90 to 90)", field)
	case "longitude":
		return fmt.Sprintf("%s must be a valid longitude (-180 to 180)", field)
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

// isTypeCompatible checks whether the actual JSON-decoded value matches the expected Go type.
func isTypeCompatible(expected reflect.Type, actual interface{}) bool {
	switch expected.Kind() {
	case reflect.String:
		_, ok := actual.(string)
		return ok
	case reflect.Bool:
		_, ok := actual.(bool)
		return ok
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		_, ok := actual.(float64) // JSON numbers decode as float64
		return ok
	case reflect.Float32, reflect.Float64:
		_, ok := actual.(float64)
		return ok
	case reflect.Slice:
		_, ok := actual.([]interface{})
		return ok
	case reflect.Map, reflect.Struct:
		_, ok := actual.(map[string]interface{})
		return ok
	default:
		return true // unknown types pass through
	}
}

// friendlyTypeName returns a human-readable name for the expected Go type.
func friendlyTypeName(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Slice:
		return "array"
	case reflect.Map, reflect.Struct:
		return "object"
	default:
		return t.String()
	}
}
