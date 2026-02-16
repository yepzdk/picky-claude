package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/jesperpedersen/picky-claude/internal/statusline"
	"github.com/spf13/cobra"
)

var statuslineCmd = &cobra.Command{
	Use:   "statusline",
	Short: "Format the status bar (reads JSON from stdin)",
	Long: `Reads JSON status data from stdin (piped by Claude Code) and outputs
a formatted status bar string with session, context, plan, and worktree info.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
		if len(data) == 0 {
			return nil
		}

		output, err := statusline.ParseAndFormat(data)
		if err != nil {
			return err
		}

		fmt.Print(output)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statuslineCmd)
}
