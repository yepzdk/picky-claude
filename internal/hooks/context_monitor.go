package hooks

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/jesperpedersen/picky-claude/internal/session"
)

func init() {
	Register("context-monitor", contextMonitorHook)
}

// contextMonitorHook parses context usage from Claude Code's system reminders
// in the transcript, persists the percentage, and emits escalating warnings at
// configured thresholds. Tracks which thresholds have been shown per session.
func contextMonitorHook(input *Input) error {
	sessionDir := resolveSessionDir()

	// Parse context usage from Claude Code's system reminders in the transcript
	pct := parseContextFromTranscript(input.TranscriptPath)
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

// contextPctRe matches "Context at NN%." in system-reminder tags.
var contextPctRe = regexp.MustCompile(`Context at (\d+)%`)

// parseContextFromTranscript scans the transcript JSONL for the last
// system-reminder containing "Context at NN%" and returns that value.
// Falls back to 0 if no context reminder is found.
func parseContextFromTranscript(transcriptPath string) float64 {
	if transcriptPath == "" {
		return 0
	}
	f, err := os.Open(transcriptPath)
	if err != nil {
		return 0
	}
	defer f.Close()

	var lastPct float64
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		// Quick check before full parse
		if !contextPctRe.Match(line) {
			continue
		}
		// Extract the percentage from the matching line
		if m := contextPctRe.FindSubmatch(line); m != nil {
			if v, err := strconv.ParseFloat(string(m[1]), 64); err == nil {
				lastPct = v
			}
		}
	}
	return lastPct
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
