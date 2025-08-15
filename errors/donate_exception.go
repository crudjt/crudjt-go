package errors

import "fmt"

type DonateException struct {
	Message string
}

func (e *DonateException) Error() string {
	return fmt.Sprintf("DonateException: %s", e.Message)
}

func NewDonateException(msg string) error {
	return &DonateException{Message: msg}
}
