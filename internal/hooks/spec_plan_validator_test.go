package hooks

import "testing"

func TestValidatePlan_ValidPlan(t *testing.T) {
	content := `# My Plan

Status: PENDING
Worktree: No

## Tasks
- [ ] Task 1
- [ ] Task 2

## Progress
Done: 0 | Left: 2
`
	errs := validatePlanContent(content)
	if len(errs) > 0 {
		t.Errorf("expected no errors for valid plan, got: %v", errs)
	}
}

func TestValidatePlan_MissingStatus(t *testing.T) {
	content := `# My Plan

Worktree: No

## Tasks
- [ ] Task 1
`
	errs := validatePlanContent(content)
	if len(errs) == 0 {
		t.Error("expected error for missing Status field")
	}
	found := false
	for _, e := range errs {
		if e == "missing required field: Status" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'missing required field: Status' in errors, got: %v", errs)
	}
}

func TestValidatePlan_MissingWorktree(t *testing.T) {
	content := `# My Plan

Status: PENDING

## Tasks
- [ ] Task 1
`
	errs := validatePlanContent(content)
	if len(errs) == 0 {
		t.Error("expected error for missing Worktree field")
	}
}

func TestValidatePlan_MissingTasks(t *testing.T) {
	content := `# My Plan

Status: PENDING
Worktree: No
`
	errs := validatePlanContent(content)
	if len(errs) == 0 {
		t.Error("expected error for missing Tasks section")
	}
}

func TestValidatePlan_InvalidStatus(t *testing.T) {
	content := `# My Plan

Status: INVALID_VALUE
Worktree: No

## Tasks
- [ ] Task 1
`
	errs := validatePlanContent(content)
	if len(errs) == 0 {
		t.Error("expected error for invalid Status value")
	}
}

func TestValidatePlan_AllStatusValues(t *testing.T) {
	for _, status := range []string{"PENDING", "COMPLETE", "VERIFIED"} {
		content := "# Plan\n\nStatus: " + status + "\nWorktree: No\n\n## Tasks\n- [ ] T\n"
		errs := validatePlanContent(content)
		if len(errs) > 0 {
			t.Errorf("Status %s should be valid, got errors: %v", status, errs)
		}
	}
}

func TestSpecPlanValidator_NonPlanFile(t *testing.T) {
	result := specPlanValidatorCheck(&Input{
		ToolName:  "Write",
		ToolInput: []byte(`{"file_path": "/tmp/foo/src/main.go"}`),
	})
	if result != nil {
		t.Error("expected nil for non-plan file")
	}
}

func TestSpecPlanValidator_PlanFile(t *testing.T) {
	result := specPlanValidatorCheck(&Input{
		ToolName:  "Write",
		ToolInput: []byte(`{"file_path": "/tmp/foo/docs/plans/2026-01-01-test.md", "content": "# Plan\n\nStatus: PENDING\nWorktree: No\n\n## Tasks\n- [ ] T\n"}`),
	})
	if result != nil {
		t.Errorf("expected nil for valid plan, got: %v", result)
	}
}

func TestSpecPlanValidator_InvalidPlanFile(t *testing.T) {
	result := specPlanValidatorCheck(&Input{
		ToolName:  "Write",
		ToolInput: []byte(`{"file_path": "/tmp/foo/docs/plans/2026-01-01-test.md", "content": "# Plan\n\nNo status here\n"}`),
	})
	if result == nil {
		t.Error("expected error message for invalid plan file")
	}
}
