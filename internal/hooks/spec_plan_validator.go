package hooks

import (
	"encoding/json"
	"fmt"
	"strings"
)

func init() {
	Register("spec-plan-validator", specPlanValidatorHook)
}

// specPlanValidatorHook validates plan file structure when a plan file is
// written or edited. Only activates for files matching docs/plans/*.md.
func specPlanValidatorHook(input *Input) error {
	msg := specPlanValidatorCheck(input)
	if msg == nil {
		ExitOK()
		return nil
	}

	WriteOutput(&Output{
		SystemMessage: *msg,
	})
	return nil
}

// specPlanValidatorCheck performs the validation and returns a warning message
// if the plan is invalid, or nil if everything is fine.
func specPlanValidatorCheck(input *Input) *string {
	if input.ToolInput == nil {
		return nil
	}

	var ti struct {
		FilePath string `json:"file_path"`
		Content  string `json:"content"`
	}
	if err := json.Unmarshal(input.ToolInput, &ti); err != nil {
		return nil
	}

	// Only validate plan files
	if !isPlanFile(ti.FilePath) {
		return nil
	}

	// For Write, content is in tool input. For Edit, we'd need to read the file.
	// Only validate when we have content (Write tool).
	if ti.Content == "" {
		return nil
	}

	errs := validatePlanContent(ti.Content)
	if len(errs) == 0 {
		return nil
	}

	msg := fmt.Sprintf("Plan validation warnings for %s:\n- %s",
		ti.FilePath, strings.Join(errs, "\n- "))
	return &msg
}

// isPlanFile checks if a file path matches the plan file pattern.
func isPlanFile(path string) bool {
	return strings.Contains(path, "docs/plans/") && strings.HasSuffix(path, ".md")
}

// validatePlanContent checks a plan file's content for required structure.
func validatePlanContent(content string) []string {
	var errs []string

	if !hasField(content, "Status") {
		errs = append(errs, "missing required field: Status")
	} else {
		status := fieldValue(content, "Status")
		validStatuses := map[string]bool{
			"PENDING": true, "COMPLETE": true, "VERIFIED": true,
		}
		if !validStatuses[strings.ToUpper(status)] {
			errs = append(errs, fmt.Sprintf("invalid Status value %q (must be PENDING, COMPLETE, or VERIFIED)", status))
		}
	}

	if !hasField(content, "Worktree") {
		errs = append(errs, "missing required field: Worktree")
	}

	if !hasTasksSection(content) {
		errs = append(errs, "missing ## Tasks section")
	}

	return errs
}

// hasField checks if a "Field: value" line exists in the content.
func hasField(content, field string) bool {
	prefix := field + ":"
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), prefix) {
			return true
		}
	}
	return false
}

// fieldValue extracts the value from a "Field: value" line.
func fieldValue(content, field string) string {
	prefix := field + ":"
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
		}
	}
	return ""
}

// hasTasksSection checks if the content contains a ## Tasks heading.
func hasTasksSection(content string) bool {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "## Tasks" || strings.HasPrefix(trimmed, "## Tasks") {
			return true
		}
	}
	return false
}
