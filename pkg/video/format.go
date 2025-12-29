package video

import "fmt"

func FormatResolution(width, height int, isVideo bool) string {
	if width <= 0 || height <= 0 {
		if isVideo {
			return "Processing"
		}
		return "Unknown"
	}
	return fmt.Sprintf("%d×%d", width, height)
}

func FormatDuration(seconds int64, isVideo bool) string {
	if seconds <= 0 {
		if isVideo {
			return "Processing"
		}
		return "Unknown"
	}
	minutes := seconds / 60
	remaining := seconds % 60
	hours := minutes / 60
	minutes = minutes % 60
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm %ds", minutes, remaining)
}
