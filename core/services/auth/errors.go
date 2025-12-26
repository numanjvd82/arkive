package auth

import "errors"

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidInput        = errors.New("invalid input")
	ErrEmailExists         = errors.New("email already exists")
	ErrBrandNameExists     = errors.New("brand name already exists")
	ErrSessionNotFound     = errors.New("session not found")

	ErrEmailRequired           = errors.New("email is required")
	ErrPasswordRequired        = errors.New("password is required")
	ErrBrandNameRequired       = errors.New("brand name is required")
	ErrConfirmPasswordRequired = errors.New("confirm password is required")
	ErrPasswordTooShort        = errors.New("password must be at least 8 characters")
	ErrPasswordMissingLower    = errors.New("password must include a lowercase letter")
	ErrPasswordMissingUpper    = errors.New("password must include an uppercase letter")
	ErrPasswordMissingSymbol   = errors.New("password must include a symbol")
	ErrPasswordMismatch        = errors.New("passwords do not match")
	ErrLoginInvalid            = errors.New("invalid email or password")
)
