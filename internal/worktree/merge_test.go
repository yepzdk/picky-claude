package worktree_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/jesperpedersen/picky-claude/internal/worktree"
)

func TestSync_MultipleCommits(t *testing.T) {
	dir := initGitRepo(t)
	mgr := worktree.NewManager(dir)

	info, err := mgr.Create("multi-commit")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Make multiple commits in the worktree
	for i, name := range []string{"a.txt", "b.txt", "c.txt"} {
		if err := os.WriteFile(filepath.Join(info.Path, name), []byte("content\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		for _, c := range [][]string{
			{"git", "add", name},
			{"git", "commit", "-m", "add " + name},
		} {
			cmd := exec.Command(c[0], c[1:]...)
			cmd.Dir = info.Path
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("commit %d failed: %v\n%s", i, err, out)
			}
		}
	}

	result, err := mgr.Sync("multi-commit")
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
	if !result.Success {
		t.Error("expected success")
	}
	if result.FilesChanged != 3 {
		t.Errorf("expected 3 files changed, got %d", result.FilesChanged)
	}
	if result.CommitHash == "" {
		t.Error("expected non-empty commit hash")
	}

	// Verify all files are in main repo
	for _, name := range []string{"a.txt", "b.txt", "c.txt"} {
		if _, err := os.Stat(filepath.Join(dir, name)); os.IsNotExist(err) {
			t.Errorf("%s should exist in main repo after sync", name)
		}
	}

	// Verify the squash produced a single commit on main (not 3)
	cmd := exec.Command("git", "log", "--oneline")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git log failed: %v\n%s", err, out)
	}
	// Should have exactly 2 commits: initial + squash merge
	lines := 0
	for _, line := range splitLines(string(out)) {
		if line != "" {
			lines++
		}
	}
	if lines != 2 {
		t.Errorf("expected 2 commits (initial + squash), got %d:\n%s", lines, out)
	}
}

func TestSync_NonExistent(t *testing.T) {
	dir := initGitRepo(t)
	mgr := worktree.NewManager(dir)

	_, err := mgr.Sync("nonexistent")
	if err == nil {
		t.Error("expected error syncing non-existent worktree")
	}
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
