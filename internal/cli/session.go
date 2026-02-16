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

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Session management commands",
}

var sessionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		portStr := os.Getenv(config.EnvPrefix + "_PORT")
		port := config.DefaultPort
		if portStr != "" {
			fmt.Sscanf(portStr, "%d", &port)
		}

		client := session.DefaultConsoleClient(port)
		resp, err := client.Get("/api/sessions")
		if err != nil {
			return fmt.Errorf("list sessions: %w", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode >= 400 {
			return fmt.Errorf("list sessions failed (HTTP %d): %s", resp.StatusCode, body)
		}

		if jsonOutput {
			cmd.OutOrStdout().Write(body)
			fmt.Fprintln(cmd.OutOrStdout())
			return nil
		}

		var sessions []map[string]any
		if err := json.Unmarshal(body, &sessions); err != nil {
			fmt.Fprintln(cmd.OutOrStdout(), "No active sessions")
			return nil
		}

		if len(sessions) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No active sessions")
			return nil
		}

		for _, s := range sessions {
			id, _ := s["ID"].(string)
			project, _ := s["Project"].(string)
			started, _ := s["StartedAt"].(string)
			msgs := s["MessageCount"]
			fmt.Fprintf(cmd.OutOrStdout(), "  %s  project=%s  started=%s  messages=%v\n",
				id, project, started, msgs)
		}
		return nil
	},
}

func init() {
	sessionCmd.AddCommand(sessionListCmd)
	rootCmd.AddCommand(sessionCmd)
}
