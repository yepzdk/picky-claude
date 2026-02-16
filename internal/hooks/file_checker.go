package hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/jesperpedersen/picky-claude/internal/hooks/checkers"
)

func init() {
	Register("file-checker", fileCheckerHook)
}

func fileCheckerHook(input *Input) error {
	filePath := extractFilePath(input)
	if filePath == "" {
		ExitOK()
		return nil
	}

	ext := filepath.Ext(filePath)
	checker := checkers.ForExtension(ext)
	if checker == nil {
		ExitOK()
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := checker.Check(ctx, filePath)
	if err != nil {
		// Non-fatal: report as warning, don't block
		WriteOutput(&Output{
			SystemMessage: fmt.Sprintf("[%s] checker error: %v", checker.Name(), err),
		})
		return nil
	}

	if len(result.Errors) == 0 && len(result.Warnings) == 0 {
		if result.Fixed {
			WriteOutput(&Output{
				SuppressOutput: true,
			})
		} else {
			ExitOK()
		}
		return nil
	}

	// Build feedback message for Claude
	var msg strings.Builder
	for _, d := range result.Errors {
		fmt.Fprintf(&msg, "[%s] ERROR: %s\n", d.Source, d.Message)
	}
	for _, d := range result.Warnings {
		fmt.Fprintf(&msg, "[%s] WARNING: %s\n", d.Source, d.Message)
	}

	// PostToolUse: use decision=block to send errors back to Claude
	if len(result.Errors) > 0 {
		WriteOutput(&Output{
			Decision: "block",
			Reason:   msg.String(),
		})
	} else {
		WriteOutput(&Output{
			SystemMessage: msg.String(),
		})
	}
	return nil
}

// extractFilePath gets the file path from the tool input, handling both
// Write and Edit tool types.
func extractFilePath(input *Input) string {
	if input.ToolInput == nil {
		return ""
	}

	var ti struct {
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal(input.ToolInput, &ti); err != nil {
		return ""
	}
	return ti.FilePath
}
