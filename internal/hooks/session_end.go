package hooks

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/jesperpedersen/picky-claude/internal/config"
	"github.com/jesperpedersen/picky-claude/internal/session"
)

func init() {
	Register("session-end", sessionEndHook)
}

// sessionEndHook posts a session summary to the console server when Claude Code exits.
func sessionEndHook(input *Input) error {
	// Read PICKY_PORT from env
	portStr := os.Getenv(config.EnvPrefix + "_PORT")
	if portStr == "" {
		// Not running in a managed picky session - exit silently
		ExitOK()
		return nil // unreachable
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		// Invalid port - exit silently
		ExitOK()
		return nil // unreachable
	}

	// Post summary to console
	client := session.DefaultConsoleClient(port)
	_ = postSummary(client, input.SessionID) // Ignore errors - never block shutdown

	// Exit cleanly
	ExitOK()
	return nil // unreachable
}

// postSummary posts a session summary to the console's /api/summaries endpoint.
func postSummary(client *session.ConsoleClient, sessionID string) error {
	payload := map[string]string{
		"session_id": sessionID,
		"text":       "Session " + sessionID + " ended",
	}

	resp, err := client.Post("/api/summaries", payload)
	if err != nil {
		return fmt.Errorf("post summary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}
