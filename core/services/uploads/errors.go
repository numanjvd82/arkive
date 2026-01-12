package uploads

import "errors"

var (
	ErrInvalidInput      = errors.New("invalid input")
	ErrNotFound          = errors.New("not found")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrUploadFailed      = errors.New("upload failed")
	ErrQuotaExceeded     = errors.New("quota exceeded")
	ErrFileTooLarge      = errors.New("file is too large")
	ErrFileTooSmall      = errors.New("file is too small")
	ErrMultipartRequired = errors.New("multipart upload required")
	ErrFileLimitReached  = errors.New("file limit reached")
	ErrQueueLimitReached = errors.New("upload queue limit reached")
	ErrConcurrentLimit   = errors.New("upload already in progress")
)

var (
	ErrFilenameRequired  = errors.New("filename is required")
	ErrFileSizeRequired  = errors.New("file size is required")
	ErrPartNumberInvalid = errors.New("part number is invalid")
	ErrPartsRequired     = errors.New("parts are required")
	ErrNoNextPart        = errors.New("no next part available")
	ErrUploadCancelled   = errors.New("upload cancelled")
)

type MissingPartsError struct {
	Missing []int32
}

func (e MissingPartsError) Error() string {
	return "missing parts"
}
