package views

import (
	"embed"
	"net/http"
)

//go:embed *.html
var FS embed.FS

func FileSystem() http.FileSystem {
	return http.FS(FS)
}
