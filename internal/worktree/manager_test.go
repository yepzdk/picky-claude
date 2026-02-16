package worktree_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/jesperpedersen/picky-claude/internal/worktree"
)

// initGitRepo creates a temp directory with a git repo containing one commit.
// Returns the repo path and a cleanup function.
func initGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	cmds := [][]string{
		{"git", "init", "-b", "main"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	}
	for _, c := range cmds {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git init step %v failed: %v\n%s", c, err, out)
		}
	}

	// Create initial file and commit
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, c := range [][]string{
		{"git", "add", "."},
		{"git", "commit", "-m", "initial"},
	} {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git commit step %v failed: %v\n%s", c, err, out)
		}
	}

	return dir
}

func TestDetect_NotFound(t *testing.T) {
	dir := initGitRepo(t)
	mgr := worktree.NewManager(dir)

	info, err := mgr.Detect("my-feature")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Found {
		t.Error("expected Found=false for non-existent worktree")
	}
}

func TestCreate_And_Detect(t *testing.T) {
	dir := initGitRepo(t)
	mgr := worktree.NewManager(dir)

	info, err := mgr.Create("my-feature")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if info.Path == "" {
		t.Error("expected non-empty Path")
	}
	if info.Branch != "spec/my-feature" {
		t.Errorf("expected branch spec/my-feature, got %s", info.Branch)
	}
	if info.BaseBranch != "main" {
		t.Errorf("expected base branch main, got %s", info.BaseBranch)
	}

	// Verify worktree directory exists
	if _, err := os.Stat(info.Path); os.IsNotExist(err) {
		t.Errorf("worktree path %s does not exist", info.Path)
	}

	// Detect should find it
	detected, err := mgr.Detect("my-feature")
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}
	if !detected.Found {
		t.Error("expected Found=true after Create")
	}
	if detected.Path != info.Path {
		t.Errorf("Detect path mismatch: got %s, want %s", detected.Path, info.Path)
	}
}

func TestCreate_Duplicate(t *testing.T) {
	dir := initGitRepo(t)
	mgr := worktree.NewManager(dir)

	if _, err := mgr.Create("dup"); err != nil {
		t.Fatalf("first Create failed: %v", err)
	}
	_, err := mgr.Create("dup")
	if err == nil {
		t.Error("expected error when creating duplicate worktree")
	}
}

func TestDiff_NoChanges(t *testing.T) {
	dir := initGitRepo(t)
	mgr := worktree.NewManager(dir)

	if _, err := mgr.Create("diff-test"); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	result, err := mgr.Diff("diff-test")
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if len(result.Files) != 0 {
		t.Errorf("expected 0 changed files, got %d", len(result.Files))
	}
}

func TestDiff_WithChanges(t *testing.T) {
	dir := initGitRepo(t)
	mgr := worktree.NewManager(dir)

	info, err := mgr.Create("diff-changes")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Make a change in the worktree and commit
	newFile := filepath.Join(info.Path, "new.txt")
	if err := os.WriteFile(newFile, []byte("hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, c := range [][]string{
		{"git", "add", "."},
		{"git", "commit", "-m", "add new file"},
	} {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = info.Path
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("commit in worktree failed: %v\n%s", err, out)
		}
	}

	result, err := mgr.Diff("diff-changes")
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if len(result.Files) != 1 {
		t.Errorf("expected 1 changed file, got %d: %v", len(result.Files), result.Files)
	}
}

func TestSync(t *testing.T) {
	dir := initGitRepo(t)
	mgr := worktree.NewManager(dir)

	info, err := mgr.Create("sync-test")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Commit a change in the worktree
	newFile := filepath.Join(info.Path, "synced.txt")
	if err := os.WriteFile(newFile, []byte("synced\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, c := range [][]string{
		{"git", "add", "."},
		{"git", "commit", "-m", "add synced file"},
	} {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = info.Path
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("commit in worktree failed: %v\n%s", err, out)
		}
	}

	result, err := mgr.Sync("sync-test")
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
	if !result.Success {
		t.Error("expected Sync success")
	}
	if result.FilesChanged == 0 {
		t.Error("expected FilesChanged > 0")
	}

	// Verify the file exists in the main repo
	if _, err := os.Stat(filepath.Join(dir, "synced.txt")); os.IsNotExist(err) {
		t.Error("synced.txt should exist in main repo after sync")
	}
}

func TestCleanup(t *testing.T) {
	dir := initGitRepo(t)
	mgr := worktree.NewManager(dir)

	info, err := mgr.Create("cleanup-test")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := mgr.Cleanup("cleanup-test"); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Worktree directory should be gone
	if _, err := os.Stat(info.Path); !os.IsNotExist(err) {
		t.Error("expected worktree directory to be removed after cleanup")
	}

	// Detect should return not found
	detected, err := mgr.Detect("cleanup-test")
	if err != nil {
		t.Fatalf("Detect after cleanup failed: %v", err)
	}
	if detected.Found {
		t.Error("expected Found=false after cleanup")
	}
}

func TestStatus_NoActiveWorktree(t *testing.T) {
	dir := initGitRepo(t)
	mgr := worktree.NewManager(dir)

	status, err := mgr.Status()
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if status.Active {
		t.Error("expected Active=false when no worktrees exist")
	}
}

func TestStatus_ActiveWorktree(t *testing.T) {
	dir := initGitRepo(t)
	mgr := worktree.NewManager(dir)

	if _, err := mgr.Create("status-test"); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	status, err := mgr.Status()
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if !status.Active {
		t.Error("expected Active=true when worktree exists")
	}
	if status.Slug != "status-test" {
		t.Errorf("expected slug status-test, got %s", status.Slug)
	}
}
