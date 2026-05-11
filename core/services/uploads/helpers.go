package uploads

import (
	"strings"
	"time"
)

const encryptedChunkEnvelopeOverheadBytes int64 = 41

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
