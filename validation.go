package crud_jt

import (
	"encoding/base64"
	"errors"
	"fmt"
	// "math"
)

// U64Max is the maximum token for an unsigned 64-bit integer.
const U64Max = ^uint64(0) // або math.MaxUint64

const MaxHashSize = 256

// Коди помилок
const (
	ErrorAlreadyStarted = 0
	ErrorNotStarted = 1
	ErrorEncryptedKeyNotSet = 2
)

// Повідомлення для кодів помилок
var ErrorMessages = map[int]string{
	ErrorAlreadyStarted: "CRUD_JT already started",
	ErrorNotStarted: "CRUD_JT has not started",
	ErrorEncryptedKeyNotSet: "Encrypted key is blank",
}

// ErrorMessage повертає повідомлення за кодом або "Unknown error (...)"
func ErrorMessage(code int) string {
	if msg, exists := ErrorMessages[code]; exists {
		return msg
	}
	return fmt.Sprintf("Unknown error (%d)", code)
}

// ValidateHashBytesize перевіряє розмір хеша
func ValidateHashBytesize(hashBytesize int) error {
	if hashBytesize > MaxHashSize {
		return fmt.Errorf("hash can not be bigger than %d bytesize", MaxHashSize)
	}
	return nil
}

// ValidateEncryptedKey перевіряє ключ у форматі Base64 та довжину
func ValidateEncryptedKey(key string) error {
	decoded, err := base64.StdEncoding.Strict().DecodeString(key)
	if err != nil {
		return errors.New("'encrypted_key' must be a valid Base64 string")
	}

	size := len(decoded)
	if size != 32 && size != 48 && size != 64 {
		return fmt.Errorf("'encrypted_key' must be exactly 32, 48, or 64 bytes. Got %d bytes", size)
	}

	return nil
}

// ValidateInsertion checks the inputs and returns an error if invalid.
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

// Validatetoken checks that a token is a non-empty string.
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
