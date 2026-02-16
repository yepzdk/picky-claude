package session

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

// FindClaudeCode locates the Claude Code CLI binary.
// Looks for "claude" in PATH.
func FindClaudeCode() (string, error) {
	path, err := exec.LookPath("claude")
	if err != nil {
		return "", fmt.Errorf("claude not found in PATH: %w", err)
	}
	return path, nil
}

// BuildClaudeArgs returns the default arguments for launching Claude Code.
// Auto-update is disabled via the DISABLE_AUTOUPDATER env var in settings.json,
// not via CLI flags.
func BuildClaudeArgs() []string {
	return nil
}

// WritePIDFile writes the current process PID to a file in the session directory.
func WritePIDFile(sessionDir string) error {
	path := filepath.Join(sessionDir, "pid")
	return os.WriteFile(path, []byte(strconv.Itoa(os.Getpid())), 0o644)
}

// ReadPIDFile reads the PID from the session directory.
func ReadPIDFile(sessionDir string) (int, error) {
	path := filepath.Join(sessionDir, "pid")
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(data))
}

// RemovePIDFile removes the PID file from the session directory.
func RemovePIDFile(sessionDir string) {
	os.Remove(filepath.Join(sessionDir, "pid"))
}
