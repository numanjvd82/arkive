package shares

import "errors"

const (
	FileStatusComplete = "complete"

	TokenMinLength    = 12
	TokenMaxLength    = 128
	PasswordMinLength = 8
)

var (
	ErrUnauthorized       = errors.New("unauthorized")
	ErrNotFound           = errors.New("share not found")
	ErrInvalidInput       = errors.New("invalid input")
	ErrShareExists        = errors.New("share already exists")
	ErrPasswordHashFailed = errors.New("password hashing failed")

	ErrFileIDRequired   = errors.New("file id is required")
	ErrPasswordRequired = errors.New("password is required")
	ErrTokenInvalid     = errors.New("token must be url-safe and 12-128 chars")
	ErrExpiryInvalid    = errors.New("expiry must be in the future")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
)
