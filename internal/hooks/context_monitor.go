package hooks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jesperpedersen/picky-claude/internal/config"
)

func init() {
	Register("context-monitor", contextMonitorHook)
}

// contextMonitorHook reads the context usage percentage from a cache file and
// emits escalating warnings at configured thresholds. Tracks which thresholds
// have already been shown per session to avoid duplicate warnings.
func contextMonitorHook(input *Input) error {
	sessionID := input.SessionID
	if sessionID == "" {
		sessionID = "default"
	}

	pct := readContextPercentage(sessionID)
	if pct < 0 {
		ExitOK()
		return nil
	}

	shown := loadShownThresholds(sessionID)
	threshold := currentThreshold(pct)
	if threshold == 0 || shown[threshold] {
		ExitOK()
		return nil
	}

	// Mark as shown
	shown[threshold] = true
	saveShownThresholds(sessionID, shown)

	msg := thresholdMessage(threshold, pct, sessionID)

	if threshold >= 90 {
		// At 90%+ the monitor returns a blocking message via stderr (exit 2)
		// which tells Claude to stop and hand off
		BlockWithError(msg)
		return nil // unreachable after os.Exit(2)
	}

	WriteOutput(&Output{
		SystemMessage: msg,
	})
	return nil
}

func currentThreshold(pct float64) int {
	thresholds := []int{95, 90, 80, 60, 40}
	for _, t := range thresholds {
		if pct >= float64(t) {
			return t
		}
	}
	return 0
}

func thresholdMessage(threshold int, pct float64, sessionID string) string {
	sessionDir := config.SessionDir(sessionID)
	contFile := filepath.Join(sessionDir, "continuation.md")

	switch {
	case threshold >= 95:
		return fmt.Sprintf(
			"CRITICAL: Context at %.0f%%. IMMEDIATE handoff required.\n"+
				"Step 1: Write continuation summary to %s\n"+
				"Step 2: Execute: picky send-clear\n"+
				"Do both in THIS turn. Do NOT start new work.",
			pct, contFile)
	case threshold >= 90:
		return fmt.Sprintf(
			"Context at %.0f%%. Mandatory handoff.\n"+
				"Step 1: Finish current tool call only\n"+
				"Step 2: Write continuation summary to %s\n"+
				"Step 3: Execute: picky send-clear\n"+
				"Do NOT start new fix cycles.",
			pct, contFile)
	case threshold >= 80:
		return fmt.Sprintf(
			"Context at %.0f%%. Prepare for handoff. "+
				"Wrap up current task, avoid starting new complex work.",
			pct)
	case threshold >= 60:
		return fmt.Sprintf("Context at %.0f%%. Monitor your progress.", pct)
	default:
		return fmt.Sprintf("Context at %.0f%%.", pct)
	}
}

// contextPctFile returns the path to the cached context percentage JSON.
type contextPctData struct {
	Percentage float64 `json:"percentage"`
}

func readContextPercentage(sessionID string) float64 {
	path := filepath.Join(config.SessionDir(sessionID), "context-pct.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return -1
	}
	var d contextPctData
	if err := json.Unmarshal(data, &d); err != nil {
		return -1
	}
	return d.Percentage
}

func shownThresholdsFile(sessionID string) string {
	return filepath.Join(config.SessionDir(sessionID), "context-thresholds.json")
}

func loadShownThresholds(sessionID string) map[int]bool {
	data, err := os.ReadFile(shownThresholdsFile(sessionID))
	if err != nil {
		return map[int]bool{}
	}
	var m map[int]bool
	json.Unmarshal(data, &m)
	if m == nil {
		return map[int]bool{}
	}
	return m
}

func saveShownThresholds(sessionID string, m map[int]bool) {
	dir := config.SessionDir(sessionID)
	os.MkdirAll(dir, 0o755)
	data, _ := json.Marshal(m)
	os.WriteFile(shownThresholdsFile(sessionID), data, 0o644)
}
