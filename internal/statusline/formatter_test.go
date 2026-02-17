package statusline

import (
	"strings"
	"testing"
)

func TestProgressBar(t *testing.T) {
	tests := []struct {
		pct  float64
		want string
	}{
		{0, "▱▱▱▱▱▱▱▱▱▱"},
		{5, "▰▱▱▱▱▱▱▱▱▱"},
		{25, "▰▰▰▱▱▱▱▱▱▱"},
		{50, "▰▰▰▰▰▱▱▱▱▱"},
		{75, "▰▰▰▰▰▰▰▰▱▱"},
		{100, "▰▰▰▰▰▰▰▰▰▰"},
		{-5, "▱▱▱▱▱▱▱▱▱▱"},
		{150, "▰▰▰▰▰▰▰▰▰▰"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := progressBar(tt.pct)
			if got != tt.want {
				t.Errorf("progressBar(%.0f) = %q, want %q", tt.pct, got, tt.want)
			}
		})
	}
}

func TestFormatContext_Normal(t *testing.T) {
	got := formatContext(42)
	// Should contain the bar and percentage, no color escapes
	if !strings.Contains(got, "42%") {
		t.Errorf("formatContext(42) should contain 42%%, got %q", got)
	}
	if strings.Contains(got, "HANDOFF") {
		t.Error("formatContext(42) should not contain HANDOFF")
	}
	// Should use dim color only
	if strings.Contains(got, yellow) || strings.Contains(got, red) {
		t.Error("formatContext(42) should not use yellow or red")
	}
}

func TestFormatContext_Warning(t *testing.T) {
	got := formatContext(83)
	if !strings.Contains(got, yellow) {
		t.Error("formatContext(83) should use yellow")
	}
	if !strings.Contains(got, "83%") {
		t.Errorf("formatContext(83) should contain 83%%, got %q", got)
	}
	if strings.Contains(got, "HANDOFF") {
		t.Error("formatContext(83) should not contain HANDOFF")
	}
}

func TestFormatContext_Critical(t *testing.T) {
	got := formatContext(92)
	if !strings.Contains(got, red) {
		t.Error("formatContext(92) should use red")
	}
	if !strings.Contains(got, "HANDOFF") {
		t.Error("formatContext(92) should contain HANDOFF")
	}
}

func TestFormatContext_Zero(t *testing.T) {
	got := formatContext(0)
	if got != "" {
		t.Errorf("formatContext(0) should be empty, got %q", got)
	}
}

func TestFormatPlan(t *testing.T) {
	p := &Plan{Name: "add-auth", Status: "PENDING", Done: 2, Total: 5}
	got := formatPlan(p)
	want := "P:add-auth 2/5"
	if got != want {
		t.Errorf("formatPlan() = %q, want %q", got, want)
	}
}

func TestFormatPlanLongName(t *testing.T) {
	p := &Plan{Name: "very-long-plan-name-that-exceeds-limit", Status: "COMPLETE", Done: 5, Total: 5}
	got := formatPlan(p)
	if !strings.HasPrefix(got, "P:very-long-plan-name-") {
		t.Errorf("formatPlan() should truncate name, got %q", got)
	}
	if !strings.Contains(got, "5/5") {
		t.Errorf("formatPlan() should contain 5/5, got %q", got)
	}
}

func TestFormat_BranchOnly(t *testing.T) {
	input := &Input{Branch: "main"}
	got := Format(input)
	if !strings.Contains(got, "main") {
		t.Errorf("Format() should contain branch, got %q", got)
	}
	// Should not contain separator when only one part
	if strings.Contains(got, "│") {
		t.Errorf("Format() should not have separator with single part, got %q", got)
	}
}

func TestFormat_BranchAndContext(t *testing.T) {
	input := &Input{Branch: "main", ContextPct: 42}
	got := Format(input)
	if !strings.Contains(got, "main") {
		t.Errorf("Format() should contain branch, got %q", got)
	}
	if !strings.Contains(got, "42%") {
		t.Errorf("Format() should contain context, got %q", got)
	}
	if !strings.Contains(got, "│") {
		t.Errorf("Format() should have separator, got %q", got)
	}
}

func TestFormat_Full(t *testing.T) {
	input := &Input{
		Branch:     "feat/auth",
		ContextPct: 45,
		Plan:       &Plan{Name: "auth", Status: "PENDING", Done: 1, Total: 4},
	}
	got := Format(input)
	if !strings.Contains(got, "feat/auth") {
		t.Errorf("Format() missing branch, got %q", got)
	}
	if !strings.Contains(got, "P:auth 1/4") {
		t.Errorf("Format() missing plan, got %q", got)
	}
	if !strings.Contains(got, "45%") {
		t.Errorf("Format() missing context, got %q", got)
	}
}

func TestFormatTasks(t *testing.T) {
	got := formatTasks(&Tasks{Completed: 2, Total: 5})
	want := "T:2/5"
	if got != want {
		t.Errorf("formatTasks() = %q, want %q", got, want)
	}
}

func TestFormatTasks_AllDone(t *testing.T) {
	got := formatTasks(&Tasks{Completed: 3, Total: 3})
	want := "T:3/3"
	if got != want {
		t.Errorf("formatTasks() = %q, want %q", got, want)
	}
}

func TestFormat_WithTasks(t *testing.T) {
	input := &Input{
		Branch: "feat/auth",
		Tasks:  &Tasks{Completed: 1, Total: 4},
	}
	got := Format(input)
	if !strings.Contains(got, "T:1/4") {
		t.Errorf("Format() should contain T:1/4, got %q", got)
	}
	if !strings.Contains(got, "feat/auth") {
		t.Errorf("Format() should contain branch, got %q", got)
	}
	if !strings.Contains(got, "│") {
		t.Errorf("Format() should have separator, got %q", got)
	}
}

func TestFormat_TasksZeroTotal(t *testing.T) {
	input := &Input{
		Branch: "main",
		Tasks:  &Tasks{Completed: 0, Total: 0},
	}
	got := Format(input)
	if strings.Contains(got, "T:") {
		t.Errorf("Format() should not show tasks with zero total, got %q", got)
	}
}

func TestFormat_FullWithTasks(t *testing.T) {
	input := &Input{
		Branch:     "feat/auth",
		ContextPct: 45,
		Plan:       &Plan{Name: "auth", Status: "PENDING", Done: 1, Total: 4},
		Tasks:      &Tasks{Completed: 2, Total: 6},
	}
	got := Format(input)
	if !strings.Contains(got, "feat/auth") {
		t.Errorf("Format() missing branch, got %q", got)
	}
	if !strings.Contains(got, "P:auth 1/4") {
		t.Errorf("Format() missing plan, got %q", got)
	}
	if !strings.Contains(got, "T:2/6") {
		t.Errorf("Format() missing tasks, got %q", got)
	}
	if !strings.Contains(got, "45%") {
		t.Errorf("Format() missing context, got %q", got)
	}
}

func TestFormat_Empty(t *testing.T) {
	input := &Input{}
	got := Format(input)
	if got != "" {
		t.Errorf("Format() should be empty for no data, got %q", got)
	}
}

func TestFormat_ContextOnly(t *testing.T) {
	input := &Input{ContextPct: 60}
	got := Format(input)
	if !strings.Contains(got, "60%") {
		t.Errorf("Format() should contain context, got %q", got)
	}
}

func TestFormat_CriticalColoring(t *testing.T) {
	input := &Input{Branch: "main", ContextPct: 95}
	got := Format(input)
	if !strings.Contains(got, red) {
		t.Error("Format() at 95% should contain red ANSI code")
	}
	if !strings.Contains(got, "HANDOFF") {
		t.Error("Format() at 95% should contain HANDOFF")
	}
}

func TestParseAndFormat(t *testing.T) {
	data := []byte(`{"branch":"main","context_pct":50}`)
	got, err := ParseAndFormat(data)
	if err != nil {
		t.Fatalf("ParseAndFormat() error: %v", err)
	}
	if !strings.Contains(got, "main") {
		t.Errorf("ParseAndFormat() should contain branch, got %q", got)
	}
	if !strings.Contains(got, "50%") {
		t.Errorf("ParseAndFormat() should contain context, got %q", got)
	}
}

func TestParseAndFormatInvalidJSON(t *testing.T) {
	_, err := ParseAndFormat([]byte(`{invalid`))
	if err == nil {
		t.Error("ParseAndFormat() expected error for invalid JSON")
	}
}
