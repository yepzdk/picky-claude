package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/jesperpedersen/picky-claude/internal/config"
	"github.com/jesperpedersen/picky-claude/internal/session"
	"github.com/spf13/cobra"
)

var registerPlanCmd = &cobra.Command{
	Use:   "register-plan <path> <status>",
	Short: "Associate a plan file with the current session",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		planPath := args[0]
		status := args[1]

		sessionID := os.Getenv(config.EnvPrefix + "_SESSION_ID")
		if sessionID == "" {
			sessionID = "default"
		}

		portStr := os.Getenv(config.EnvPrefix + "_PORT")
		port := config.DefaultPort
		if portStr != "" {
			fmt.Sscanf(portStr, "%d", &port)
		}

		client := session.DefaultConsoleClient(port)
		resp, err := client.Post("/api/plans", map[string]string{
			"path":       planPath,
			"session_id": sessionID,
			"status":     status,
		})
		if err != nil {
			return fmt.Errorf("register plan: %w", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode >= 400 {
			return fmt.Errorf("register plan failed (HTTP %d): %s", resp.StatusCode, body)
		}

		if jsonOutput {
			cmd.OutOrStdout().Write(body)
			fmt.Fprintln(cmd.OutOrStdout())
			return nil
		}

		var result map[string]any
		json.Unmarshal(body, &result)
		fmt.Fprintf(cmd.OutOrStdout(), "Plan registered: %s (status: %s)\n", planPath, status)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(registerPlanCmd)
}
