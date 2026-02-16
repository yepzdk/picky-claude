package cli

import (
	"github.com/jesperpedersen/picky-claude/internal/hooks"
	"github.com/spf13/cobra"

	// Register all hook implementations.
	_ "github.com/jesperpedersen/picky-claude/internal/hooks/checkers"
)

var hookCmd = &cobra.Command{
	Use:   "hook <name>",
	Short: "Run a Claude Code hook by name",
	Long: `Executes a specific hook. Called by Claude Code's hooks.json, not typically
invoked directly. Available hooks: file-checker, tdd-enforcer, context-monitor,
tool-redirect, spec-stop-guard, session-start, session-end, notify.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return hooks.Dispatch(args[0])
	},
}

func init() {
	rootCmd.AddCommand(hookCmd)
}
