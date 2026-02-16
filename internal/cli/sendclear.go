package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jesperpedersen/picky-claude/internal/config"
	"github.com/jesperpedersen/picky-claude/internal/session"
	"github.com/spf13/cobra"
)

var sendClearGeneral bool

var sendClearCmd = &cobra.Command{
	Use:   "send-clear [plan-path]",
	Short: "Trigger Endless Mode session restart",
	Long: `Triggers a session restart for Endless Mode continuation.
Provide a plan file path to restart with plan context, or use --general
to restart without a plan.

Steps:
1. Waits for memory capture (10s)
2. Writes clear signal to session directory
3. Waits for session end hooks (5s)
4. Outputs continuation prompt`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sessionID := os.Getenv(config.EnvPrefix + "_SESSION_ID")
		if sessionID == "" {
			sessionID = "default"
		}

		var planPath string
		if len(args) == 1 {
			planPath = args[0]
		} else if !sendClearGeneral {
			return fmt.Errorf("provide a plan path or use --general")
		}

		sessionDir := config.SessionDir(sessionID)
		if err := session.EnsureSessionDir(sessionDir); err != nil {
			return fmt.Errorf("create session dir: %w", err)
		}

		// Step 1: Wait for memory capture
		fmt.Fprintln(cmd.ErrOrStderr(), "Waiting for memory capture (10s)...")
		time.Sleep(10 * time.Second)

		// Step 2: Write clear signal
		if err := session.WriteClearSignal(sessionDir, planPath); err != nil {
			return fmt.Errorf("write clear signal: %w", err)
		}

		// Step 3: Wait for session end hooks
		fmt.Fprintln(cmd.ErrOrStderr(), "Waiting for session end hooks (5s)...")
		time.Sleep(5 * time.Second)

		// Step 4: Output continuation prompt
		prompt := session.BuildContinuationPrompt(planPath)

		if jsonOutput {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]string{
				"status":     "clear_sent",
				"session_id": sessionID,
				"plan_path":  planPath,
				"prompt":     prompt,
			})
		}

		fmt.Fprintln(cmd.OutOrStdout(), prompt)
		return nil
	},
}

func init() {
	sendClearCmd.Flags().BoolVar(&sendClearGeneral, "general", false, "restart without plan context")
	rootCmd.AddCommand(sendClearCmd)
}
