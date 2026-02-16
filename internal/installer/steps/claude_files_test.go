package steps

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jesperpedersen/picky-claude/internal/installer"
)

func TestClaudeFiles_Name(t *testing.T) {
	step := &ClaudeFiles{}
	if step.Name() != "claude-files" {
		t.Errorf("expected name 'claude-files', got %q", step.Name())
	}
}

func TestClaudeFiles_Run(t *testing.T) {
	dir := t.TempDir()
	ctx := &installer.Context{
		ProjectDir: dir,
		ClaudeDir:  filepath.Join(dir, ".claude"),
	}

	step := &ClaudeFiles{}
	if err := step.Run(ctx); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Verify .claude/rules/ was created with at least one file
	rulesDir := filepath.Join(dir, ".claude", "rules")
	entries, err := os.ReadDir(rulesDir)
	if err != nil {
		t.Fatalf("failed to read rules dir: %v", err)
	}
	if len(entries) == 0 {
		t.Error("expected at least one rule file extracted")
	}
}

func TestClaudeFiles_Rollback(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	ctx := &installer.Context{
		ProjectDir: dir,
		ClaudeDir:  claudeDir,
	}

	step := &ClaudeFiles{}
	step.Run(ctx)

	// Verify .claude exists before rollback
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		t.Fatal(".claude should exist before rollback")
	}

	step.Rollback(ctx)

	// After rollback, .claude should be removed
	if _, err := os.Stat(claudeDir); !os.IsNotExist(err) {
		t.Error(".claude should be removed after rollback")
	}
}
