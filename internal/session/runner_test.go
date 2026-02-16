package session

import (
	"os"
	"testing"
)

func TestFindClaudeCode(t *testing.T) {
	// This test verifies FindClaudeCode returns a path or error.
	// We can't guarantee claude is installed in CI, so just verify the function works.
	path, err := FindClaudeCode()
	if err != nil {
		// Not installed — that's OK in test environments
		t.Skipf("claude not found: %v", err)
	}
	if path == "" {
		t.Error("FindClaudeCode returned empty path with no error")
	}
}

func TestBuildClaudeArgs(t *testing.T) {
	args := BuildClaudeArgs()
	// BuildClaudeArgs returns nil — no default flags.
	// Auto-update is disabled via env vars in settings.json, not CLI flags.
	if args != nil {
		t.Errorf("expected nil args, got %v", args)
	}
}

func TestBuildEnvSetsAllRequired(t *testing.T) {
	env := BuildEnv("test-123", 41777)

	required := map[string]bool{
		"PICKY_SESSION_ID":         false,
		"PICKY_PORT":               false,
		"PICKY_HOME":               false,
		"CLAUDE_CODE_TASK_LIST_ID": false,
	}

	for _, e := range env {
		for key := range required {
			if len(e) > len(key) && e[:len(key)+1] == key+"=" {
				required[key] = true
			}
		}
	}

	for key, found := range required {
		if !found {
			t.Errorf("missing env var: %s", key)
		}
	}
}

func TestPidFilePath(t *testing.T) {
	dir := t.TempDir()

	// Write PID file
	if err := WritePIDFile(dir); err != nil {
		t.Fatalf("WritePIDFile: %v", err)
	}

	// Read it back
	pid, err := ReadPIDFile(dir)
	if err != nil {
		t.Fatalf("ReadPIDFile: %v", err)
	}
	if pid != os.Getpid() {
		t.Errorf("pid = %d, want %d", pid, os.Getpid())
	}

	// Remove it
	RemovePIDFile(dir)
	_, err = ReadPIDFile(dir)
	if err == nil {
		t.Error("expected error after removing PID file")
	}
}
