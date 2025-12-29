package video

import "time"

const (
	LargeSizeBytes       int64 = 3 * 1024 * 1024 * 1024
	LargeDurationSeconds int64 = int64(80 * time.Minute / time.Second)
	MaxWidth                   = 1920
	MaxHeight                  = 1080
)

func IsLarge(sizeBytes int64, durationSeconds int64, width int, height int) bool {
	bytesCondition := sizeBytes >= LargeSizeBytes
	durationCondition := durationSeconds >= LargeDurationSeconds
	dimensionsCondition := width > MaxWidth || height > MaxHeight

	if bytesCondition && durationCondition && dimensionsCondition {
		return true
	}
	return false
}
