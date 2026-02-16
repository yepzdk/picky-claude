package steps

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jesperpedersen/picky-claude/internal/installer"
)

func TestShellConfig_Name(t *testing.T) {
	step := &ShellConfig{}
	if step.Name() != "shell-config" {
		t.Errorf("expected name 'shell-config', got %q", step.Name())
	}
}

func TestShellConfig_AddsToZshrc(t *testing.T) {
	dir := t.TempDir()
	zshrc := filepath.Join(dir, ".zshrc")
	os.WriteFile(zshrc, []byte("# existing config\n"), 0o644)

	step := &ShellConfig{shellFiles: []string{zshrc}}
	ctx := &installer.Context{ProjectDir: dir}

	if err := step.Run(ctx); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	data, _ := os.ReadFile(zshrc)
	content := string(data)
	if !strings.Contains(content, "# existing config") {
		t.Error("existing content should be preserved")
	}
	if !strings.Contains(content, pickyMarker) {
		t.Error("expected picky marker in shell config")
	}
}

func TestShellConfig_Idempotent(t *testing.T) {
	dir := t.TempDir()
	zshrc := filepath.Join(dir, ".zshrc")
	os.WriteFile(zshrc, []byte("# existing\n"), 0o644)

	step := &ShellConfig{shellFiles: []string{zshrc}}
	ctx := &installer.Context{ProjectDir: dir}

	// Run twice
	step.Run(ctx)
	step.Run(ctx)

	data, _ := os.ReadFile(zshrc)
	// Should only have the marker once
	count := strings.Count(string(data), pickyMarker)
	if count != 1 {
		t.Errorf("expected marker to appear once, appeared %d times", count)
	}
}

func TestShellConfig_Rollback(t *testing.T) {
	dir := t.TempDir()
	zshrc := filepath.Join(dir, ".zshrc")
	original := "# my config\n"
	os.WriteFile(zshrc, []byte(original), 0o644)

	step := &ShellConfig{shellFiles: []string{zshrc}}
	ctx := &installer.Context{ProjectDir: dir}

	step.Run(ctx)
	step.Rollback(ctx)

	data, _ := os.ReadFile(zshrc)
	if string(data) != original {
		t.Errorf("expected rollback to restore original content, got: %q", string(data))
	}
}

func TestShellConfig_SkipsMissingFiles(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, ".nonexistent_rc")

	step := &ShellConfig{shellFiles: []string{missing}}
	ctx := &installer.Context{ProjectDir: dir}

	// Should not error when file doesn't exist
	if err := step.Run(ctx); err != nil {
		t.Fatalf("Run should not fail for missing shell file: %v", err)
	}
}
