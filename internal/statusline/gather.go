package statusline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Gather populates an Input from the filesystem and environment.
// workDir is the working directory (for .git/ and docs/plans/).
// sessionDir is the session state directory (for context-pct.json).
// Fields already set on input are not overwritten.
func Gather(input *Input, workDir, sessionDir string) {
	if input.Branch == "" {
		input.Branch = gatherBranch(workDir)
	}
	if input.ContextPct == 0 && sessionDir != "" {
		input.ContextPct = gatherContextPct(sessionDir)
	}
	if input.Plan == nil {
		input.Plan = gatherPlan(workDir)
	}
	if input.Tasks == nil && sessionDir != "" {
		input.Tasks = gatherTasks(sessionDir)
	}
}

// gatherBranch reads the current git branch from .git/HEAD.
// Returns short hash for detached HEAD, empty string if not a git repo.
func gatherBranch(workDir string) string {
	data, err := os.ReadFile(filepath.Join(workDir, ".git", "HEAD"))
	if err != nil {
		return ""
	}
	head := strings.TrimSpace(string(data))

	// Normal branch: "ref: refs/heads/main"
	if strings.HasPrefix(head, "ref: refs/heads/") {
		return strings.TrimPrefix(head, "ref: refs/heads/")
	}

	// Detached HEAD: raw commit hash — show first 7 chars
	if len(head) >= 7 {
		return head[:7]
	}
	return head
}

// gatherContextPct reads the context percentage from the session directory.
func gatherContextPct(sessionDir string) float64 {
	data, err := os.ReadFile(filepath.Join(sessionDir, "context-pct.json"))
	if err != nil {
		return 0
	}
	var d struct {
		Percentage float64 `json:"percentage"`
	}
	if err := json.Unmarshal(data, &d); err != nil {
		return 0
	}
	return d.Percentage
}

// gatherTasks reads the task summary from the session directory.
func gatherTasks(sessionDir string) *Tasks {
	data, err := os.ReadFile(filepath.Join(sessionDir, "tasks.json"))
	if err != nil {
		return nil
	}
	var t Tasks
	if err := json.Unmarshal(data, &t); err != nil {
		return nil
	}
	if t.Total == 0 {
		return nil
	}
	return &t
}

var (
	statusRe = regexp.MustCompile(`(?m)^Status:\s*(\S+)`)
	taskDone = regexp.MustCompile(`- \[x\]`)
	taskTodo = regexp.MustCompile(`- \[ \]`)
)

// gatherPlan finds the most recent active plan (non-VERIFIED) in docs/plans/.
func gatherPlan(workDir string) *Plan {
	plansDir := filepath.Join(workDir, "docs", "plans")
	entries, err := os.ReadDir(plansDir)
	if err != nil {
		return nil
	}

	// Collect .md files, sorted by name descending (newest first)
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			files = append(files, e.Name())
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(files)))

	for _, name := range files {
		data, err := os.ReadFile(filepath.Join(plansDir, name))
		if err != nil {
			continue
		}
		content := string(data)

		// Parse status
		matches := statusRe.FindStringSubmatch(content)
		if len(matches) < 2 {
			continue
		}
		status := matches[1]

		// Skip completed plans
		if status == "VERIFIED" {
			continue
		}

		// Count tasks
		done := len(taskDone.FindAllString(content, -1))
		todo := len(taskTodo.FindAllString(content, -1))

		// Extract slug from filename: "2026-02-17-add-auth.md" → "add-auth"
		slug := strings.TrimSuffix(name, ".md")
		if parts := strings.SplitN(slug, "-", 4); len(parts) == 4 {
			slug = parts[3]
		}

		return &Plan{
			Name:   slug,
			Status: status,
			Done:   done,
			Total:  done + todo,
		}
	}

	return nil
}
