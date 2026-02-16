package steps

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jesperpedersen/picky-claude/internal/config"
	"github.com/jesperpedersen/picky-claude/internal/installer"
)

const pickyMarker = "# Added by " + config.DisplayName

// shellBlock returns the lines to add to shell config files.
func shellBlock(binDir string) string {
	return fmt.Sprintf("\n%s\nexport PATH=\"%s:$PATH\"\n", pickyMarker, binDir)
}

// ShellConfig adds the picky binary directory to PATH in shell config files.
type ShellConfig struct {
	shellFiles []string          // Override for testing; nil = auto-detect
	backups    map[string][]byte // Original content for rollback
}

func (s *ShellConfig) Name() string { return "shell-config" }

func (s *ShellConfig) Run(ctx *installer.Context) error {
	s.backups = make(map[string][]byte)

	exe, err := os.Executable()
	if err != nil {
		exe = os.Args[0]
	}
	binDir := filepath.Dir(exe)
	if binDir == "." || binDir == "" {
		binDir = "/usr/local/bin"
	}

	files := s.shellFiles
	if files == nil {
		files = detectShellFiles()
	}

	block := shellBlock(binDir)

	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			continue // File doesn't exist, skip
		}

		content := string(data)
		if strings.Contains(content, pickyMarker) {
			ctx.Messages = append(ctx.Messages, fmt.Sprintf("  ✓ %s already configured", filepath.Base(path)))
			continue
		}

		s.backups[path] = data

		if err := os.WriteFile(path, []byte(content+block), 0o644); err != nil {
			return fmt.Errorf("update %s: %w", path, err)
		}
		ctx.Messages = append(ctx.Messages, fmt.Sprintf("  + Updated %s", filepath.Base(path)))
	}

	return nil
}

func (s *ShellConfig) Rollback(ctx *installer.Context) {
	for path, data := range s.backups {
		os.WriteFile(path, data, 0o644) //nolint:errcheck
	}
}

// detectShellFiles returns the list of shell config files to modify.
func detectShellFiles() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	candidates := []string{
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".bashrc"),
		filepath.Join(home, ".bash_profile"),
	}

	var found []string
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			found = append(found, path)
		}
	}
	return found
}
