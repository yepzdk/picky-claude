// Package statusline formats a minimal status bar with git branch,
// plan progress, and context usage with ANSI coloring.
package statusline

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// ANSI escape codes.
const (
	dim    = "\033[2m"
	yellow = "\033[33m"
	red    = "\033[31m"
	reset  = "\033[0m"
)

// Input is the data used to render the status bar.
type Input struct {
	SessionID  string  `json:"session_id"`
	ContextPct float64 `json:"context_pct"`
	Plan       *Plan   `json:"plan"`
	Tasks      *Tasks  `json:"tasks"`
	Worktree   *Wt     `json:"worktree"`
	Duration   int     `json:"duration_secs"`
	Messages   int     `json:"messages"`
	Branch     string  `json:"branch"`
}

// Tasks represents task/todo completion counts.
type Tasks struct {
	Completed int `json:"completed"`
	Total     int `json:"total"`
}

// Plan represents plan metadata in the status input.
type Plan struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Done   int    `json:"done"`
	Total  int    `json:"total"`
}

// Wt represents worktree metadata in the status input.
type Wt struct {
	Active bool   `json:"active"`
	Branch string `json:"branch"`
	Slug   string `json:"slug"`
}

// Format renders the status bar string with ANSI colors.
// Layout: branch │ P:name done/total │ CTX ▰▰▱▱ pct%
// Empty parts are omitted. Color only for context >= 80%.
func Format(input *Input) string {
	var parts []string

	if input.Branch != "" {
		parts = append(parts, input.Branch)
	}

	if input.Plan != nil {
		parts = append(parts, formatPlan(input.Plan))
	}

	if input.Tasks != nil && input.Tasks.Total > 0 {
		parts = append(parts, formatTasks(input.Tasks))
	}

	if ctx := formatContext(input.ContextPct); ctx != "" {
		parts = append(parts, ctx)
	}

	if len(parts) == 0 {
		return ""
	}

	return dim + strings.Join(parts, " │ ") + reset
}

// ParseAndFormat parses JSON bytes and formats the status bar.
func ParseAndFormat(data []byte) (string, error) {
	var input Input
	if err := json.Unmarshal(data, &input); err != nil {
		return "", fmt.Errorf("parse status input: %w", err)
	}
	return Format(&input), nil
}

func formatContext(pct float64) string {
	if pct <= 0 {
		return ""
	}

	bar := progressBar(pct)
	text := fmt.Sprintf("CTX %s %.0f%%", bar, pct)

	switch {
	case pct >= 90:
		return reset + red + text + " HANDOFF" + reset + dim
	case pct >= 80:
		return reset + yellow + text + reset + dim
	default:
		return text
	}
}

func formatTasks(t *Tasks) string {
	return fmt.Sprintf("T:%d/%d", t.Completed, t.Total)
}

func formatPlan(p *Plan) string {
	name := p.Name
	if len(name) > 20 {
		name = name[:20]
	}
	return fmt.Sprintf("P:%s %d/%d", name, p.Done, p.Total)
}

// progressBar returns a 10-segment bar: ▰ for filled, ▱ for empty.
func progressBar(pct float64) string {
	filled := int(math.Round(pct / 10))
	if filled < 0 {
		filled = 0
	}
	if filled > 10 {
		filled = 10
	}
	return strings.Repeat("▰", filled) + strings.Repeat("▱", 10-filled)
}
