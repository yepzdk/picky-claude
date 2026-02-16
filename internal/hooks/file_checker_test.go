package hooks

import (
	"encoding/json"
	"testing"
)

func TestExtractFilePath_Write(t *testing.T) {
	input := &Input{
		ToolName:  "Write",
		ToolInput: json.RawMessage(`{"file_path": "/tmp/test.py", "content": "print('hello')"}`),
	}
	got := extractFilePath(input)
	if got != "/tmp/test.py" {
		t.Errorf("extractFilePath() = %q, want /tmp/test.py", got)
	}
}

func TestExtractFilePath_Edit(t *testing.T) {
	input := &Input{
		ToolName:  "Edit",
		ToolInput: json.RawMessage(`{"file_path": "/tmp/test.ts", "old_string": "a", "new_string": "b"}`),
	}
	got := extractFilePath(input)
	if got != "/tmp/test.ts" {
		t.Errorf("extractFilePath() = %q, want /tmp/test.ts", got)
	}
}

func TestExtractFilePath_NoInput(t *testing.T) {
	input := &Input{ToolName: "Write"}
	got := extractFilePath(input)
	if got != "" {
		t.Errorf("extractFilePath() = %q, want empty", got)
	}
}

func TestExtractFilePath_InvalidJSON(t *testing.T) {
	input := &Input{
		ToolName:  "Write",
		ToolInput: json.RawMessage(`not json`),
	}
	got := extractFilePath(input)
	if got != "" {
		t.Errorf("extractFilePath() = %q, want empty", got)
	}
}

func TestFileCheckerRegistered(t *testing.T) {
	_, ok := registry["file-checker"]
	if !ok {
		t.Error("file-checker not registered")
	}
}
