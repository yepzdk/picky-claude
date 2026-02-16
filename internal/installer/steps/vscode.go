package steps

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jesperpedersen/picky-claude/internal/installer"
)

// VSCode creates a .vscode/extensions.json with recommended extensions.
type VSCode struct {
	created bool // Track if we created .vscode/ for rollback
}

func (v *VSCode) Name() string { return "vscode" }

func (v *VSCode) Run(ctx *installer.Context) error {
	vscodeDir := filepath.Join(ctx.ProjectDir, ".vscode")
	extPath := filepath.Join(vscodeDir, "extensions.json")

	if _, err := os.Stat(extPath); err == nil {
		ctx.Messages = append(ctx.Messages, "  ✓ .vscode/extensions.json already exists")
		return nil
	}

	if _, err := os.Stat(vscodeDir); os.IsNotExist(err) {
		v.created = true
	}

	if err := os.MkdirAll(vscodeDir, 0o755); err != nil {
		return fmt.Errorf("create .vscode: %w", err)
	}

	data, _ := json.MarshalIndent(extensionsJSON(), "", "  ")
	if err := os.WriteFile(extPath, data, 0o644); err != nil {
		return fmt.Errorf("write extensions.json: %w", err)
	}

	ctx.Messages = append(ctx.Messages, "  + Created .vscode/extensions.json with recommended extensions")
	return nil
}

func (v *VSCode) Rollback(ctx *installer.Context) {
	if v.created {
		os.RemoveAll(filepath.Join(ctx.ProjectDir, ".vscode")) //nolint:errcheck
	}
}

// extensionsJSON returns the recommended VS Code extensions.
func extensionsJSON() map[string]any {
	return map[string]any{
		"recommendations": []string{
			"anthropics.claude-code",
		},
	}
}
