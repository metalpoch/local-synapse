package validator

import (
	"errors"
	"regexp"
)

var (
    ErrPasswordTooShort    = errors.New("password must be at least 8 characters long")
    ErrPasswordsMismatch   = errors.New("passwords do not match")
    ErrNoSpecialChar       = errors.New("password must contain at least one special character")
    ErrNoNumber            = errors.New("password must contain at least one number")
    ErrNoUppercase         = errors.New("password must contain at least one uppercase letter")
    ErrNoLowercase         = errors.New("password must contain at least one lowercase letter")
)

func Password(password string) error {
    if len(password) < 8 {
        return ErrPasswordTooShort
    }
    if !regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password) {
        return ErrNoSpecialChar
    }
    if !regexp.MustCompile(`\d`).MatchString(password) {
        return ErrNoNumber
    }
    if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
        return ErrNoUppercase
    }
    if !regexp.MustCompile(`[a-z]`).MatchString(password) {
        return ErrNoLowercase
    }
    
    return nil
}
