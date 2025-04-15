package crud_jt

import (
	"errors"
	"fmt"
	// "math"
)

// U64Max is the maximum token for an unsigned 64-bit integer.
const U64Max = ^uint64(0) // або math.MaxUint64

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
