// Package assets embeds the compiled React frontend into the Go binary.
package assets

import (
	"embed"
	"io/fs"
	"log"
)

//go:embed dist
var embedded embed.FS

// Frontend returns a filesystem rooted at the dist/ directory.
func Frontend() fs.FS {
	sub, err := fs.Sub(embedded, "dist")
	if err != nil {
		log.Fatalf("assets: embed sub error: %v", err)
	}
	return sub
}
