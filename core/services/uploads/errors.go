package uploads

import "errors"

var (
	ErrInvalidInput         = errors.New("invalid input")
	ErrNotFound             = errors.New("not found")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrUploadFailed         = errors.New("upload failed")
	ErrStorageLimitExceeded = errors.New("storage limit exceeded")
	ErrFileTooLarge         = errors.New("file is too large")
	ErrFileTooSmall         = errors.New("file is too small")
	ErrFileLimitReached     = errors.New("file limit reached")
	ErrQueueLimitReached    = errors.New("upload queue limit reached")
	ErrUploadCancelled      = errors.New("upload cancelled")
	ErrPartsRequired        = errors.New("all upload parts are required")
)

var (
	ErrFilenameRequired = errors.New("filename is required")
	ErrFileSizeRequired = errors.New("file size is required")
)

var ErrSizeRequired = errors.New("file size is required")
