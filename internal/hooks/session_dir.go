package hooks

import (
	"os"

	"github.com/jesperpedersen/picky-claude/internal/config"
)

// resolveSessionDir returns the session directory to use for local state files.
// Prefers the PICKY_SESSION_ID env var (set by picky run), falls back to "default".
// This ensures hooks and the statusline command use the same directory.
func resolveSessionDir() string {
	sessionID := os.Getenv(config.EnvPrefix + "_SESSION_ID")
	if sessionID == "" {
		sessionID = "default"
	}
	return config.SessionDir(sessionID)
}
