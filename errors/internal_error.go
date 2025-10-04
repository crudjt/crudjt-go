package errors

import "fmt"

type InternalError struct {
	Message string
}

func (e *InternalError) Error() string {
	return fmt.Sprintf("InternalError: %s", e.Message)
}

func NewInternalError(msg string) error {
	return &InternalError{Message: msg}
}
