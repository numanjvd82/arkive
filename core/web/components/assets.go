package components

import (
	"embed"
	"strings"
)

//go:embed *.css
var embeddedFiles embed.FS

func readAsset(path string) ([]byte, error) {
	cleanPath := strings.TrimPrefix(path, "/")
	if after, ok := strings.CutPrefix(cleanPath, "web/components/"); ok {
		cleanPath = after
	}
	return embeddedFiles.ReadFile(cleanPath)
}
