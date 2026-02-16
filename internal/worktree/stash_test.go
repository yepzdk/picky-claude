package worktree_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/jesperpedersen/picky-claude/internal/worktree"
)

func TestCreate_WithDirtyWorkingTree_Staged(t *testing.T) {
	dir := initGitRepo(t)
	mgr := worktree.NewManager(dir)

	// Stage a new file (without committing)
	if err := os.WriteFile(filepath.Join(dir, "dirty.txt"), []byte("dirty\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "add", "dirty.txt")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add failed: %v\n%s", err, out)
	}

	// Create should succeed despite dirty tree (auto-stash)
	info, err := mgr.Create("stash-staged")
	if err != nil {
		t.Fatalf("Create with dirty tree failed: %v", err)
	}
	if info.Path == "" {
		t.Error("expected non-empty Path")
	}

	// The dirty file should be restored in the main working tree
	if _, err := os.Stat(filepath.Join(dir, "dirty.txt")); os.IsNotExist(err) {
		t.Error("dirty.txt should exist in main repo after auto-stash restore")
	}
}

func TestCreate_WithDirtyWorkingTree_Unstaged(t *testing.T) {
	dir := initGitRepo(t)
	mgr := worktree.NewManager(dir)

	// Modify a tracked file without staging
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Modified\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create should succeed despite dirty tree (auto-stash)
	info, err := mgr.Create("stash-unstaged")
	if err != nil {
		t.Fatalf("Create with modified tracked file failed: %v", err)
	}
	if info.Path == "" {
		t.Error("expected non-empty Path")
	}

	// The modification should be restored
	data, err := os.ReadFile(filepath.Join(dir, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "# Modified\n" {
		t.Errorf("expected modified content to be restored, got %q", string(data))
	}
}

func TestCreate_CleanWorkingTree(t *testing.T) {
	dir := initGitRepo(t)
	mgr := worktree.NewManager(dir)

	// Create on clean tree should work fine (no stash needed)
	info, err := mgr.Create("clean-test")
	if err != nil {
		t.Fatalf("Create with clean tree failed: %v", err)
	}
	if info.Path == "" {
		t.Error("expected non-empty Path")
	}
}
