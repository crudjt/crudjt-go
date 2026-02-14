package errors

var ERRORS = map[string]func(string) error{
	"XX000": NewInternalError,
	"55JT01": NewInvalidState,
}
