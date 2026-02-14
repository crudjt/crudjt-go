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
