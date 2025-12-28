package components

import (
	"embed"
	"strings"
)

//go:embed *.css *.js
var embeddedFiles embed.FS

func readAsset(path string) ([]byte, error) {
	cleanPath := strings.TrimPrefix(path, "/")
	if strings.HasPrefix(cleanPath, "web/components/") {
		cleanPath = strings.TrimPrefix(cleanPath, "web/components/")
	}
	return embeddedFiles.ReadFile(cleanPath)
}
