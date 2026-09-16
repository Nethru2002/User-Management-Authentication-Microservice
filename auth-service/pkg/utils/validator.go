package utils

import (
	"errors"
	"net/mail"
	"strings"
)

func ValidateAndCleanEmail(email string) (string, error) {
	cleaned := strings.TrimSpace(strings.ToLower(email))
	if cleaned == "" {
		return "", errors.New("email is required")
	}
	addr, err := mail.ParseAddress(cleaned)
	if err != nil || addr.Address != cleaned {
		return "", errors.New("invalid email format")
	}
	return cleaned, nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	// bcrypt limitation check
	if len([]byte(password)) > 72 {
		return errors.New("password cannot exceed 72 bytes")
	}
	return nil
}