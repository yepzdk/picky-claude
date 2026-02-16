package hooks

import (
	"encoding/json"
	"testing"
)

func TestValidateVerifyResult_Valid(t *testing.T) {
	result := map[string]any{
		"verdict":  "pass",
		"findings": []any{},
	}
	data, _ := json.Marshal(result)

	errs := validateVerifyResult(data)
	if len(errs) > 0 {
		t.Errorf("expected no errors for valid result, got: %v", errs)
	}
}

func TestValidateVerifyResult_WithFindings(t *testing.T) {
	result := map[string]any{
		"verdict": "fail",
		"findings": []any{
			map[string]any{
				"severity": "must_fix",
				"file":     "src/main.go",
				"message":  "missing error handling",
			},
		},
	}
	data, _ := json.Marshal(result)

	errs := validateVerifyResult(data)
	if len(errs) > 0 {
		t.Errorf("expected no errors for valid result with findings, got: %v", errs)
	}
}

func TestValidateVerifyResult_MissingVerdict(t *testing.T) {
	result := map[string]any{
		"findings": []any{},
	}
	data, _ := json.Marshal(result)

	errs := validateVerifyResult(data)
	if len(errs) == 0 {
		t.Error("expected error for missing verdict")
	}
}

func TestValidateVerifyResult_MissingFindings(t *testing.T) {
	result := map[string]any{
		"verdict": "pass",
	}
	data, _ := json.Marshal(result)

	errs := validateVerifyResult(data)
	if len(errs) == 0 {
		t.Error("expected error for missing findings")
	}
}

func TestValidateVerifyResult_InvalidVerdict(t *testing.T) {
	result := map[string]any{
		"verdict":  "maybe",
		"findings": []any{},
	}
	data, _ := json.Marshal(result)

	errs := validateVerifyResult(data)
	if len(errs) == 0 {
		t.Error("expected error for invalid verdict value")
	}
}

func TestValidateVerifyResult_InvalidJSON(t *testing.T) {
	errs := validateVerifyResult([]byte("not json"))
	if len(errs) == 0 {
		t.Error("expected error for invalid JSON")
	}
}

func TestSpecVerifyValidator_NonVerifyFile(t *testing.T) {
	result := specVerifyValidatorCheck(&Input{
		ToolName:  "Write",
		ToolInput: []byte(`{"file_path": "/tmp/src/main.go"}`),
	})
	if result != nil {
		t.Error("expected nil for non-verify file")
	}
}

func TestSpecVerifyValidator_VerifyFile(t *testing.T) {
	content := `{"verdict": "pass", "findings": []}`
	input := map[string]string{
		"file_path": "/tmp/.picky/sessions/123/verify-compliance.json",
		"content":   content,
	}
	data, _ := json.Marshal(input)

	result := specVerifyValidatorCheck(&Input{
		ToolName:  "Write",
		ToolInput: data,
	})
	if result != nil {
		t.Errorf("expected nil for valid verify file, got: %v", result)
	}
}
