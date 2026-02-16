package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/jesperpedersen/picky-claude/internal/config"
)

func init() {
	Register("tdd-enforcer", tddEnforcerHook)
}

// tddEnforcerHook tracks whether test files or production files are being
// edited. If production code is written without a test file in the same
// session, it emits a warning.
func tddEnforcerHook(input *Input) error {
	filePath := extractFilePath(input)
	if filePath == "" {
		ExitOK()
		return nil
	}

	// Detect if this is a test file
	isTest := isTestFile(filePath)

	stateFile := tddStateFile(input.SessionID)
	state := loadTDDState(stateFile)

	if isTest {
		state.TestWritten = true
		saveTDDState(stateFile, state)
		ExitOK()
		return nil
	}

	// Production file was edited
	if !state.TestWritten && !state.Warned {
		state.Warned = true
		saveTDDState(stateFile, state)

		WriteOutput(&Output{
			SystemMessage: "TDD reminder: You wrote production code before writing a test. " +
				"Consider writing a failing test first (RED), then implementing " +
				"the code to make it pass (GREEN).",
		})
		return nil
	}

	saveTDDState(stateFile, state)
	ExitOK()
	return nil
}

// isTestFile checks if a file path looks like a test file.
func isTestFile(path string) bool {
	base := filepath.Base(path)
	lower := strings.ToLower(base)

	// Common test file patterns
	patterns := []string{
		"_test.go",
		"_test.py",
		"test_",
		".test.",
		".spec.",
		"_spec.",
		"__tests__",
	}
	for _, p := range patterns {
		if strings.Contains(lower, p) {
			return true
		}
	}

	// Check if in a test directory
	dir := strings.ToLower(path)
	testDirs := []string{"/tests/", "/test/", "/__tests__/", "/spec/"}
	for _, d := range testDirs {
		if strings.Contains(dir, d) {
			return true
		}
	}

	return false
}

type tddState struct {
	TestWritten bool `json:"test_written"`
	Warned      bool `json:"warned"`
}

func tddStateFile(sessionID string) string {
	if sessionID == "" {
		sessionID = "default"
	}
	return filepath.Join(config.SessionDir(sessionID), "tdd-state.json")
}

func loadTDDState(path string) *tddState {
	data, err := os.ReadFile(path)
	if err != nil {
		return &tddState{}
	}
	var state tddState
	json.Unmarshal(data, &state)
	return &state
}

func saveTDDState(path string, state *tddState) {
	os.MkdirAll(filepath.Dir(path), 0o755)
	data, _ := json.Marshal(state)
	os.WriteFile(path, data, 0o644)
}
