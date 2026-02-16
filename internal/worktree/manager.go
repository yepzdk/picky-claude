// Package worktree manages git worktrees for isolated spec development.
// Worktrees are created at .worktrees/spec-<slug>-<hash>/ with branch spec/<slug>.
package worktree

import (
	"crypto/sha256"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// WorktreeInfo holds metadata about a worktree.
type WorktreeInfo struct {
	Found      bool   `json:"found"`
	Path       string `json:"path,omitempty"`
	Branch     string `json:"branch,omitempty"`
	BaseBranch string `json:"base_branch,omitempty"`
}

// DiffResult contains the list of files changed in a worktree relative to its base.
type DiffResult struct {
	Files []string `json:"files"`
}

// SyncResult holds the outcome of a squash merge back to the base branch.
type SyncResult struct {
	Success      bool   `json:"success"`
	FilesChanged int    `json:"files_changed"`
	CommitHash   string `json:"commit_hash,omitempty"`
}

// StatusInfo describes the currently active worktree (if any).
type StatusInfo struct {
	Active     bool   `json:"active"`
	Slug       string `json:"slug,omitempty"`
	Path       string `json:"path,omitempty"`
	Branch     string `json:"branch,omitempty"`
	BaseBranch string `json:"base_branch,omitempty"`
}

// Manager provides the full worktree lifecycle: create, detect, diff, sync, cleanup.
type Manager struct {
	repoDir string
}

// NewManager creates a Manager rooted at the given git repository directory.
// The path is resolved to handle symlinks (e.g. /var → /private/var on macOS).
func NewManager(repoDir string) *Manager {
	resolved, err := filepath.EvalSymlinks(repoDir)
	if err != nil {
		resolved = repoDir
	}
	return &Manager{repoDir: resolved}
}

// branchName returns the git branch name for a slug.
func branchName(slug string) string {
	return "spec/" + slug
}

// worktreeDirName returns the directory name under .worktrees/ for a slug.
func worktreeDirName(slug string) string {
	h := sha256.Sum256([]byte(slug))
	return fmt.Sprintf("spec-%s-%x", slug, h[:4])
}

// worktreePath returns the full path to a worktree directory.
func (m *Manager) worktreePath(slug string) string {
	return filepath.Join(m.repoDir, ".worktrees", worktreeDirName(slug))
}

// currentBranch returns the current branch name.
func (m *Manager) currentBranch() (string, error) {
	out, err := m.git("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// Detect checks whether a worktree for the given slug exists.
func (m *Manager) Detect(slug string) (*WorktreeInfo, error) {
	branch := branchName(slug)
	wtPath := m.worktreePath(slug)

	out, err := m.git("worktree", "list", "--porcelain")
	if err != nil {
		return &WorktreeInfo{Found: false}, nil
	}

	// Parse porcelain output for matching branch
	for _, block := range parseWorktreeBlocks(out) {
		if block.branch == "refs/heads/"+branch {
			baseBranch, _ := m.currentBranch()
			return &WorktreeInfo{
				Found:      true,
				Path:       block.path,
				Branch:     branch,
				BaseBranch: baseBranch,
			}, nil
		}
		// Also match by path
		if block.path == wtPath {
			baseBranch, _ := m.currentBranch()
			return &WorktreeInfo{
				Found:      true,
				Path:       block.path,
				Branch:     branch,
				BaseBranch: baseBranch,
			}, nil
		}
	}

	return &WorktreeInfo{Found: false}, nil
}

// Create creates a new worktree and branch for the given slug.
// Returns an error if a worktree for this slug already exists.
func (m *Manager) Create(slug string) (*WorktreeInfo, error) {
	// Check for existing worktree
	existing, err := m.Detect(slug)
	if err != nil {
		return nil, fmt.Errorf("detect existing: %w", err)
	}
	if existing.Found {
		return nil, fmt.Errorf("worktree for slug %q already exists at %s", slug, existing.Path)
	}

	baseBranch, err := m.currentBranch()
	if err != nil {
		return nil, fmt.Errorf("get current branch: %w", err)
	}

	branch := branchName(slug)
	wtPath := m.worktreePath(slug)

	// Auto-stash uncommitted changes before worktree creation
	stashed, err := stashIfDirty(m.repoDir)
	if err != nil {
		return nil, fmt.Errorf("auto-stash: %w", err)
	}

	// Create the branch and worktree in one command
	if _, err := m.git("worktree", "add", "-b", branch, wtPath); err != nil {
		// Restore stash on failure
		if stashed {
			stashPop(m.repoDir) //nolint:errcheck
		}
		return nil, fmt.Errorf("git worktree add: %w", err)
	}

	// Restore stashed changes
	if stashed {
		if err := stashPop(m.repoDir); err != nil {
			return nil, fmt.Errorf("auto-stash restore: %w", err)
		}
	}

	return &WorktreeInfo{
		Found:      true,
		Path:       wtPath,
		Branch:     branch,
		BaseBranch: baseBranch,
	}, nil
}

// Diff returns the files changed in the worktree branch compared to its base.
func (m *Manager) Diff(slug string) (*DiffResult, error) {
	info, err := m.Detect(slug)
	if err != nil {
		return nil, err
	}
	if !info.Found {
		return nil, fmt.Errorf("worktree for slug %q not found", slug)
	}

	branch := branchName(slug)
	baseBranch, err := m.currentBranch()
	if err != nil {
		return nil, fmt.Errorf("get base branch: %w", err)
	}

	files, _ := diffNameOnly(m.repoDir, baseBranch, branch)
	return &DiffResult{Files: files}, nil
}

// Sync performs a squash merge of the worktree branch into the base branch.
func (m *Manager) Sync(slug string) (*SyncResult, error) {
	info, err := m.Detect(slug)
	if err != nil {
		return nil, err
	}
	if !info.Found {
		return nil, fmt.Errorf("worktree for slug %q not found", slug)
	}

	branch := branchName(slug)
	baseBranch, err := m.currentBranch()
	if err != nil {
		return nil, fmt.Errorf("get base branch: %w", err)
	}

	files, _ := diffNameOnly(m.repoDir, baseBranch, branch)

	msg := fmt.Sprintf("spec/%s: squash merge from worktree", slug)
	commitHash, err := squashMerge(m.repoDir, branch, msg)
	if err != nil {
		return nil, err
	}

	return &SyncResult{
		Success:      true,
		FilesChanged: len(files),
		CommitHash:   commitHash,
	}, nil
}

// Cleanup removes the worktree directory and deletes the branch.
func (m *Manager) Cleanup(slug string) error {
	branch := branchName(slug)
	wtPath := m.worktreePath(slug)

	// Remove the worktree (--force handles if there are untracked files)
	if _, err := m.git("worktree", "remove", "--force", wtPath); err != nil {
		// Worktree might already be gone, try prune
		m.git("worktree", "prune") //nolint:errcheck
	}

	// Delete the branch
	if _, err := m.git("branch", "-D", branch); err != nil {
		// Branch might already be gone — not an error
		return nil
	}

	return nil
}

// Status returns info about the first active spec worktree found.
func (m *Manager) Status() (*StatusInfo, error) {
	out, err := m.git("worktree", "list", "--porcelain")
	if err != nil {
		return &StatusInfo{Active: false}, nil
	}

	baseBranch, _ := m.currentBranch()

	for _, block := range parseWorktreeBlocks(out) {
		if strings.HasPrefix(block.branch, "refs/heads/spec/") {
			slug := strings.TrimPrefix(block.branch, "refs/heads/spec/")
			return &StatusInfo{
				Active:     true,
				Slug:       slug,
				Path:       block.path,
				Branch:     "spec/" + slug,
				BaseBranch: baseBranch,
			}, nil
		}
	}

	return &StatusInfo{Active: false}, nil
}

// git runs a git command in the repository directory and returns stdout.
func (m *Manager) git(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = m.repoDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, out)
	}
	return string(out), nil
}

// worktreeBlock represents a parsed block from `git worktree list --porcelain`.
type worktreeBlock struct {
	path   string
	branch string
}

// parseWorktreeBlocks parses the porcelain output of `git worktree list`.
func parseWorktreeBlocks(output string) []worktreeBlock {
	var blocks []worktreeBlock
	var current worktreeBlock

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			if current.path != "" {
				blocks = append(blocks, current)
				current = worktreeBlock{}
			}
			continue
		}
		if strings.HasPrefix(line, "worktree ") {
			current.path = strings.TrimPrefix(line, "worktree ")
		} else if strings.HasPrefix(line, "branch ") {
			current.branch = strings.TrimPrefix(line, "branch ")
		}
	}
	// Don't forget the last block if output doesn't end with blank line
	if current.path != "" {
		blocks = append(blocks, current)
	}

	return blocks
}

// splitNonEmpty splits a string by newlines and removes empty entries.
func splitNonEmpty(s string) []string {
	if s == "" {
		return []string{}
	}
	var result []string
	for _, line := range strings.Split(s, "\n") {
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}
