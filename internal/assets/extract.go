package assets

import (
	"fmt"
	"os"
	"path/filepath"
)

// ExtractTo extracts all embedded rules, commands, and agents to the given
// target directory (typically .claude/). Creates subdirectories as needed.
// Existing files are overwritten.
func ExtractTo(targetDir string) error {
	for _, category := range categories {
		files, err := ListAssets(category)
		if err != nil {
			continue // Category might be empty
		}

		for _, relPath := range files {
			data, err := ReadAsset(relPath)
			if err != nil {
				return fmt.Errorf("read embedded %s: %w", relPath, err)
			}

			destPath := filepath.Join(targetDir, relPath)
			if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
				return fmt.Errorf("create dir for %s: %w", destPath, err)
			}

			if err := os.WriteFile(destPath, data, 0o644); err != nil {
				return fmt.Errorf("write %s: %w", destPath, err)
			}
		}
	}
	return nil
}
