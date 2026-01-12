package pagination

import (
	"strconv"
	"strings"
)

func ParsePageParam(value string) int {
	page, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || page < 1 {
		return 1
	}
	return page
}

func ParsePageSizeParam(value string) int {
	size, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || size <= 0 {
		return 25
	}
	switch size {
	case 25, 50, 100:
		return size
	default:
		return 25
	}
}
