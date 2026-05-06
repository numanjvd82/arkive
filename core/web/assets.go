package web

import (
	"embed"
	"io/fs"
	"net/http"

	"arkive/core/web/components"
)

//go:embed static/*.css static/*.js static/*.txt static/*.xml static/icons/* static/assets/images/* static/vendor/arkive-crypto/* static/vendor/plyr/* static/features/*.js static/workers/*.js pages/*.css
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
