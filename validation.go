package crudjt

import (
	"encoding/base64"
	"errors"
	"fmt"
)

const U64Max = ^uint64(0)

const MaxHashSize = 256

const (
	ErrorAlreadyStarted = 0
	ErrorNotStarted = 1
	ErrorSecretKeyNotSet = 2
)

var ErrorMessages = map[int]string{
	ErrorAlreadyStarted: "CRUD_JT already started",
	ErrorNotStarted: "CRUD_JT has not started",
	ErrorSecretKeyNotSet: "Secret key is blank",
}

func ErrorMessage(code int) string {
	if msg, exists := ErrorMessages[code]; exists {
		return msg
	}
	return fmt.Sprintf("Unknown error (%d)", code)
}

func ValidateHashBytesize(hashBytesize int) error {
	if hashBytesize > MaxHashSize {
		return fmt.Errorf("hash can not be bigger than %d bytesize", MaxHashSize)
	}
	return nil
}

func ValidateSecretKey(key string) error {
	decoded, err := base64.StdEncoding.Strict().DecodeString(key)
	if err != nil {
		return errors.New("'secret_key' must be a valid Base64 string")
	}

	size := len(decoded)
	if size != 32 && size != 48 && size != 64 {
		return fmt.Errorf("'secret_key' must be exactly 32, 48, or 64 bytes. Got %d bytes", size)
	}

	return nil
}

func ValidateInsertion(hash *map[string]interface{}, ttl *int, silence_read *int) error {
	if hash == nil {
		return errors.New("must be Hash (map)")
	}

	if ttl != nil {
		if uint64(*ttl) <= 0 || uint64(*ttl) > U64Max {
			return fmt.Errorf("ttl should be greater than 0 and less than 2^64")
		}
	}

	if silence_read != nil {
		if uint64(*silence_read) <= 0 || uint64(*silence_read) > U64Max {
			return fmt.Errorf("silence_read should be greater than 0 and less than 2^64")
		}
	}

	return nil
}

func ValidateToken(token interface{}) error {
	str, ok := token.(string)
	if !ok {
		return errors.New("token must be string")
	}
	if len(str) == 0 {
		return errors.New("token can't be blank")
	}
	return nil
}
