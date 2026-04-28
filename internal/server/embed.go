package server

import (
	"embed"
	"io/fs"
)

// frontendFiles holds the built Svelte SPA. The frontend is built into
// ../../frontend/dist/ and embedded into the binary at compile time.
//
// During Phase 0 the dist/ directory contains only a placeholder index.html;
// the real SPA lands in Phase 1.
//
//go:embed all:frontend_dist
var frontendFiles embed.FS

// FrontendFS returns a sub-filesystem rooted at the build output.
func FrontendFS() (fs.FS, error) {
	return fs.Sub(frontendFiles, "frontend_dist")
}
