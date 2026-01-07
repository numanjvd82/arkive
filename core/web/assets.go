package web

import (
	"embed"
	"io/fs"
	"net/http"

	"arkive/core/web/components"
)

//go:embed static/*.css static/*.js static/icons/* static/assets/images/* pages/*.css
var embeddedFiles embed.FS

func init() {
	iconFS, err := fs.Sub(embeddedFiles, "static/icons")
	if err != nil {
		return
	}
	components.SetIconFS(iconFS)
}

func StaticFS(dir string) http.FileSystem {
	sub, err := fs.Sub(embeddedFiles, dir)
	if err != nil {
		return http.FS(embeddedFiles)
	}
	return http.FS(sub)
}
