package worktree

import (
	"fmt"
	"os/exec"
	"strings"
)

// stashIfDirty stashes any uncommitted changes in the given repo directory.
// Returns true if a stash was created, false if the tree was clean.
func stashIfDirty(repoDir string) (bool, error) {
	if !isDirty(repoDir) {
		return false, nil
	}

	cmd := exec.Command("git", "stash", "push", "-u", "-m", "picky: auto-stash before worktree create")
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return false, fmt.Errorf("git stash push: %w\n%s", err, out)
	}
	return true, nil
}

// stashPop restores the most recent stash entry.
func stashPop(repoDir string) error {
	cmd := exec.Command("git", "stash", "pop")
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git stash pop: %w\n%s", err, out)
	}
	return nil
}

// isDirty returns true if the working tree has uncommitted changes (staged,
// unstaged, or untracked files).
func isDirty(repoDir string) bool {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}
