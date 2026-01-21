package web

import (
	"io/fs"
	"net/http"
	"strings"
)

// SPAHandler handles Single Page Application routing
// It serves static files and falls back to index.html for client-side routing
type SPAHandler struct {
	staticFS   http.Handler
	fileSystem fs.FS
}

// NewSPAHandler creates a new SPA handler with the given filesystem
func NewSPAHandler(fsys fs.FS) *SPAHandler {
	return &SPAHandler{
		staticFS:   http.FileServer(http.FS(fsys)),
		fileSystem: fsys,
	}
}

// ServeHTTP implements http.Handler
// For existing files, it serves them directly
// For non-existent files (client-side routes), it serves index.html
func (h *SPAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		path = "index.html"
	}

	// Check if the file exists
	_, err := fs.Stat(h.fileSystem, path)
	if err != nil {
		// File not found, serve index.html for SPA routing
		r.URL.Path = "/"
	}

	h.staticFS.ServeHTTP(w, r)
}
