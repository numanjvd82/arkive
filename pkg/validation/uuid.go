package validation

import (
	"regexp"
	"strings"
)

var uuidPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func NormalizeUUID(value string) (string, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || !uuidPattern.MatchString(trimmed) {
		return "", false
	}
	return strings.ToLower(trimmed), true
}
