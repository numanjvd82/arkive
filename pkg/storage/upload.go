package storage

import (
	"fmt"

	"arkive/pkg/tokens"
)

func BuildObjectKey(userID string) (string, error) {
	token, _, err := tokens.Generate()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("u/%s/%s", userID, token), nil
}

func ChooseChunkSize(sizeBytes int64, maxFileSizeBytes int64) int {
	chunk := 10 * 1024 * 1024
	if sizeBytes > maxFileSizeBytes {
		chunk = 25 * 1024 * 1024
	}
	return chunk
}

func TotalParts(sizeBytes int64, chunkSize int) int {
	return int((sizeBytes + int64(chunkSize) - 1) / int64(chunkSize))
}
