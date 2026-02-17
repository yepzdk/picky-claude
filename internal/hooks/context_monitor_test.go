package hooks

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jesperpedersen/picky-claude/internal/config"
)

func TestCurrentThreshold(t *testing.T) {
	tests := []struct {
		pct  float64
		want int
	}{
		{10, 0},
		{39, 0},
		{40, 40},
		{55, 40},
		{60, 60},
		{79, 60},
		{80, 80},
		{89, 80},
		{90, 90},
		{94, 90},
		{95, 95},
		{100, 95},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := currentThreshold(tt.pct)
			if got != tt.want {
				t.Errorf("currentThreshold(%.0f) = %d, want %d", tt.pct, got, tt.want)
			}
		})
	}
}

func TestThresholdMessage(t *testing.T) {
	sessionDir := t.TempDir()

	msg := thresholdMessage(95, 96.5, sessionDir)
	if msg == "" {
		t.Error("expected non-empty message for threshold 95")
	}

	msg = thresholdMessage(80, 82.0, sessionDir)
	if msg == "" {
		t.Error("expected non-empty message for threshold 80")
	}
}

func TestContextMonitorRegistered(t *testing.T) {
	_, ok := registry["context-monitor"]
	if !ok {
		t.Error("context-monitor not registered")
	}
}

func TestParseContextFromTranscript(t *testing.T) {
	// Empty path returns 0
	if got := parseContextFromTranscript(""); got != 0 {
		t.Errorf("empty path: got %v, want 0", got)
	}

	// Missing file returns 0
	if got := parseContextFromTranscript("/nonexistent/file"); got != 0 {
		t.Errorf("missing file: got %v, want 0", got)
	}

	dir := t.TempDir()

	// Transcript with no context reminders returns 0
	noCtx := filepath.Join(dir, "no-ctx.jsonl")
	os.WriteFile(noCtx, []byte(`{"type":"user","message":{"role":"user","content":"hello"}}`+"\n"), 0o644)
	if got := parseContextFromTranscript(noCtx); got != 0 {
		t.Errorf("no context reminders: got %v, want 0", got)
	}

	// Transcript with context reminder returns parsed value
	withCtx := filepath.Join(dir, "with-ctx.jsonl")
	os.WriteFile(withCtx, []byte(
		`{"type":"user","message":{"role":"user","content":"hello"}}`+"\n"+
			`{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"Context at 42%."}]}}`+"\n",
	), 0o644)
	if got := parseContextFromTranscript(withCtx); got != 42 {
		t.Errorf("single reminder: got %v, want 42", got)
	}

	// Uses the last context reminder when multiple exist
	multiCtx := filepath.Join(dir, "multi-ctx.jsonl")
	os.WriteFile(multiCtx, []byte(
		`{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"Context at 20%."}]}}`+"\n"+
			`{"type":"user","message":{"role":"user","content":"do more"}}`+"\n"+
			`{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"Context at 61%."}]}}`+"\n",
	), 0o644)
	if got := parseContextFromTranscript(multiCtx); got != 61 {
		t.Errorf("multiple reminders: got %v, want 61", got)
	}
}

func TestResolveSessionDir(t *testing.T) {
	tmpDir := t.TempDir()
	orig := os.Getenv(config.EnvPrefix + "_HOME")
	os.Setenv(config.EnvPrefix+"_HOME", tmpDir)
	defer func() {
		if orig == "" {
			os.Unsetenv(config.EnvPrefix + "_HOME")
		} else {
			os.Setenv(config.EnvPrefix+"_HOME", orig)
		}
	}()

	// Without PICKY_SESSION_ID set, should use "default"
	origSID := os.Getenv(config.EnvPrefix + "_SESSION_ID")
	os.Unsetenv(config.EnvPrefix + "_SESSION_ID")
	defer func() {
		if origSID != "" {
			os.Setenv(config.EnvPrefix+"_SESSION_ID", origSID)
		}
	}()

	got := resolveSessionDir()
	want := config.SessionDir("default")
	if got != want {
		t.Errorf("resolveSessionDir() without env = %q, want %q", got, want)
	}

	// With PICKY_SESSION_ID set, should use that
	os.Setenv(config.EnvPrefix+"_SESSION_ID", "my-session-123")
	got = resolveSessionDir()
	want = config.SessionDir("my-session-123")
	if got != want {
		t.Errorf("resolveSessionDir() with env = %q, want %q", got, want)
	}
}
