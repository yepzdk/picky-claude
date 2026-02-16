package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jesperpedersen/picky-claude/internal/config"
)

// ClearSignal is written to the session directory to trigger an Endless Mode
// restart. The `picky run` command watches for this file.
type ClearSignal struct {
	PlanPath string `json:"plan_path"`
	General  bool   `json:"general"`
}

const clearSignalFile = "clear-signal.json"

// WriteClearSignal writes a clear signal to the session directory.
// If planPath is empty, it's a general (no-plan) continuation.
func WriteClearSignal(sessionDir string, planPath string) error {
	signal := ClearSignal{
		PlanPath: planPath,
		General:  planPath == "",
	}
	data, err := json.Marshal(signal)
	if err != nil {
		return fmt.Errorf("marshal clear signal: %w", err)
	}
	path := filepath.Join(sessionDir, clearSignalFile)
	return os.WriteFile(path, data, 0o644)
}

// ReadClearSignal reads the clear signal from the session directory.
func ReadClearSignal(sessionDir string) (*ClearSignal, error) {
	path := filepath.Join(sessionDir, clearSignalFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read clear signal: %w", err)
	}
	var signal ClearSignal
	if err := json.Unmarshal(data, &signal); err != nil {
		return nil, fmt.Errorf("parse clear signal: %w", err)
	}
	return &signal, nil
}

// RemoveClearSignal deletes the clear signal file.
func RemoveClearSignal(sessionDir string) {
	os.Remove(filepath.Join(sessionDir, clearSignalFile))
}

// BuildContinuationPrompt generates the prompt sent after /clear to resume
// the session. If planPath is set, the prompt references the plan.
func BuildContinuationPrompt(planPath string) string {
	if planPath != "" {
		return fmt.Sprintf(
			"Continue the session. Read the plan at `%s` and the continuation file "+
				"at `~/%s/sessions/$%s_SESSION_ID/continuation.md` to understand "+
				"where the previous session left off. Resume from the next incomplete task.",
			planPath, config.ConfigDirName, config.EnvPrefix,
		)
	}
	return fmt.Sprintf(
		"Continue the session. Read the continuation file at "+
			"`~/%s/sessions/$%s_SESSION_ID/continuation.md` to understand "+
			"where the previous session left off. Resume from the next step.",
		config.ConfigDirName, config.EnvPrefix,
	)
}
