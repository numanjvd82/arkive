package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static/*.css static/*.js static/icons/* pages/*.css
var embeddedFiles embed.FS

func StaticFS(dir string) http.FileSystem {
	sub, err := fs.Sub(embeddedFiles, dir)
	if err != nil {
		return http.FS(embeddedFiles)
	}
	return http.FS(sub)
}
