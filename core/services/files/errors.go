package files

import "errors"

var (
	ErrUnauthorized    = errors.New("unauthorized")
	ErrNotFound        = errors.New("not found")
	ErrInvalidInput    = errors.New("invalid input")
	ErrUploadCancelled = errors.New("upload cancelled")
)
