package hooks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/jesperpedersen/picky-claude/internal/config"
	"github.com/jesperpedersen/picky-claude/internal/session"
)

func init() {
	Register("session-start", sessionStartHook)
}

// sessionStartHook fetches context from the console server and injects it
// into Claude Code's session via additionalContext.
func sessionStartHook(input *Input) error {
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

	// Fetch context from console
	client := session.DefaultConsoleClient(port)
	context, err := fetchContext(client, input.SessionID)
	if err != nil || context == "" {
		// Console not ready or no context - exit silently (never block session start)
		ExitOK()
		return nil // unreachable
	}

	// Return context via additionalContext
	WriteOutput(&Output{
		HookSpecific: &HookSpecificOuput{
			HookEventName:     "SessionStart",
			AdditionalContext: context,
		},
	})
	return nil
}

// fetchContext fetches context from the console's /api/context/inject endpoint.
// Returns the context string and any error encountered.
func fetchContext(client *session.ConsoleClient, sessionID string) (string, error) {
	path := fmt.Sprintf("/api/context/inject?session_id=%s", url.QueryEscape(sessionID))
	resp, err := client.Get(path)
	if err != nil {
		return "", fmt.Errorf("get context: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result struct {
		Context string `json:"context"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return result.Context, nil
}
