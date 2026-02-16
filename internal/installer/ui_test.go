package installer

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestUI_Banner(t *testing.T) {
	var buf bytes.Buffer
	ui := NewUI(&buf)
	ui.Banner()

	out := buf.String()
	if !strings.Contains(out, "Installer") {
		t.Errorf("expected 'Installer' in banner, got: %s", out)
	}
}

func TestRunWithUI_Success(t *testing.T) {
	var buf bytes.Buffer
	s1 := &mockStep{name: "step-1"}
	s2 := &mockStep{name: "step-2"}

	inst := New(t.TempDir(), s1, s2)
	result := inst.RunWithUI(&buf)

	if !result.Success {
		t.Errorf("expected success, got failure: %v", result.Error)
	}
	out := buf.String()
	if !strings.Contains(out, "step-1") {
		t.Error("expected step-1 in output")
	}
	if !strings.Contains(out, "installed successfully") {
		t.Error("expected success message")
	}
}

func TestRunWithUI_Failure(t *testing.T) {
	var buf bytes.Buffer
	s1 := &mockStep{name: "step-1"}
	s2 := &mockStep{name: "step-2", runErr: errors.New("oops")}

	inst := New(t.TempDir(), s1, s2)
	result := inst.RunWithUI(&buf)

	if result.Success {
		t.Error("expected failure")
	}
	out := buf.String()
	if !strings.Contains(out, "failed") {
		t.Error("expected failure message in output")
	}
	if !strings.Contains(out, "Rolling back") {
		t.Error("expected rollback message")
	}
}
