// Package assets provides access to embedded rule, command, and agent files.
// Files under assets/ are compiled into the binary via go:embed and can be
// listed, read, or extracted to a target directory at install time.
package assets

import (
	"embed"
	"io/fs"
	"path/filepath"
	"strings"
)

//go:embed all:rules all:commands all:agents
var embedded embed.FS

//go:embed all:viewer
var viewerFS embed.FS

// ViewerFS returns a filesystem rooted at the viewer/ directory, suitable
// for serving with http.FileServer.
func ViewerFS() (fs.FS, error) {
	return fs.Sub(viewerFS, "viewer")
}

// categories lists the asset directories that are embedded.
var categories = []string{"rules", "commands", "agents"}

// ListAssets returns the relative paths of all files in the given category
// (e.g. "rules", "commands", "agents"). Excludes .gitkeep files.
func ListAssets(category string) ([]string, error) {
	var files []string
	err := fs.WalkDir(embedded, category, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Base(path) == ".gitkeep" {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(files) == 0 && !isValidCategory(category) {
		return nil, fs.ErrNotExist
	}
	return files, nil
}

// ReadAsset reads the contents of an embedded asset by its relative path
// (e.g. "rules/example.md").
func ReadAsset(path string) ([]byte, error) {
	return embedded.ReadFile(path)
}

// isValidCategory checks if a category name is one of the known asset dirs.
func isValidCategory(category string) bool {
	for _, c := range categories {
		if strings.EqualFold(c, category) {
			return true
		}
	}
	return false
}
