package worktree

import (
	"fmt"
	"os/exec"
	"strings"
)

// squashMerge performs a squash merge of sourceBranch into the current branch
// in the given repository directory. Returns the resulting commit hash.
func squashMerge(repoDir, sourceBranch, message string) (string, error) {
	if _, err := gitIn(repoDir, "merge", "--squash", sourceBranch); err != nil {
		return "", fmt.Errorf("squash merge %s: %w", sourceBranch, err)
	}

	if _, err := gitIn(repoDir, "commit", "-m", message); err != nil {
		return "", fmt.Errorf("commit squash merge: %w", err)
	}

	out, err := gitIn(repoDir, "rev-parse", "--short", "HEAD")
	if err != nil {
		return "", fmt.Errorf("get commit hash: %w", err)
	}
	return strings.TrimSpace(out), nil
}

// diffNameOnly returns the list of files changed between two refs.
func diffNameOnly(repoDir, base, head string) ([]string, error) {
	out, err := gitIn(repoDir, "diff", "--name-only", base+"..."+head)
	if err != nil {
		return []string{}, nil
	}
	return splitNonEmpty(strings.TrimSpace(out)), nil
}

// gitIn runs a git command in the given directory and returns stdout.
func gitIn(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, out)
	}
	return string(out), nil
}
