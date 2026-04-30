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
