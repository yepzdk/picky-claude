package steps

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jesperpedersen/picky-claude/internal/installer"
)

func TestConfigFiles_Name(t *testing.T) {
	step := &ConfigFiles{}
	if step.Name() != "config-files" {
		t.Errorf("expected name 'config-files', got %q", step.Name())
	}
}

func TestConfigFiles_Run(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	os.MkdirAll(claudeDir, 0o755)

	ctx := &installer.Context{
		ProjectDir: dir,
		ClaudeDir:  claudeDir,
	}

	step := &ConfigFiles{}
	if err := step.Run(ctx); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Verify settings.json exists and is valid JSON with hooks
	settingsPath := filepath.Join(claudeDir, "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("failed to read settings.json: %v", err)
	}
	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Errorf("settings.json is not valid JSON: %v", err)
	}

	// Verify hooks are embedded in settings.json
	if _, ok := settings["hooks"]; !ok {
		t.Error("settings.json should contain hooks section")
	}

	// Verify env section exists
	if _, ok := settings["env"]; !ok {
		t.Error("settings.json should contain env section")
	}

	// Verify .mcp.json exists and is valid JSON
	mcpPath := filepath.Join(claudeDir, ".mcp.json")
	data, err = os.ReadFile(mcpPath)
	if err != nil {
		t.Fatalf("failed to read .mcp.json: %v", err)
	}
	var mcpConfig map[string]any
	if err := json.Unmarshal(data, &mcpConfig); err != nil {
		t.Errorf(".mcp.json is not valid JSON: %v", err)
	}

	// Verify MCP servers are configured
	servers, ok := mcpConfig["mcpServers"].(map[string]any)
	if !ok {
		t.Fatal(".mcp.json should contain mcpServers")
	}
	for _, name := range []string{"context7", "mem-search", "web-search", "grep-mcp", "web-fetch"} {
		if _, ok := servers[name]; !ok {
			t.Errorf(".mcp.json missing server: %s", name)
		}
	}
}

func TestConfigFiles_DoesNotOverwrite(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	os.MkdirAll(claudeDir, 0o755)

	// Pre-create settings.json with custom content
	settingsPath := filepath.Join(claudeDir, "settings.json")
	custom := `{"custom": true}`
	os.WriteFile(settingsPath, []byte(custom), 0o644)

	ctx := &installer.Context{
		ProjectDir: dir,
		ClaudeDir:  claudeDir,
	}

	step := &ConfigFiles{}
	step.Run(ctx)

	// settings.json should NOT be overwritten
	data, _ := os.ReadFile(settingsPath)
	if string(data) != custom {
		t.Error("existing settings.json should not be overwritten")
	}
}

func TestHooksConfigIncludesSessionStart(t *testing.T) {
	hooks := hooksConfig("/test/bin/picky")

	sessionStart, ok := hooks["SessionStart"]
	if !ok {
		t.Fatal("hooksConfig should include SessionStart")
	}

	// Verify SessionStart structure matches SessionEnd pattern
	startHooks, ok := sessionStart.([]map[string]any)
	if !ok || len(startHooks) != 1 {
		t.Fatal("SessionStart should have one hook entry")
	}

	hookList, ok := startHooks[0]["hooks"].([]map[string]any)
	if !ok || len(hookList) != 1 {
		t.Fatal("SessionStart should have one hook command")
	}

	hook := hookList[0]
	if hook["type"] != "command" {
		t.Errorf("expected type=command, got %v", hook["type"])
	}
	if hook["command"] != "/test/bin/picky hook session-start" {
		t.Errorf("expected command with session-start, got %v", hook["command"])
	}
	if hook["timeout"] != 15 {
		t.Errorf("expected timeout=15, got %v", hook["timeout"])
	}
}

func TestConfigFiles_Rollback(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	os.MkdirAll(claudeDir, 0o755)

	ctx := &installer.Context{
		ProjectDir: dir,
		ClaudeDir:  claudeDir,
	}

	step := &ConfigFiles{}
	step.Run(ctx)

	step.Rollback(ctx)

	// Newly created files should be removed
	for _, name := range []string{"settings.json", ".mcp.json"} {
		path := filepath.Join(claudeDir, name)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("%s should be removed after rollback", name)
		}
	}
}
