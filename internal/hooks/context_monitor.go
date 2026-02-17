package hooks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jesperpedersen/picky-claude/internal/session"
)

func init() {
	Register("context-monitor", contextMonitorHook)
}

// contextWindowTokens is the assumed context window size for percentage calculation.
const contextWindowTokens = 200_000

// contextMonitorHook estimates context usage from the transcript file, persists
// the percentage, and emits escalating warnings at configured thresholds.
// Tracks which thresholds have already been shown per session to avoid duplicates.
func contextMonitorHook(input *Input) error {
	sessionDir := resolveSessionDir()

	// Estimate context usage from transcript and persist it
	pct := estimateContextFromTranscript(input.TranscriptPath)
	if pct > 0 {
		session.WriteContextPercentage(sessionDir, pct)
	}

	if pct <= 0 {
		ExitOK()
		return nil
	}

	shown := loadShownThresholds(sessionDir)
	threshold := currentThreshold(pct)
	if threshold == 0 || shown[threshold] {
		ExitOK()
		return nil
	}

	// Mark as shown
	shown[threshold] = true
	saveShownThresholds(sessionDir, shown)

	msg := thresholdMessage(threshold, pct, sessionDir)

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

// estimateContextFromTranscript estimates the context usage percentage by
// reading the transcript file size and using a chars-to-tokens heuristic.
func estimateContextFromTranscript(transcriptPath string) float64 {
	if transcriptPath == "" {
		return 0
	}
	info, err := os.Stat(transcriptPath)
	if err != nil {
		return 0
	}
	// Estimate tokens: ~4 bytes per token for English text.
	// Transcript files include JSON overhead, so we use a conservative
	// factor of 5 bytes per token to avoid overestimating.
	estimatedTokens := float64(info.Size()) / 5.0
	pct := (estimatedTokens / contextWindowTokens) * 100
	if pct > 100 {
		pct = 100
	}
	return pct
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

func thresholdMessage(threshold int, pct float64, sessionDir string) string {
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

// readContextPct reads the persisted context percentage from a session directory.
func readContextPct(sessionDir string) float64 {
	path := filepath.Join(sessionDir, "context-pct.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return -1
	}
	var d struct {
		Percentage float64 `json:"percentage"`
	}
	if err := json.Unmarshal(data, &d); err != nil {
		return -1
	}
	return d.Percentage
}

func shownThresholdsFile(sessionDir string) string {
	return filepath.Join(sessionDir, "context-thresholds.json")
}

func loadShownThresholds(sessionDir string) map[int]bool {
	data, err := os.ReadFile(shownThresholdsFile(sessionDir))
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

func saveShownThresholds(sessionDir string, m map[int]bool) {
	os.MkdirAll(sessionDir, 0o755)
	data, _ := json.Marshal(m)
	os.WriteFile(shownThresholdsFile(sessionDir), data, 0o644)
}
