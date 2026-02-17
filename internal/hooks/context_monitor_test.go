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

func TestEstimateContextFromTranscript(t *testing.T) {
	// Empty path returns 0
	if got := estimateContextFromTranscript(""); got != 0 {
		t.Errorf("empty path: got %v, want 0", got)
	}

	// Missing file returns 0
	if got := estimateContextFromTranscript("/nonexistent/file"); got != 0 {
		t.Errorf("missing file: got %v, want 0", got)
	}

	// Small file: 5000 bytes → 1000 tokens → 0.5%
	dir := t.TempDir()
	smallFile := filepath.Join(dir, "small.jsonl")
	os.WriteFile(smallFile, make([]byte, 5000), 0o644)
	got := estimateContextFromTranscript(smallFile)
	if got < 0.4 || got > 0.6 {
		t.Errorf("small file: got %.2f%%, want ~0.5%%", got)
	}

	// Large file: 500_000 bytes → 100_000 tokens → 50%
	largeFile := filepath.Join(dir, "large.jsonl")
	os.WriteFile(largeFile, make([]byte, 500_000), 0o644)
	got = estimateContextFromTranscript(largeFile)
	if got < 49 || got > 51 {
		t.Errorf("large file: got %.2f%%, want ~50%%", got)
	}

	// Huge file: capped at 100%
	hugeFile := filepath.Join(dir, "huge.jsonl")
	os.WriteFile(hugeFile, make([]byte, 2_000_000), 0o644)
	got = estimateContextFromTranscript(hugeFile)
	if got != 100 {
		t.Errorf("huge file: got %.2f%%, want 100%%", got)
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
