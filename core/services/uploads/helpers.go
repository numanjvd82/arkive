package uploads

import (
	"context"
	"strings"
	"time"

	"arkive/core/models"
	"arkive/pkg/validation"
)

const encryptedChunkEnvelopeOverheadBytes int64 = 41
const thumbnailMaxEncryptedBytes int64 = 150 * 1024
const thumbnailMimeWebP = "image/webp"

func validateUserID(userID string) (string, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", ErrUnauthorized
	}
	return userID, nil
}

func validateUploadID(uploadID string) (string, error) {
	uploadID = strings.TrimSpace(uploadID)
	if uploadID == "" {
		return "", ErrInvalidInput
	}
	return uploadID, nil
}

func validateOptionalFolderID(folderID *string) (*string, error) {
	if folderID == nil {
		return nil, nil
	}
	normalized, ok := validation.NormalizeUUID(*folderID)
	if !ok {
		return nil, ErrInvalidInput
	}
	return &normalized, nil
}

func expiresAtPtr(t time.Time) *time.Time {
	return &t
}

func encryptedChunkSize(plaintextSize int64) int64 {
	if plaintextSize <= 0 {
		return 0
	}
	return plaintextSize + encryptedChunkEnvelopeOverheadBytes
}

func encryptedFileSize(plaintextSize int64, chunkCount int) int64 {
	if plaintextSize <= 0 || chunkCount <= 0 {
		return 0
	}
	return plaintextSize + int64(chunkCount)*encryptedChunkEnvelopeOverheadBytes
}

func reservedUploadSize(plaintextSize int64, chunkCount int) int64 {
	return encryptedFileSize(plaintextSize, chunkCount) + thumbnailMaxEncryptedBytes
}

func totalStoredSize(file models.File) int64 {
	total := file.ActualEncryptedSize
	if file.ThumbnailSizeBytes > 0 {
		total += file.ThumbnailSizeBytes
	}
	return total
}

func objectSizeWithRetry(ctx context.Context, measure func(context.Context) (int64, error)) (int64, error) {
	backoffs := []time.Duration{0, 200 * time.Millisecond, 500 * time.Millisecond}
	var lastErr error
	for _, backoff := range backoffs {
		if backoff > 0 {
			select {
			case <-ctx.Done():
				return 0, ctx.Err()
			case <-time.After(backoff):
			}
		}
		size, err := measure(ctx)
		if err == nil {
			return size, nil
		}
		lastErr = err
	}
	return 0, lastErr
}
