package steps

import (
	"runtime"
	"testing"
)

func TestCheckCommand_Git(t *testing.T) {
	// git should be available in CI and dev environments
	err := checkCommand("git", "--version")
	if err != nil {
		t.Skipf("git not available: %v", err)
	}
}

func TestCheckCommand_NonExistent(t *testing.T) {
	err := checkCommand("nonexistent-binary-12345", "--version")
	if err == nil {
		t.Error("expected error for non-existent command")
	}
}

func TestCheckOS_Current(t *testing.T) {
	err := checkOS()
	switch runtime.GOOS {
	case "darwin", "linux":
		if err != nil {
			t.Errorf("expected no error on %s, got: %v", runtime.GOOS, err)
		}
	default:
		// On other platforms, may or may not error — just don't crash
	}
}

func TestPrerequisites_Name(t *testing.T) {
	step := &Prerequisites{}
	if step.Name() != "prerequisites" {
		t.Errorf("expected name 'prerequisites', got %q", step.Name())
	}
}
