package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestWithNotFound(t *testing.T) {
	fsys := http.FS(fstest.MapFS{
		"index.html": {Data: []byte("home")},
		"404.html":   {Data: []byte("<title>Page not found</title>")},
		"app.js":     {Data: []byte("js")},
		"sw.js":      {Data: []byte("sw")},
		"assets/a-1.js": {Data: []byte("chunk")},
	})
	h := withNotFound(fsys)
	for _, c := range []struct {
		path string
		code int
		body string
	}{
		{"/", 200, "home"},
		{"/app.js", 200, "js"},
		{"/nope", 404, "Page not found"},
		{"/a/b/../../missing.png", 404, "Page not found"},
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", c.path, nil))
		if rec.Code != c.code || !strings.Contains(rec.Body.String(), c.body) {
			t.Errorf("%s: got %d %q", c.path, rec.Code, rec.Body.String())
		}
	}
	for path, want := range map[string]string{
		"/":             "no-cache",
		"/sw.js":        "no-cache",
		"/assets/a-1.js": "public, max-age=31536000, immutable",
		"/app.js":       "",
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		if got := rec.Header().Get("Cache-Control"); got != want {
			t.Errorf("%s: Cache-Control %q, want %q", path, got, want)
		}
	}
}
