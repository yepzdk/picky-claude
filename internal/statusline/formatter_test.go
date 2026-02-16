package statusline

import (
	"testing"
)

func TestFormatSession(t *testing.T) {
	tests := []struct {
		name     string
		input    *Input
		contains string
	}{
		{
			name:     "basic session info",
			input:    &Input{SessionID: "picky-123-1", Duration: 90, Messages: 5},
			contains: "S:picky-123-1 1m30s M:5",
		},
		{
			name:     "long session ID truncated",
			input:    &Input{SessionID: "picky-very-long-session-id-12345", Duration: 0, Messages: 0},
			contains: "S:picky-very-long- 0s M:0",
		},
		{
			name:     "hours formatting",
			input:    &Input{SessionID: "s1", Duration: 3661, Messages: 100},
			contains: "S:s1 1h01m M:100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSession(tt.input)
			if result != tt.contains {
				t.Errorf("formatSession() = %q, want %q", result, tt.contains)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		secs int
		want string
	}{
		{0, "0s"},
		{30, "30s"},
		{60, "1m00s"},
		{90, "1m30s"},
		{3600, "1h00m"},
		{3661, "1h01m"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatDuration(tt.secs)
			if got != tt.want {
				t.Errorf("formatDuration(%d) = %q, want %q", tt.secs, got, tt.want)
			}
		})
	}
}

func TestFormatContext(t *testing.T) {
	tests := []struct {
		pct  float64
		want string
	}{
		{0, ""},
		{30, "CTX:30%"},
		{60, "CTX:~ 60%"},
		{80, "CTX:! 80%"},
		{95, "CTX:!! 95%"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatContext(tt.pct)
			if got != tt.want {
				t.Errorf("formatContext(%v) = %q, want %q", tt.pct, got, tt.want)
			}
		})
	}
}

func TestFormatPlan(t *testing.T) {
	p := &Plan{Name: "add-auth", Status: "PENDING", Done: 2, Total: 5}
	got := formatPlan(p)
	want := "P:add-auth [PENDING 2/5]"
	if got != want {
		t.Errorf("formatPlan() = %q, want %q", got, want)
	}
}

func TestFormatPlanLongName(t *testing.T) {
	p := &Plan{Name: "very-long-plan-name-that-exceeds-limit", Status: "COMPLETE", Done: 5, Total: 5}
	got := formatPlan(p)
	want := "P:very-long-plan-name- [COMPLETE 5/5]"
	if got != want {
		t.Errorf("formatPlan() = %q, want %q", got, want)
	}
}

func TestFormatWorktree(t *testing.T) {
	wt := &Wt{Active: true, Branch: "spec/add-auth"}
	got := formatWorktree(wt)
	want := "WT:spec/add-auth"
	if got != want {
		t.Errorf("formatWorktree() = %q, want %q", got, want)
	}
}

func TestFormat(t *testing.T) {
	input := &Input{
		SessionID:  "picky-42-1",
		ContextPct: 45,
		Duration:   120,
		Messages:   10,
	}
	got := Format(input)
	want := "S:picky-42-1 2m00s M:10 | CTX:45%"
	if got != want {
		t.Errorf("Format() = %q, want %q", got, want)
	}
}

func TestFormatWithPlanAndWorktree(t *testing.T) {
	input := &Input{
		SessionID:  "s1",
		ContextPct: 85,
		Duration:   60,
		Messages:   3,
		Plan:       &Plan{Name: "auth", Status: "PENDING", Done: 1, Total: 4},
		Worktree:   &Wt{Active: true, Branch: "spec/auth"},
	}
	got := Format(input)
	want := "S:s1 1m00s M:3 | CTX:! 85% | P:auth [PENDING 1/4] | WT:spec/auth | TIP: Wrap up current task"
	if got != want {
		t.Errorf("Format() = %q, want %q", got, want)
	}
}

func TestParseAndFormat(t *testing.T) {
	data := []byte(`{"session_id":"test-1","context_pct":50,"duration_secs":30,"messages":2}`)
	got, err := ParseAndFormat(data)
	if err != nil {
		t.Fatalf("ParseAndFormat() error: %v", err)
	}
	want := "S:test-1 30s M:2 | CTX:50%"
	if got != want {
		t.Errorf("ParseAndFormat() = %q, want %q", got, want)
	}
}

func TestParseAndFormatInvalidJSON(t *testing.T) {
	_, err := ParseAndFormat([]byte(`{invalid`))
	if err == nil {
		t.Error("ParseAndFormat() expected error for invalid JSON")
	}
}

func TestSelectTip(t *testing.T) {
	tests := []struct {
		name  string
		input *Input
		want  string
	}{
		{"at 90% context", &Input{ContextPct: 90}, "TIP: Handoff imminent"},
		{"at 80% context", &Input{ContextPct: 80}, "TIP: Wrap up current task"},
		{"verified plan", &Input{Plan: &Plan{Status: "VERIFIED"}, Messages: 5}, "TIP: Plan verified, done!"},
		{"new session", &Input{Messages: 0}, "TIP: Session started"},
		{"normal state", &Input{Messages: 10, ContextPct: 40}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selectTip(tt.input)
			if got != tt.want {
				t.Errorf("selectTip() = %q, want %q", got, tt.want)
			}
		})
	}
}
