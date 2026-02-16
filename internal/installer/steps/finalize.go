package steps

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jesperpedersen/picky-claude/internal/config"
	"github.com/jesperpedersen/picky-claude/internal/installer"
)

// Finalize verifies the installation and produces a summary.
type Finalize struct{}

func (f *Finalize) Name() string { return "finalize" }

func (f *Finalize) Run(ctx *installer.Context) error {
	// Verify .claude/ exists with expected structure
	checks := []struct {
		path string
		desc string
	}{
		{ctx.ClaudeDir, ".claude/ directory"},
		{filepath.Join(ctx.ClaudeDir, "rules"), ".claude/rules/"},
		{filepath.Join(ctx.ClaudeDir, "commands"), ".claude/commands/"},
		{filepath.Join(ctx.ClaudeDir, "agents"), ".claude/agents/"},
	}

	for _, c := range checks {
		if _, err := os.Stat(c.path); os.IsNotExist(err) {
			return fmt.Errorf("verification failed: %s not found at %s", c.desc, c.path)
		}
	}

	// Count installed assets
	ruleCount := countFiles(filepath.Join(ctx.ClaudeDir, "rules"))
	cmdCount := countFiles(filepath.Join(ctx.ClaudeDir, "commands"))
	agentCount := countFiles(filepath.Join(ctx.ClaudeDir, "agents"))

	ctx.Messages = append(ctx.Messages,
		fmt.Sprintf("  ✓ Verified: %d rules, %d commands, %d agents", ruleCount, cmdCount, agentCount),
		fmt.Sprintf("  ✓ %s installation complete", config.DisplayName),
	)

	return nil
}

func (f *Finalize) Rollback(ctx *installer.Context) {
	// Nothing to undo — this step only verifies.
}

// countFiles returns the number of non-hidden files in a directory.
func countFiles(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() && e.Name()[0] != '.' {
			count++
		}
	}
	return count
}
