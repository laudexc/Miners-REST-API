package myHttp

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static
var staticFiles embed.FS

func webHandler() http.Handler {
	static, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return http.NotFoundHandler()
	}

	return http.FileServer(http.FS(static))
}
