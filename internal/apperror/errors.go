package apperror

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrInvalidUUID  = errors.New("invalid UUID")
	ErrInvalidInput = errors.New("invalid input")
)

type InvalidInputError struct {
	Err error
}

func (e *InvalidInputError) Error() string {
	return fmt.Sprintf("invalid input: %v", e.Err)
}

func (e *InvalidInputError) Unwrap() error {
	return e.Err
}

func (e *InvalidInputError) Is(target error) bool {
	return target == ErrInvalidInput
}

func NewInvalidInputError(err error) error {
	return &InvalidInputError{Err: err}
}

func NewInvalidInputErrorWithMessage(msg string) error {
	return &InvalidInputError{Err: errors.New(msg)}
}
