package domain

import "errors"

var (
	ErrNotFound      = errors.New("resource not found")
	ErrConflict      = errors.New("resource already exists")
	ErrUnauthorized  = errors.New("invalid email or password")
	ErrInvalidToken  = errors.New("invalid or expired session token")
	ErrValidation    = errors.New("input validation failure")
	ErrInternalError = errors.New("internal server error")
)