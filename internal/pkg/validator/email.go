package validator

import (
	"errors"
	"regexp"
)

var (
	ErrEmailEmpty   = errors.New("email cannot be empty")
	ErrEmailInvalid = errors.New("email format is invalid")
	ErrEmailTooLong = errors.New("email cannot exceed 254 characters")
)

func Email(email string) error {
	if email == "" {
		return ErrEmailEmpty
	}

	if len(email) > 254 {
		return ErrEmailTooLong
	}

	// RFC 5322
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return ErrEmailInvalid
	}

	return nil
}
