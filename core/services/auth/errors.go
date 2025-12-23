package auth

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailExists        = errors.New("email already exists")
	ErrBrandNameExists    = errors.New("brand name already exists")
	ErrRefreshTokenInvalid = errors.New("refresh token invalid")
	ErrSessionNotFound    = errors.New("session not found")
)
