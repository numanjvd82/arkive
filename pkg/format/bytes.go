package format

import "fmt"

func Bytes(size int64) string {
	if size <= 0 {
		return "0 B"
	}
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	if size >= GB {
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	}
	if size >= MB {
		return fmt.Sprintf("%.1f MB", float64(size)/float64(MB))
	}
	if size >= KB {
		return fmt.Sprintf("%.1f KB", float64(size)/float64(KB))
	}
	return fmt.Sprintf("%d B", size)
}
