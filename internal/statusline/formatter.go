// Package statusline reads Claude Code status JSON from stdin and formats a
// status bar string with widgets for session, context, plan, worktree, and tips.
package statusline

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Input is the JSON structure that Claude Code pipes to the statusline command.
type Input struct {
	SessionID  string  `json:"session_id"`
	ContextPct float64 `json:"context_pct"`
	Plan       *Plan   `json:"plan"`
	Worktree   *Wt     `json:"worktree"`
	Duration   int     `json:"duration_secs"`
	Messages   int     `json:"messages"`
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

// Format reads the Input and produces a formatted status bar string.
func Format(input *Input) string {
	var parts []string

	parts = append(parts, formatSession(input))

	if ctx := formatContext(input.ContextPct); ctx != "" {
		parts = append(parts, ctx)
	}

	if input.Plan != nil {
		parts = append(parts, formatPlan(input.Plan))
	}

	if input.Worktree != nil && input.Worktree.Active {
		parts = append(parts, formatWorktree(input.Worktree))
	}

	if tip := selectTip(input); tip != "" {
		parts = append(parts, tip)
	}

	return strings.Join(parts, " | ")
}

// ParseAndFormat parses JSON bytes and formats the status bar.
func ParseAndFormat(data []byte) (string, error) {
	var input Input
	if err := json.Unmarshal(data, &input); err != nil {
		return "", fmt.Errorf("parse status input: %w", err)
	}
	return Format(&input), nil
}

func formatSession(input *Input) string {
	id := input.SessionID
	if len(id) > 16 {
		id = id[:16]
	}
	dur := formatDuration(input.Duration)
	return fmt.Sprintf("S:%s %s M:%d", id, dur, input.Messages)
}

func formatDuration(secs int) string {
	if secs < 60 {
		return fmt.Sprintf("%ds", secs)
	}
	m := secs / 60
	s := secs % 60
	if m < 60 {
		return fmt.Sprintf("%dm%02ds", m, s)
	}
	h := m / 60
	m = m % 60
	return fmt.Sprintf("%dh%02dm", h, m)
}

func formatContext(pct float64) string {
	if pct <= 0 {
		return ""
	}
	indicator := contextIndicator(pct)
	return fmt.Sprintf("CTX:%s%.0f%%", indicator, pct)
}

func contextIndicator(pct float64) string {
	switch {
	case pct >= 90:
		return "!! "
	case pct >= 80:
		return "! "
	case pct >= 60:
		return "~ "
	default:
		return ""
	}
}

func formatPlan(p *Plan) string {
	name := p.Name
	if len(name) > 20 {
		name = name[:20]
	}
	return fmt.Sprintf("P:%s [%s %d/%d]", name, p.Status, p.Done, p.Total)
}

func formatWorktree(wt *Wt) string {
	return fmt.Sprintf("WT:%s", wt.Branch)
}
