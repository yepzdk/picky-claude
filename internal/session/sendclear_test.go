package session

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteClearSignal(t *testing.T) {
	dir := t.TempDir()

	err := WriteClearSignal(dir, "docs/plans/test.md")
	if err != nil {
		t.Fatalf("WriteClearSignal: %v", err)
	}

	// Verify signal file exists
	data, err := os.ReadFile(filepath.Join(dir, "clear-signal.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty signal file")
	}
}

func TestWriteClearSignalGeneral(t *testing.T) {
	dir := t.TempDir()

	err := WriteClearSignal(dir, "")
	if err != nil {
		t.Fatalf("WriteClearSignal: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "clear-signal.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty signal file")
	}
}

func TestReadClearSignal(t *testing.T) {
	dir := t.TempDir()

	WriteClearSignal(dir, "docs/plans/test.md")

	signal, err := ReadClearSignal(dir)
	if err != nil {
		t.Fatalf("ReadClearSignal: %v", err)
	}
	if signal.PlanPath != "docs/plans/test.md" {
		t.Errorf("PlanPath = %q, want docs/plans/test.md", signal.PlanPath)
	}
}

func TestReadClearSignalMissing(t *testing.T) {
	dir := t.TempDir()

	_, err := ReadClearSignal(dir)
	if err == nil {
		t.Error("expected error for missing signal file")
	}
}

func TestRemoveClearSignal(t *testing.T) {
	dir := t.TempDir()

	WriteClearSignal(dir, "")
	RemoveClearSignal(dir)

	_, err := ReadClearSignal(dir)
	if err == nil {
		t.Error("expected error after removing signal file")
	}
}

func TestBuildContinuationPrompt(t *testing.T) {
	prompt := BuildContinuationPrompt("docs/plans/test.md")
	if prompt == "" {
		t.Fatal("expected non-empty prompt")
	}

	generalPrompt := BuildContinuationPrompt("")
	if generalPrompt == "" {
		t.Fatal("expected non-empty prompt for general continuation")
	}

	if prompt == generalPrompt {
		t.Error("plan-based and general prompts should differ")
	}
}
