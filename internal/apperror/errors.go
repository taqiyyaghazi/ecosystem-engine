package apperror

import "errors"

// Sentinel errors for domain-level conditions.
var (
	ErrNotFound    = errors.New("not found")
	ErrInvalidUUID = errors.New("invalid UUID")
)
