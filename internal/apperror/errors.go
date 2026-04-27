package apperror

import "errors"

var (
	ErrNotFound     = errors.New("not found")
	ErrInvalidUUID  = errors.New("invalid UUID")
	ErrInvalidInput = errors.New("invalid input")
)
