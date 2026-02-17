package steps

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/jesperpedersen/picky-claude/internal/config"
	"github.com/jesperpedersen/picky-claude/internal/installer"
)

// ConfigFiles writes settings.json and .mcp.json to the project's .claude/
// directory. Existing files are not overwritten.
type ConfigFiles struct {
	createdFiles []string // Track created files for rollback
}

func (c *ConfigFiles) Name() string { return "config-files" }

func (c *ConfigFiles) Run(ctx *installer.Context) error {
	if err := os.MkdirAll(ctx.ClaudeDir, 0o755); err != nil {
		return fmt.Errorf("create .claude: %w", err)
	}

	binPath := resolveBinaryPath()

	configs := []struct {
		name    string
		content []byte
	}{
		{"settings.json", settingsJSON(binPath)},
		{".mcp.json", mcpJSON(config.DefaultPort)},
	}

	for _, cfg := range configs {
		path := filepath.Join(ctx.ClaudeDir, cfg.name)
		if _, err := os.Stat(path); err == nil {
			ctx.Messages = append(ctx.Messages, fmt.Sprintf("  ✓ %s already exists", cfg.name))
			continue
		}

		if err := os.WriteFile(path, cfg.content, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", cfg.name, err)
		}
		c.createdFiles = append(c.createdFiles, path)
		ctx.Messages = append(ctx.Messages, fmt.Sprintf("  + Created %s", cfg.name))
	}

	return nil
}

func (c *ConfigFiles) Rollback(ctx *installer.Context) {
	for _, path := range c.createdFiles {
		os.Remove(path) //nolint:errcheck
	}
}

// resolveBinaryPath returns the full path to the current picky binary.
// Falls back to just the binary name if the path can't be determined.
func resolveBinaryPath() string {
	exe, err := os.Executable()
	if err != nil {
		return config.BinaryName
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return exe
	}
	return resolved
}

// settingsJSON returns the project-level Claude Code settings with hooks,
// env vars, permissions, and statusline config.
func settingsJSON(binPath string) []byte {
	settings := map[string]any{
		"env": map[string]string{
			"CLAUDE_CODE_ENABLE_TASKS":                 "true",
			"CLAUDE_CODE_DISABLE_AUTO_MEMORY":          "true",
			"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "true",
			"DISABLE_AUTOUPDATER":                      "true",
			"DISABLE_AUTO_COMPACT":                     "true",
			"DISABLE_COMPACT":                          "true",
			"DISABLE_MICROCOMPACT":                     "false",
			"DISABLE_AUTO_COMPACT_THRESHOLD":           "100",
			"DISABLE_NON_ESSENTIAL_MODEL_CALLS":        "true",
			"DISABLE_INSTALLATION_CHECKS":              "true",
			"DISABLE_ERROR_REPORTING":                  "true",
			"DISABLE_TELEMETRY":                        "true",
			"ENABLE_TOOL_SEARCH":                       "true",
			"ENABLE_LSP_TOOL":                          "true",
		},
		"permissions": map[string]any{
			"allow": []string{
				"Bash",
				"Bash(" + config.BinaryName + " *)",
				"Bash(git status*)",
				"Bash(git diff*)",
				"Bash(git log*)",
				"Bash(git branch*)",
				"Bash(git show*)",
				"Edit",
				"Glob",
				"Grep",
				"NotebookEdit",
				"Read",
				"TodoWrite",
				"Write",
				"Skill(spec)",
				"Skill(spec-plan)",
				"Skill(spec-implement)",
				"Skill(spec-verify)",
				"LSP",
			},
			"deny": []string{},
		},
		"hooks": hooksConfig(binPath),
		"statusLine": map[string]any{
			"type":    "command",
			"command": binPath + " statusline",
			"padding": 0,
		},
		"enableAllProjectMcpServers": true,
		"respectGitignore":           false,
		"alwaysThinkingEnabled":      true,
		"spinnerTipsEnabled":         false,
		"prefersReducedMotion":       true,
		"showTurnDuration":           false,
		"cleanupPeriodDays":          7,
		"companyAnnouncements": []string{
			"Console: http://localhost:" + strconv.Itoa(config.DefaultPort) + " | /spec — plan, build & verify",
		},
	}
	data, _ := json.MarshalIndent(settings, "", "  ")
	return data
}

// hooksConfig returns the hooks section for settings.json using the current
// Claude Code hooks format: each entry has a matcher and a "hooks" array of
// handler objects with type, command, and optional async/timeout fields.
func hooksConfig(binPath string) map[string]any {
	return map[string]any{
		"PreToolUse": []map[string]any{
			{
				"matcher": "Bash|WebSearch|WebFetch|Grep|Task|EnterPlanMode|ExitPlanMode",
				"hooks": []map[string]any{
					{
						"type":    "command",
						"command": binPath + " hook tool-redirect",
						"timeout": 15,
					},
				},
			},
			{
				"matcher": "Bash",
				"hooks": []map[string]any{
					{
						"type":    "command",
						"command": binPath + " hook branch-guard",
						"timeout": 15,
					},
				},
			},
		},
		"PostToolUse": []map[string]any{
			{
				"matcher": "Write|Edit|MultiEdit",
				"hooks": []map[string]any{
					{
						"type":    "command",
						"command": binPath + " hook file-checker",
						"timeout": 15,
					},
				},
			},
			{
				"matcher": "Write|Edit|MultiEdit",
				"hooks": []map[string]any{
					{
						"type":    "command",
						"command": binPath + " hook tdd-enforcer",
						"async":   true,
						"timeout": 15,
					},
				},
			},
			{
				"matcher": "Read|Write|Edit|MultiEdit|Bash|Task|Skill|Grep|Glob",
				"hooks": []map[string]any{
					{
						"type":    "command",
						"command": binPath + " hook context-monitor",
						"async":   true,
						"timeout": 15,
					},
				},
			},
			{
				"matcher": "Write|Edit",
				"hooks": []map[string]any{
					{
						"type":    "command",
						"command": binPath + " hook spec-plan-validator",
						"async":   true,
						"timeout": 15,
					},
				},
			},
			{
				"matcher": "Write|Edit",
				"hooks": []map[string]any{
					{
						"type":    "command",
						"command": binPath + " hook spec-verify-validator",
						"async":   true,
						"timeout": 15,
					},
				},
			},
			{
				"matcher": "TaskCreate|TaskUpdate|TodoWrite",
				"hooks": []map[string]any{
					{
						"type":    "command",
						"command": binPath + " hook task-tracker",
						"async":   true,
						"timeout": 15,
					},
				},
			},
		},
		"Stop": []map[string]any{
			{
				"hooks": []map[string]any{
					{
						"type":    "command",
						"command": binPath + " hook spec-stop-guard",
						"timeout": 15,
					},
				},
			},
		},
		"SessionStart": []map[string]any{
			{
				"hooks": []map[string]any{
					{
						"type":    "command",
						"command": binPath + " hook session-start",
						"timeout": 15,
					},
				},
			},
			{
				"hooks": []map[string]any{
					{
						"type":    "command",
						"command": binPath + " hook branch-guard",
						"timeout": 15,
					},
				},
			},
		},
		"SessionEnd": []map[string]any{
			{
				"hooks": []map[string]any{
					{
						"type":    "command",
						"command": binPath + " hook session-end",
						"timeout": 15,
					},
				},
			},
		},
	}
}

// mcpJSON returns the .mcp.json configuration for MCP servers.
func mcpJSON(port int) []byte {
	mcpConfig := map[string]any{
		"mcpServers": map[string]any{
			"context7": map[string]any{
				"command": "npx",
				"args":    []string{"-y", "@upstash/context7-mcp"},
			},
			"mem-search": map[string]any{
				"type": "http",
				"url":  "http://localhost:" + strconv.Itoa(port) + "/mcp",
			},
			"web-search": map[string]any{
				"command": "npx",
				"args":    []string{"-y", "open-websearch@latest"},
				"env": map[string]string{
					"MODE":                   "stdio",
					"DEFAULT_SEARCH_ENGINE":  "duckduckgo",
					"ALLOWED_SEARCH_ENGINES": "duckduckgo,bing",
				},
			},
			"grep-mcp": map[string]any{
				"type": "http",
				"url":  "https://mcp.grep.app",
			},
			"web-fetch": map[string]any{
				"command": "npx",
				"args":    []string{"-y", "fetcher-mcp"},
			},
		},
	}
	data, _ := json.MarshalIndent(mcpConfig, "", "  ")
	return data
}
