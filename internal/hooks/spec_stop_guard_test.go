package hooks

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jesperpedersen/picky-claude/internal/config"
)

func TestSpecStopGuard_NoPlanFile(t *testing.T) {
	// With no plan file, the guard should allow stopping
	input := &Input{
		HookEventName: "Stop",
		SessionID:     "test-stop-noplan",
		Cwd:           t.TempDir(),
	}

	result := specStopGuardCheck(input)
	if result != nil {
		t.Errorf("expected nil (allow stop) with no plan, got: %v", result)
	}
}

func TestSpecStopGuard_PlanVerified(t *testing.T) {
	dir := t.TempDir()
	planDir := filepath.Join(dir, "docs", "plans")
	os.MkdirAll(planDir, 0o755)

	planContent := "# Plan\n\nStatus: VERIFIED\n\n## Tasks\n- [x] Done\n"
	os.WriteFile(filepath.Join(planDir, "2026-01-01-test.md"), []byte(planContent), 0o644)

	input := &Input{
		HookEventName: "Stop",
		SessionID:     "test-stop-verified",
		Cwd:           dir,
	}

	result := specStopGuardCheck(input)
	if result != nil {
		t.Errorf("expected nil (allow stop) with VERIFIED plan, got: %v", result)
	}
}

func TestSpecStopGuard_PlanPending(t *testing.T) {
	dir := t.TempDir()
	planDir := filepath.Join(dir, "docs", "plans")
	os.MkdirAll(planDir, 0o755)

	planContent := "# Plan\n\nStatus: PENDING\n\n## Tasks\n- [ ] Not done\n"
	os.WriteFile(filepath.Join(planDir, "2026-01-01-test.md"), []byte(planContent), 0o644)

	input := &Input{
		HookEventName: "Stop",
		SessionID:     "test-stop-pending",
		Cwd:           dir,
	}

	result := specStopGuardCheck(input)
	if result == nil {
		t.Error("expected block message for PENDING plan, got nil")
	}
}

func TestSpecStopGuard_PlanComplete(t *testing.T) {
	dir := t.TempDir()
	planDir := filepath.Join(dir, "docs", "plans")
	os.MkdirAll(planDir, 0o755)

	planContent := "# Plan\n\nStatus: COMPLETE\n\n## Tasks\n- [x] Done\n"
	os.WriteFile(filepath.Join(planDir, "2026-01-01-test.md"), []byte(planContent), 0o644)

	input := &Input{
		HookEventName: "Stop",
		SessionID:     "test-stop-complete",
		Cwd:           dir,
	}

	result := specStopGuardCheck(input)
	if result == nil {
		t.Error("expected block message for COMPLETE plan (needs verification), got nil")
	}
}

func TestSpecStopGuard_HighContext(t *testing.T) {
	dir := t.TempDir()
	planDir := filepath.Join(dir, "docs", "plans")
	os.MkdirAll(planDir, 0o755)

	planContent := "# Plan\n\nStatus: PENDING\n\n## Tasks\n- [ ] Not done\n"
	os.WriteFile(filepath.Join(planDir, "2026-01-01-test.md"), []byte(planContent), 0o644)

	// Write high context percentage
	sessionID := "test-stop-highctx"
	sessionDir := config.SessionDir(sessionID)
	os.MkdirAll(sessionDir, 0o755)
	os.WriteFile(filepath.Join(sessionDir, "context-pct.json"), []byte(`{"percentage": 92.0}`), 0o644)

	input := &Input{
		HookEventName: "Stop",
		SessionID:     sessionID,
		Cwd:           dir,
	}

	result := specStopGuardCheck(input)
	if result != nil {
		t.Error("expected nil (allow stop) at high context even with PENDING plan")
	}
}
