package hooks

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func init() {
	Register("spec-stop-guard", specStopGuardHook)
}

// specStopGuardHook prevents Claude from stopping during an active /spec
// workflow. If a plan file with status PENDING or COMPLETE exists, the hook
// blocks the stop and tells Claude to continue. At high context (>=90%), the
// stop is allowed so the session can hand off.
func specStopGuardHook(input *Input) error {
	msg := specStopGuardCheck(input)
	if msg == nil {
		ExitOK()
		return nil
	}

	BlockWithError(*msg)
	return nil // unreachable after os.Exit(2)
}

// specStopGuardCheck performs the guard logic and returns a blocking message
// if the stop should be prevented, or nil if the stop is allowed.
// Extracted for testability (avoids os.Exit calls in tests).
func specStopGuardCheck(input *Input) *string {
	// At high context, always allow stop for handoff
	if isHighContext(input.SessionID) {
		return nil
	}

	// Find the most recent plan file
	status := findActivePlanStatus(input.Cwd)
	if status == "" {
		return nil // No plan found, allow stop
	}

	switch strings.ToUpper(status) {
	case "VERIFIED":
		return nil // Plan complete, allow stop
	case "PENDING":
		msg := "Stop blocked: /spec workflow is active (Status: PENDING). " +
			"Continue implementing the plan tasks. " +
			"If context is high, use picky send-clear for handoff."
		return &msg
	case "COMPLETE":
		msg := "Stop blocked: /spec workflow needs verification (Status: COMPLETE). " +
			"Run spec-verify before stopping. " +
			"If context is high, use picky send-clear for handoff."
		return &msg
	default:
		return nil
	}
}

// isHighContext returns true if context usage is at 90% or above.
func isHighContext(sessionID string) bool {
	if sessionID == "" {
		sessionID = "default"
	}
	pct := readContextPercentage(sessionID)
	return pct >= 90
}

// findActivePlanStatus looks for the most recent plan file in docs/plans/
// and returns its Status value. Returns empty string if no plan is found.
func findActivePlanStatus(cwd string) string {
	planDir := filepath.Join(cwd, "docs", "plans")
	entries, err := os.ReadDir(planDir)
	if err != nil {
		return ""
	}

	// Collect .md files, sort descending (most recent first by filename)
	var planFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			planFiles = append(planFiles, e.Name())
		}
	}
	if len(planFiles) == 0 {
		return ""
	}
	sort.Sort(sort.Reverse(sort.StringSlice(planFiles)))

	// Read the most recent plan and extract status
	data, err := os.ReadFile(filepath.Join(planDir, planFiles[0]))
	if err != nil {
		return ""
	}
	return extractPlanStatus(string(data))
}

// extractPlanStatus finds the "Status: <VALUE>" line in a plan file.
func extractPlanStatus(content string) string {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Status:") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "Status:"))
		}
	}
	return ""
}

// readContextPercentage is declared in context_monitor.go and reused here.
// No redeclaration needed — the function is package-level in the same package.
var _ = readContextPercentage // ensure it's available (compile-time check)
