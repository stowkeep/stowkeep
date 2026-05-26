// Package web serves the embedded Stowkeep frontend assets.
package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed dist/*
var dist embed.FS

// Handler returns an http.Handler that serves the SPA with cache headers.
func Handler() http.Handler {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		panic("web: embed dist: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}
		if strings.HasSuffix(path, ".html") {
			w.Header().Set("Cache-Control", "no-cache")
		} else if strings.Contains(path, "-") && (strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".css")) {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		if _, err := fs.Stat(sub, strings.TrimPrefix(path, "/")); err != nil {
			r.URL.Path = "/index.html"
			w.Header().Set("Cache-Control", "no-cache")
		} else {
			r.URL.Path = path
		}
		fileServer.ServeHTTP(w, r)
	})
}
