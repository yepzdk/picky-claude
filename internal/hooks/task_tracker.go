package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/jesperpedersen/picky-claude/internal/config"
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
	sessionID := input.SessionID
	if sessionID == "" {
		sessionID = "default"
	}

	summary := loadTaskSummary(sessionID)

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

	saveTaskSummary(sessionID, &summary)
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

func tasksFile(sessionID string) string {
	return filepath.Join(config.SessionDir(sessionID), "tasks.json")
}

func loadTaskSummary(sessionID string) taskSummary {
	data, err := os.ReadFile(tasksFile(sessionID))
	if err != nil {
		return taskSummary{}
	}
	var s taskSummary
	json.Unmarshal(data, &s) //nolint:errcheck
	return s
}

func saveTaskSummary(sessionID string, s *taskSummary) {
	dir := config.SessionDir(sessionID)
	os.MkdirAll(dir, 0o755) //nolint:errcheck
	data, _ := json.Marshal(s)
	os.WriteFile(tasksFile(sessionID), data, 0o644) //nolint:errcheck
}
