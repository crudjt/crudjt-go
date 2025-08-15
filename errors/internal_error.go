package errors

import "fmt"

// InternalError — аналог MyLib::Errors::InternalError
type InternalError struct {
	Message string
}

func (e *InternalError) Error() string {
	return fmt.Sprintf("InternalError: %s", e.Message)
}

// NewInternalError — конструктор
func NewInternalError(msg string) error {
	return &InternalError{Message: msg}
}
