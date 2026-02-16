package steps

import (
	"fmt"
	"os"

	"github.com/jesperpedersen/picky-claude/internal/assets"
	"github.com/jesperpedersen/picky-claude/internal/installer"
)

// ClaudeFiles extracts embedded rules, commands, and agents to the project's
// .claude/ directory.
type ClaudeFiles struct {
	created bool // Track if we created .claude/ for rollback
}

func (c *ClaudeFiles) Name() string { return "claude-files" }

func (c *ClaudeFiles) Run(ctx *installer.Context) error {
	// Track if .claude/ existed before we ran
	if _, err := os.Stat(ctx.ClaudeDir); os.IsNotExist(err) {
		c.created = true
	}

	if err := assets.ExtractTo(ctx.ClaudeDir); err != nil {
		return fmt.Errorf("extract assets: %w", err)
	}

	ctx.Messages = append(ctx.Messages, "  + Extracted rules, commands, and agents to .claude/")
	return nil
}

func (c *ClaudeFiles) Rollback(ctx *installer.Context) {
	if c.created {
		os.RemoveAll(ctx.ClaudeDir) //nolint:errcheck
	}
}
