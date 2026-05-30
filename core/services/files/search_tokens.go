package files

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"strings"

	"arkive/core/models"
)

const (
	MaxSearchTokensPerFile = 128
	MaxSearchQueryTokens   = 32
	searchTokenHashBytes   = 32
	searchTokenWeightMin   = 1
	searchTokenWeightMax   = 100
)

var allowedSearchFields = map[string]struct{}{
	"name":   {},
	"prefix": {},
	"ext":    {},
	"mime":   {},
}

func NormalizeSearchTokens(tokens []models.FileSearchToken, maxTokens int) ([]models.FileSearchToken, error) {
	if maxTokens <= 0 {
		maxTokens = MaxSearchTokensPerFile
	}
	if len(tokens) == 0 || len(tokens) > maxTokens {
		return nil, ErrInvalidInput
	}

	normalized := make([]models.FileSearchToken, 0, len(tokens))
	seen := make(map[string]struct{}, len(tokens))
	for _, token := range tokens {
		field := strings.TrimSpace(token.Field)
		if _, ok := allowedSearchFields[field]; !ok {
			return nil, ErrInvalidInput
		}
		if len(token.TokenHash) != searchTokenHashBytes {
			return nil, ErrInvalidInput
		}
		if token.Weight < searchTokenWeightMin || token.Weight > searchTokenWeightMax {
			return nil, ErrInvalidInput
		}
		key := field + ":" + hex.EncodeToString(token.TokenHash)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, models.FileSearchToken{
			TokenHash: bytes.Clone(token.TokenHash),
			Field:     field,
			Weight:    token.Weight,
		})
	}
	if len(normalized) == 0 || len(normalized) > maxTokens {
		return nil, ErrInvalidInput
	}
	return normalized, nil
}

func DecodeSearchTokenString(value string) ([]byte, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, ErrInvalidInput
	}
	if decoded, err := base64.RawURLEncoding.DecodeString(trimmed); err == nil {
		if len(decoded) != searchTokenHashBytes {
			return nil, ErrInvalidInput
		}
		return decoded, nil
	}
	decoded, err := hex.DecodeString(trimmed)
	if err != nil || len(decoded) != searchTokenHashBytes {
		return nil, ErrInvalidInput
	}
	return decoded, nil
}
