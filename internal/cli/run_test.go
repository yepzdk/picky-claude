package cli

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestUpdateSettingsAnnouncement(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))

	t.Run("updates port in companyAnnouncements", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "settings.json")

		initial := map[string]any{
			"companyAnnouncements": []string{
				"Console: http://localhost:41777 | /spec — plan, build & verify",
			},
			"alwaysThinkingEnabled": true,
		}
		data, _ := json.MarshalIndent(initial, "", "  ")
		os.WriteFile(path, data, 0o644)

		updateSettingsAnnouncement(path, 42000, logger)

		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read updated file: %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(got, &result); err != nil {
			t.Fatalf("invalid JSON after update: %v", err)
		}

		announcements, ok := result["companyAnnouncements"].([]any)
		if !ok || len(announcements) != 1 {
			t.Fatalf("expected 1 announcement, got %v", result["companyAnnouncements"])
		}

		want := "Console: http://localhost:42000 | /spec — plan, build & verify"
		if announcements[0] != want {
			t.Errorf("announcement = %q, want %q", announcements[0], want)
		}

		// Verify other fields are preserved
		if result["alwaysThinkingEnabled"] != true {
			t.Error("expected alwaysThinkingEnabled to be preserved")
		}
	})

	t.Run("skips gracefully when file missing", func(t *testing.T) {
		updateSettingsAnnouncement("/nonexistent/path/settings.json", 42000, logger)
		// Should not panic
	})

	t.Run("skips gracefully on invalid JSON", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "settings.json")
		os.WriteFile(path, []byte("not json"), 0o644)

		updateSettingsAnnouncement(path, 42000, logger)
		// Should not panic, file should remain unchanged
	})
}

func TestUpdateMCPPort(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))

	t.Run("updates mem-search URL port", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, ".mcp.json")

		initial := map[string]any{
			"mcpServers": map[string]any{
				"mem-search": map[string]any{
					"type": "http",
					"url":  "http://localhost:41777/mcp",
				},
				"context7": map[string]any{
					"command": "npx",
					"args":    []string{"-y", "@upstash/context7-mcp"},
				},
			},
		}
		data, _ := json.MarshalIndent(initial, "", "  ")
		os.WriteFile(path, data, 0o644)

		updateMCPPort(path, 42000, logger)

		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read updated file: %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(got, &result); err != nil {
			t.Fatalf("invalid JSON after update: %v", err)
		}

		servers := result["mcpServers"].(map[string]any)
		memSearch := servers["mem-search"].(map[string]any)

		want := "http://localhost:42000/mcp"
		if memSearch["url"] != want {
			t.Errorf("mem-search url = %q, want %q", memSearch["url"], want)
		}

		// Verify other servers are preserved
		if servers["context7"] == nil {
			t.Error("expected context7 to be preserved")
		}
	})

	t.Run("skips gracefully when file missing", func(t *testing.T) {
		updateMCPPort("/nonexistent/path/.mcp.json", 42000, logger)
	})

	t.Run("skips gracefully when no mem-search server", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, ".mcp.json")

		initial := map[string]any{
			"mcpServers": map[string]any{
				"context7": map[string]any{
					"command": "npx",
				},
			},
		}
		data, _ := json.MarshalIndent(initial, "", "  ")
		os.WriteFile(path, data, 0o644)

		updateMCPPort(path, 42000, logger)
		// Should not panic, file should remain unchanged
	})
}
