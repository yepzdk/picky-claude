package installer

import (
	"errors"
	"testing"
)

// mockStep implements Step for testing.
type mockStep struct {
	name       string
	runErr     error
	rolled     bool
	rollbackFn func()
}

func (m *mockStep) Name() string { return m.name }
func (m *mockStep) Run(ctx *Context) error {
	return m.runErr
}
func (m *mockStep) Rollback(ctx *Context) {
	m.rolled = true
	if m.rollbackFn != nil {
		m.rollbackFn()
	}
}

func TestInstaller_AllStepsSucceed(t *testing.T) {
	s1 := &mockStep{name: "step-1"}
	s2 := &mockStep{name: "step-2"}
	s3 := &mockStep{name: "step-3"}

	inst := New(t.TempDir(), s1, s2, s3)
	result := inst.Run()

	if !result.Success {
		t.Errorf("expected success, got failure: %v", result.Error)
	}
	if len(result.Completed) != 3 {
		t.Errorf("expected 3 completed steps, got %d", len(result.Completed))
	}
	if result.FailedStep != "" {
		t.Errorf("expected no failed step, got %q", result.FailedStep)
	}
}

func TestInstaller_StepFails_RollsBack(t *testing.T) {
	s1 := &mockStep{name: "step-1"}
	s2 := &mockStep{name: "step-2", runErr: errors.New("boom")}
	s3 := &mockStep{name: "step-3"}

	inst := New(t.TempDir(), s1, s2, s3)
	result := inst.Run()

	if result.Success {
		t.Error("expected failure")
	}
	if result.FailedStep != "step-2" {
		t.Errorf("expected failed step step-2, got %q", result.FailedStep)
	}
	// Only step-1 should have completed (step-2 failed, step-3 never ran)
	if len(result.Completed) != 1 {
		t.Errorf("expected 1 completed step, got %d: %v", len(result.Completed), result.Completed)
	}

	// step-1 should be rolled back (it completed before step-2 failed)
	if !s1.rolled {
		t.Error("expected step-1 to be rolled back")
	}
	// step-2 failed, so it should NOT be rolled back
	if s2.rolled {
		t.Error("step-2 should not be rolled back (it failed)")
	}
	// step-3 never ran, should not be rolled back
	if s3.rolled {
		t.Error("step-3 should not be rolled back (never ran)")
	}
}

func TestInstaller_FirstStepFails_NoRollback(t *testing.T) {
	s1 := &mockStep{name: "step-1", runErr: errors.New("fail")}
	s2 := &mockStep{name: "step-2"}

	inst := New(t.TempDir(), s1, s2)
	result := inst.Run()

	if result.Success {
		t.Error("expected failure")
	}
	if len(result.Completed) != 0 {
		t.Errorf("expected 0 completed steps, got %d", len(result.Completed))
	}
	if s1.rolled {
		t.Error("step-1 should not be rolled back (it failed, didn't complete)")
	}
}

func TestInstaller_RollbackOrder(t *testing.T) {
	var order []string
	s1 := &mockStep{name: "step-1", rollbackFn: func() { order = append(order, "step-1") }}
	s2 := &mockStep{name: "step-2", rollbackFn: func() { order = append(order, "step-2") }}
	s3 := &mockStep{name: "step-3", runErr: errors.New("fail")}

	inst := New(t.TempDir(), s1, s2, s3)
	inst.Run()

	// Rollback should be in reverse order
	if len(order) != 2 {
		t.Fatalf("expected 2 rollbacks, got %d", len(order))
	}
	if order[0] != "step-2" || order[1] != "step-1" {
		t.Errorf("expected rollback order [step-2, step-1], got %v", order)
	}
}

func TestContext_ProjectDir(t *testing.T) {
	dir := t.TempDir()
	inst := New(dir)
	ctx := inst.context()

	if ctx.ProjectDir != dir {
		t.Errorf("expected project dir %s, got %s", dir, ctx.ProjectDir)
	}
}
