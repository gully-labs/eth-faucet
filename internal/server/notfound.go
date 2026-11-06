package server

import (
	"io"
	"net/http"
	"path"
	"strings"
)

// withNotFound serves the static app and answers unknown paths with the branded web/public/404.html (status 404,
// the Gully Chain header and links to every key surface) instead of Go's plain "404 page not found".
func withNotFound(fsys http.FileSystem) http.Handler {
	files := http.FileServer(fsys)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := path.Clean("/" + r.URL.Path)
		if p != "/" {
			f, err := fsys.Open(p)
			if err == nil {
				f.Close()
			} else if page, perr := fsys.Open("/404.html"); perr == nil {
				defer page.Close()
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Header().Set("Cache-Control", "no-cache")
				w.WriteHeader(http.StatusNotFound)
				_, _ = io.Copy(w, page)
				return
			}
		}
		switch {
		case strings.HasPrefix(p, "/assets/"):
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable") // hashed names
		case p == "/" || p == "/sw.js" || p == "/sw-config.json":
			w.Header().Set("Cache-Control", "no-cache") // revalidated, so a deploy and the kill switch are seen
		}
		files.ServeHTTP(w, r)
	})
}
