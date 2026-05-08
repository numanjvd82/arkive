package storage

import (
	"fmt"
)

func BuildObjectKey(userID, fileID string) (string, error) {
	if userID == "" {
		return "", fmt.Errorf("userID is required")
	}
	if fileID == "" {
		return "", fmt.Errorf("fileID is required")
	}
	return fmt.Sprintf("u/%s/%s", userID, fileID), nil
}
