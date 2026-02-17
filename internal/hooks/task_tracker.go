package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func init() {
	Register("task-tracker", taskTrackerHook)
}

// taskSummary is the JSON structure persisted to {sessionDir}/tasks.json.
type taskSummary struct {
	Completed int `json:"completed"`
	Total     int `json:"total"`
}

// todoWriteInput matches the TodoWrite tool_input schema.
type todoWriteInput struct {
	Todos []todoItem `json:"todos"`
}

type todoItem struct {
	Status string `json:"status"`
}

// taskUpdateInput matches the TaskUpdate tool_input schema.
type taskUpdateInput struct {
	Status string `json:"status"`
}

// taskTrackerHook intercepts TaskCreate, TaskUpdate, and TodoWrite calls
// to maintain a running task count in the session directory.
func taskTrackerHook(input *Input) error {
	sessionDir := resolveSessionDir()

	summary := loadTaskSummary(sessionDir)

	switch input.ToolName {
	case "TodoWrite":
		summary = countFromTodoWrite(input.ToolInput)
	case "TaskCreate":
		summary.Total++
	case "TaskUpdate":
		var tu taskUpdateInput
		if err := json.Unmarshal(input.ToolInput, &tu); err == nil {
			switch tu.Status {
			case "completed":
				summary.Completed++
			case "deleted":
				if summary.Total > 0 {
					summary.Total--
				}
			}
		}
	}

	saveTaskSummary(sessionDir, &summary)
	ExitOK()
	return nil
}

// countFromTodoWrite parses a TodoWrite tool_input and counts tasks by status.
func countFromTodoWrite(raw json.RawMessage) taskSummary {
	var tw todoWriteInput
	if err := json.Unmarshal(raw, &tw); err != nil {
		return taskSummary{}
	}

	var s taskSummary
	for _, t := range tw.Todos {
		switch t.Status {
		case "completed":
			s.Completed++
			s.Total++
		case "deleted":
			// don't count deleted
		default:
			s.Total++
		}
	}
	return s
}

func tasksFile(sessionDir string) string {
	return filepath.Join(sessionDir, "tasks.json")
}

func loadTaskSummary(sessionDir string) taskSummary {
	data, err := os.ReadFile(tasksFile(sessionDir))
	if err != nil {
		return taskSummary{}
	}
	var s taskSummary
	json.Unmarshal(data, &s) //nolint:errcheck
	return s
}

func saveTaskSummary(sessionDir string, s *taskSummary) {
	os.MkdirAll(sessionDir, 0o755) //nolint:errcheck
	data, _ := json.Marshal(s)
	os.WriteFile(tasksFile(sessionDir), data, 0o644) //nolint:errcheck
}
