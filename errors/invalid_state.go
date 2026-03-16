// This binding was generated automatically to ensure consistency across languages
// Generated using ChatGPT (GPT-5) from the canonical Ruby SDK
// API is stable and production-ready

package errors

import "fmt"

type InvalidState struct {
	Message string
}

func (e *InvalidState) Error() string {
	return fmt.Sprintf("InvalidState: %s", e.Message)
}

func NewInvalidState(msg string) error {
	return &InvalidState{Message: msg}
}
