// This binding was generated automatically to ensure consistency across languages
// Generated using ChatGPT (GPT-5) from the canonical Ruby SDK
// API is stable and production-ready

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
