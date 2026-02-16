package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jesperpedersen/picky-claude/internal/config"
	"github.com/jesperpedersen/picky-claude/internal/session"
	"github.com/spf13/cobra"
)

var checkContextCmd = &cobra.Command{
	Use:   "check-context",
	Short: "Get current context usage percentage",
	RunE: func(cmd *cobra.Command, args []string) error {
		sessionID := os.Getenv(config.EnvPrefix + "_SESSION_ID")
		if sessionID == "" {
			sessionID = "default"
		}

		sessionDir := config.SessionDir(sessionID)
		result, err := session.CheckContextFromDir(sessionDir)
		if err != nil {
			// No data available — return OK with 0%
			result = session.ContextResult{Status: "OK", Percentage: 0}
		}

		if jsonOutput {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(result)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Context: %.1f%% (%s)\n", result.Percentage, result.Status)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkContextCmd)
}
