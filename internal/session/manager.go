// Package session manages Claude Code session lifecycle: ID generation,
// session directories, context percentage tracking, and environment setup.
package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"

	"github.com/jesperpedersen/picky-claude/internal/config"
)

// counter ensures unique IDs even within the same process.
var counter atomic.Int64

// NewID generates a unique session ID based on the current PID and a counter.
func NewID() string {
	n := counter.Add(1)
	return fmt.Sprintf("picky-%d-%d", os.Getpid(), n)
}

// EnsureSessionDir creates the session directory if it doesn't exist.
func EnsureSessionDir(dir string) error {
	return os.MkdirAll(dir, 0o755)
}

// contextPctData is the JSON structure for context percentage cache.
type contextPctData struct {
	Percentage float64 `json:"percentage"`
}

// WriteContextPercentage writes the current context usage percentage to the
// session directory's context-pct.json file.
func WriteContextPercentage(sessionDir string, pct float64) error {
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		return fmt.Errorf("create session dir: %w", err)
	}
	data, err := json.Marshal(contextPctData{Percentage: pct})
	if err != nil {
		return fmt.Errorf("marshal context pct: %w", err)
	}
	path := filepath.Join(sessionDir, "context-pct.json")
	return os.WriteFile(path, data, 0o644)
}

// ReadContextPercentage reads the context usage percentage from the session
// directory's context-pct.json file.
func ReadContextPercentage(sessionDir string) (float64, error) {
	path := filepath.Join(sessionDir, "context-pct.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("read context pct: %w", err)
	}
	var d contextPctData
	if err := json.Unmarshal(data, &d); err != nil {
		return 0, fmt.Errorf("parse context pct: %w", err)
	}
	return d.Percentage, nil
}

// BuildEnv constructs the environment variables that should be set when
// launching Claude Code. It inherits the current process environment and
// adds/overrides session-specific variables.
func BuildEnv(sessionID string, port int) []string {
	env := os.Environ()
	env = setEnv(env, config.EnvPrefix+"_SESSION_ID", sessionID)
	env = setEnv(env, config.EnvPrefix+"_PORT", strconv.Itoa(port))
	env = setEnv(env, config.EnvPrefix+"_HOME", config.HomeDir())
	env = setEnv(env, "CLAUDE_CODE_TASK_LIST_ID", config.BinaryName+"-"+sessionID)
	return env
}

// setEnv sets or replaces an environment variable in a slice.
func setEnv(env []string, key, value string) []string {
	prefix := key + "="
	for i, e := range env {
		if len(e) >= len(prefix) && e[:len(prefix)] == prefix {
			env[i] = prefix + value
			return env
		}
	}
	return append(env, prefix+value)
}
