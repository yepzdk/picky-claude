package statusline

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGatherBranch_NormalRef(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	os.MkdirAll(gitDir, 0o755)
	os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main\n"), 0o644)

	got := gatherBranch(dir)
	if got != "main" {
		t.Errorf("gatherBranch() = %q, want %q", got, "main")
	}
}

func TestGatherBranch_FeatureBranch(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	os.MkdirAll(gitDir, 0o755)
	os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/feat/auth\n"), 0o644)

	got := gatherBranch(dir)
	if got != "feat/auth" {
		t.Errorf("gatherBranch() = %q, want %q", got, "feat/auth")
	}
}

func TestGatherBranch_DetachedHead(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	os.MkdirAll(gitDir, 0o755)
	os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("abc1234def5678901234567890abcdef12345678\n"), 0o644)

	got := gatherBranch(dir)
	if got != "abc1234" {
		t.Errorf("gatherBranch() = %q, want %q", got, "abc1234")
	}
}

func TestGatherBranch_NoGitDir(t *testing.T) {
	dir := t.TempDir()
	got := gatherBranch(dir)
	if got != "" {
		t.Errorf("gatherBranch() = %q, want empty", got)
	}
}

func TestGatherContextPct(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(
		filepath.Join(dir, "context-pct.json"),
		[]byte(`{"percentage": 42.5}`),
		0o644,
	)

	got := gatherContextPct(dir)
	if got != 42.5 {
		t.Errorf("gatherContextPct() = %v, want 42.5", got)
	}
}

func TestGatherContextPct_NoFile(t *testing.T) {
	dir := t.TempDir()
	got := gatherContextPct(dir)
	if got != 0 {
		t.Errorf("gatherContextPct() = %v, want 0", got)
	}
}

func TestGatherPlan(t *testing.T) {
	dir := t.TempDir()
	plansDir := filepath.Join(dir, "docs", "plans")
	os.MkdirAll(plansDir, 0o755)

	planContent := `# Add Authentication

Status: PENDING
Approved: Yes

## Tasks

- [x] Task 1
- [ ] Task 2
- [ ] Task 3

Progress: Done 1 / Left 2 / Total 3
`
	os.WriteFile(filepath.Join(plansDir, "2026-02-17-add-auth.md"), []byte(planContent), 0o644)

	got := gatherPlan(dir)
	if got == nil {
		t.Fatal("gatherPlan() should return a plan")
	}
	if got.Name != "add-auth" {
		t.Errorf("plan.Name = %q, want %q", got.Name, "add-auth")
	}
	if got.Status != "PENDING" {
		t.Errorf("plan.Status = %q, want %q", got.Status, "PENDING")
	}
	if got.Done != 1 || got.Total != 3 {
		t.Errorf("plan progress = %d/%d, want 1/3", got.Done, got.Total)
	}
}

func TestGatherPlan_NoPlans(t *testing.T) {
	dir := t.TempDir()
	got := gatherPlan(dir)
	if got != nil {
		t.Errorf("gatherPlan() should return nil when no plans, got %+v", got)
	}
}

func TestGatherTasks(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(
		filepath.Join(dir, "tasks.json"),
		[]byte(`{"completed":2,"total":5}`),
		0o644,
	)

	got := gatherTasks(dir)
	if got == nil {
		t.Fatal("gatherTasks() should return tasks")
	}
	if got.Completed != 2 {
		t.Errorf("tasks.Completed = %d, want 2", got.Completed)
	}
	if got.Total != 5 {
		t.Errorf("tasks.Total = %d, want 5", got.Total)
	}
}

func TestGatherTasks_NoFile(t *testing.T) {
	dir := t.TempDir()
	got := gatherTasks(dir)
	if got != nil {
		t.Errorf("gatherTasks() should return nil when no file, got %+v", got)
	}
}

func TestGatherTasks_ZeroTotal(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(
		filepath.Join(dir, "tasks.json"),
		[]byte(`{"completed":0,"total":0}`),
		0o644,
	)

	got := gatherTasks(dir)
	if got != nil {
		t.Errorf("gatherTasks() should return nil for zero total, got %+v", got)
	}
}

func TestGatherPlan_VerifiedPlanSkipped(t *testing.T) {
	dir := t.TempDir()
	plansDir := filepath.Join(dir, "docs", "plans")
	os.MkdirAll(plansDir, 0o755)

	planContent := `# Done Feature

Status: VERIFIED

## Tasks

- [x] Task 1

Progress: Done 1 / Left 0 / Total 1
`
	os.WriteFile(filepath.Join(plansDir, "2026-02-17-done.md"), []byte(planContent), 0o644)

	got := gatherPlan(dir)
	if got != nil {
		t.Errorf("gatherPlan() should skip VERIFIED plans, got %+v", got)
	}
}
