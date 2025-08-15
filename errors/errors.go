package errors

// ERRORS — аналог Ruby-хешу з кодами помилок
var ERRORS = map[string]func(string) error{
	"XX000": NewInternalError,
	"DE000": NewDonateException,
}
