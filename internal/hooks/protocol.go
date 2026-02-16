// Package hooks implements Claude Code hook handlers. Each hook reads JSON
// from stdin (the hook input) and writes JSON to stdout (the hook output).
// Exit code 0 = success, exit code 2 = blocking error (stderr shown to Claude).
package hooks

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Input is the common JSON structure received from Claude Code on stdin.
// Event-specific fields are decoded separately by each hook.
type Input struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	Cwd            string `json:"cwd"`
	PermissionMode string `json:"permission_mode"`
	HookEventName  string `json:"hook_event_name"`

	// Tool events (PreToolUse, PostToolUse)
	ToolName  string          `json:"tool_name,omitempty"`
	ToolInput json.RawMessage `json:"tool_input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`

	// PostToolUse
	ToolResponse json.RawMessage `json:"tool_response,omitempty"`

	// Stop / SubagentStop
	StopHookActive bool `json:"stop_hook_active,omitempty"`

	// SessionStart
	Source string `json:"source,omitempty"`

	// UserPromptSubmit
	Prompt string `json:"prompt,omitempty"`
}

// WriteToolInput contains fields from a Write tool call.
type WriteToolInput struct {
	FilePath string `json:"file_path"`
	Content  string `json:"content"`
}

// EditToolInput contains fields from an Edit tool call.
type EditToolInput struct {
	FilePath   string `json:"file_path"`
	OldString  string `json:"old_string"`
	NewString  string `json:"new_string"`
	ReplaceAll bool   `json:"replace_all"`
}

// BashToolInput contains fields from a Bash tool call.
type BashToolInput struct {
	Command     string `json:"command"`
	Description string `json:"description,omitempty"`
}

// Output is the JSON structure written to stdout on exit 0.
type Output struct {
	Decision       string             `json:"decision,omitempty"`
	Reason         string             `json:"reason,omitempty"`
	Continue       *bool              `json:"continue,omitempty"`
	StopReason     string             `json:"stopReason,omitempty"`
	SuppressOutput bool               `json:"suppressOutput,omitempty"`
	SystemMessage  string             `json:"systemMessage,omitempty"`
	HookSpecific   *HookSpecificOuput `json:"hookSpecificOutput,omitempty"`
}

// HookSpecificOuput holds event-specific output fields.
type HookSpecificOuput struct {
	HookEventName            string `json:"hookEventName"`
	PermissionDecision       string `json:"permissionDecision,omitempty"`
	PermissionDecisionReason string `json:"permissionDecisionReason,omitempty"`
	AdditionalContext        string `json:"additionalContext,omitempty"`
}

// ReadInput reads and parses the hook input JSON from stdin.
func ReadInput() (*Input, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("read stdin: %w", err)
	}
	var input Input
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("parse hook input: %w", err)
	}
	return &input, nil
}

// WriteOutput writes the hook output JSON to stdout and exits with code 0.
func WriteOutput(out *Output) {
	json.NewEncoder(os.Stdout).Encode(out)
}

// BlockWithError writes a message to stderr and exits with code 2.
// This tells Claude Code to block the action and show the error to Claude.
func BlockWithError(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(2)
}

// ExitOK exits with code 0 and no output (allow action to proceed).
func ExitOK() {
	os.Exit(0)
}
