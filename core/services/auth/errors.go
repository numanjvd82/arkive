package auth

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidInput       = errors.New("invalid input")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrSessionNotFound    = errors.New("session not found")
	ErrVaultNotConfigured = errors.New("vault is not configured")

	ErrEmailRequired    = errors.New("email is required")
	ErrPasswordRequired = errors.New("password is required")
	ErrLoginInvalid     = errors.New("invalid email or password")
)
