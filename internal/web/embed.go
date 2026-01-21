// Package web provides embedded static files for the Web UI
package web

import (
	"embed"
	"io/fs"
)

//go:embed dist/*
var embeddedFiles embed.FS

// GetFS returns the embedded filesystem for the web UI
// The dist directory contains the production build of the React frontend
func GetFS() (fs.FS, error) {
	return fs.Sub(embeddedFiles, "dist")
}
