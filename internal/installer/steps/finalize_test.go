package steps

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jesperpedersen/picky-claude/internal/installer"
)

func TestFinalize_Name(t *testing.T) {
	step := &Finalize{}
	if step.Name() != "finalize" {
		t.Errorf("expected name 'finalize', got %q", step.Name())
	}
}

func TestFinalize_Run_WithClaudeDir(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	for _, sub := range []string{"rules", "commands", "agents"} {
		os.MkdirAll(filepath.Join(claudeDir, sub), 0o755)
	}
	os.WriteFile(filepath.Join(claudeDir, "rules", "test.md"), []byte("# test"), 0o644)

	ctx := &installer.Context{
		ProjectDir: dir,
		ClaudeDir:  claudeDir,
	}

	step := &Finalize{}
	if err := step.Run(ctx); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Should have added verification messages
	if len(ctx.Messages) == 0 {
		t.Error("expected at least one message from finalize")
	}
}

func TestFinalize_Run_MissingClaudeDir(t *testing.T) {
	dir := t.TempDir()
	ctx := &installer.Context{
		ProjectDir: dir,
		ClaudeDir:  filepath.Join(dir, ".claude"),
	}

	step := &Finalize{}
	err := step.Run(ctx)
	if err == nil {
		t.Error("expected error when .claude/ doesn't exist")
	}
}
