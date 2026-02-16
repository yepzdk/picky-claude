package steps

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jesperpedersen/picky-claude/internal/installer"
)

func TestVSCode_Name(t *testing.T) {
	step := &VSCode{}
	if step.Name() != "vscode" {
		t.Errorf("expected name 'vscode', got %q", step.Name())
	}
}

func TestVSCode_CreatesExtensionsJSON(t *testing.T) {
	dir := t.TempDir()
	vscodeDir := filepath.Join(dir, ".vscode")

	ctx := &installer.Context{ProjectDir: dir}
	step := &VSCode{}

	if err := step.Run(ctx); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	extPath := filepath.Join(vscodeDir, "extensions.json")
	data, err := os.ReadFile(extPath)
	if err != nil {
		t.Fatalf("failed to read extensions.json: %v", err)
	}

	var ext map[string]any
	if err := json.Unmarshal(data, &ext); err != nil {
		t.Fatalf("extensions.json is not valid JSON: %v", err)
	}

	recs, ok := ext["recommendations"].([]any)
	if !ok {
		t.Fatal("expected recommendations array")
	}
	if len(recs) == 0 {
		t.Error("expected at least one recommendation")
	}
}

func TestVSCode_DoesNotOverwrite(t *testing.T) {
	dir := t.TempDir()
	vscodeDir := filepath.Join(dir, ".vscode")
	os.MkdirAll(vscodeDir, 0o755)

	extPath := filepath.Join(vscodeDir, "extensions.json")
	custom := `{"recommendations":["custom.ext"]}`
	os.WriteFile(extPath, []byte(custom), 0o644)

	ctx := &installer.Context{ProjectDir: dir}
	step := &VSCode{}
	step.Run(ctx)

	data, _ := os.ReadFile(extPath)
	if string(data) != custom {
		t.Error("existing extensions.json should not be overwritten")
	}
}

func TestVSCode_Rollback(t *testing.T) {
	dir := t.TempDir()
	ctx := &installer.Context{ProjectDir: dir}

	step := &VSCode{}
	step.Run(ctx)

	step.Rollback(ctx)

	vscodeDir := filepath.Join(dir, ".vscode")
	if _, err := os.Stat(vscodeDir); !os.IsNotExist(err) {
		t.Error(".vscode should be removed after rollback")
	}
}
