package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewIDFromPID(t *testing.T) {
	id := NewID()
	if id == "" {
		t.Fatal("expected non-empty session ID")
	}
	// Should contain the PID prefix
	if !strings.HasPrefix(id, "picky-") {
		t.Errorf("ID %q does not start with 'picky-'", id)
	}
}

func TestNewIDUnique(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := NewID()
		if ids[id] {
			t.Fatalf("duplicate ID: %s", id)
		}
		ids[id] = true
	}
}

func TestEnsureSessionDir(t *testing.T) {
	tmpDir := t.TempDir()
	dir := filepath.Join(tmpDir, "sessions", "test-session")

	err := EnsureSessionDir(dir)
	if err != nil {
		t.Fatalf("EnsureSessionDir: %v", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory")
	}

	// Calling again should be idempotent
	err = EnsureSessionDir(dir)
	if err != nil {
		t.Fatalf("second EnsureSessionDir: %v", err)
	}
}

func TestWriteAndReadContextPercentage(t *testing.T) {
	dir := t.TempDir()

	if err := WriteContextPercentage(dir, 47.5); err != nil {
		t.Fatalf("WriteContextPercentage: %v", err)
	}

	pct, err := ReadContextPercentage(dir)
	if err != nil {
		t.Fatalf("ReadContextPercentage: %v", err)
	}
	if pct != 47.5 {
		t.Errorf("percentage = %f, want 47.5", pct)
	}
}

func TestReadContextPercentageMissing(t *testing.T) {
	dir := t.TempDir()

	_, err := ReadContextPercentage(dir)
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestBuildEnv(t *testing.T) {
	env := BuildEnv("test-session-123", 41777)

	found := map[string]string{}
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			found[parts[0]] = parts[1]
		}
	}

	if v := found["PICKY_SESSION_ID"]; v != "test-session-123" {
		t.Errorf("PICKY_SESSION_ID = %q, want test-session-123", v)
	}
	if v := found["CLAUDE_CODE_TASK_LIST_ID"]; v != "picky-test-session-123" {
		t.Errorf("CLAUDE_CODE_TASK_LIST_ID = %q, want picky-test-session-123", v)
	}
	if v := found["PICKY_PORT"]; v != "41777" {
		t.Errorf("PICKY_PORT = %q, want 41777", v)
	}
}
