package errors

var ERRORS = map[string]func(string) error{
	"XX000": NewInternalError,
	"DE000": NewDonateException,
}
