package auth

import (
	"net/mail"
	"strings"
	"unicode/utf8"
)

const (
	maxNameLength     = 100
	maxEmailLength    = 254
	minPasswordLength = 10
	maxPasswordLength = 128
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validateName(name string) error {
	length := utf8.RuneCountInString(name)
	if length == 0 || length > maxNameLength {
		return &ValidationError{Field: "name", Message: "Name must be between 1 and 100 characters."}
	}
	return nil
}

func validateEmail(email string) error {
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || len(email) > maxEmailLength {
		return &ValidationError{Field: "email", Message: "Enter a valid email address."}
	}
	return nil
}

func validatePassword(password string) error {
	length := utf8.RuneCountInString(password)
	if length < minPasswordLength || length > maxPasswordLength {
		return &ValidationError{Field: "password", Message: "Password must be between 10 and 128 characters."}
	}
	return nil
}
