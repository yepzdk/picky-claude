package hooks

import (
	"encoding/json"
	"fmt"
	"strings"
)

func init() {
	Register("spec-verify-validator", specVerifyValidatorHook)
}

// specVerifyValidatorHook validates verification result JSON files when they
// are written. Only activates for files matching verify-*.json patterns in
// session directories.
func specVerifyValidatorHook(input *Input) error {
	msg := specVerifyValidatorCheck(input)
	if msg == nil {
		ExitOK()
		return nil
	}

	WriteOutput(&Output{
		SystemMessage: *msg,
	})
	return nil
}

// specVerifyValidatorCheck performs the validation and returns a warning
// message if the result is invalid, or nil if valid.
func specVerifyValidatorCheck(input *Input) *string {
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

	if !isVerifyResultFile(ti.FilePath) {
		return nil
	}

	if ti.Content == "" {
		return nil
	}

	errs := validateVerifyResult([]byte(ti.Content))
	if len(errs) == 0 {
		return nil
	}

	msg := fmt.Sprintf("Verification result validation warnings for %s:\n- %s",
		ti.FilePath, strings.Join(errs, "\n- "))
	return &msg
}

// isVerifyResultFile checks if a path matches the verify result file pattern.
func isVerifyResultFile(path string) bool {
	return strings.Contains(path, "verify-") && strings.HasSuffix(path, ".json")
}

// validateVerifyResult checks the JSON structure of a verification result.
func validateVerifyResult(data []byte) []string {
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return []string{fmt.Sprintf("invalid JSON: %v", err)}
	}

	var errs []string

	verdict, ok := result["verdict"]
	if !ok {
		errs = append(errs, "missing required field: verdict")
	} else {
		v, isStr := verdict.(string)
		if !isStr {
			errs = append(errs, "verdict must be a string")
		} else {
			validVerdicts := map[string]bool{"pass": true, "fail": true}
			if !validVerdicts[strings.ToLower(v)] {
				errs = append(errs, fmt.Sprintf("invalid verdict value %q (must be pass or fail)", v))
			}
		}
	}

	if _, ok := result["findings"]; !ok {
		errs = append(errs, "missing required field: findings")
	}

	return errs
}
