package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jesperpedersen/picky-claude/internal/config"
)

func setupTaskTrackerTest(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	orig := os.Getenv(config.EnvPrefix + "_HOME")
	os.Setenv(config.EnvPrefix+"_HOME", tmpDir)
	return tmpDir, func() {
		if orig == "" {
			os.Unsetenv(config.EnvPrefix + "_HOME")
		} else {
			os.Setenv(config.EnvPrefix+"_HOME", orig)
		}
	}
}

func readSummary(t *testing.T, sessionDir string) taskSummary {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(sessionDir, "tasks.json"))
	if err != nil {
		t.Fatalf("reading tasks.json: %v", err)
	}
	var s taskSummary
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatalf("parsing tasks.json: %v", err)
	}
	return s
}

func TestCountFromTodoWrite(t *testing.T) {
	tests := []struct {
		name      string
		input     todoWriteInput
		wantDone  int
		wantTotal int
	}{
		{
			name: "mixed statuses",
			input: todoWriteInput{
				Todos: []todoItem{
					{Status: "completed"},
					{Status: "pending"},
					{Status: "in_progress"},
					{Status: "completed"},
				},
			},
			wantDone:  2,
			wantTotal: 4,
		},
		{
			name: "all completed",
			input: todoWriteInput{
				Todos: []todoItem{
					{Status: "completed"},
					{Status: "completed"},
				},
			},
			wantDone:  2,
			wantTotal: 2,
		},
		{
			name: "none completed",
			input: todoWriteInput{
				Todos: []todoItem{
					{Status: "pending"},
					{Status: "in_progress"},
				},
			},
			wantDone:  0,
			wantTotal: 2,
		},
		{
			name: "with deleted",
			input: todoWriteInput{
				Todos: []todoItem{
					{Status: "completed"},
					{Status: "deleted"},
					{Status: "pending"},
				},
			},
			wantDone:  1,
			wantTotal: 2,
		},
		{
			name:      "empty",
			input:     todoWriteInput{},
			wantDone:  0,
			wantTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, _ := json.Marshal(tt.input)
			got := countFromTodoWrite(raw)
			if got.Completed != tt.wantDone {
				t.Errorf("completed = %d, want %d", got.Completed, tt.wantDone)
			}
			if got.Total != tt.wantTotal {
				t.Errorf("total = %d, want %d", got.Total, tt.wantTotal)
			}
		})
	}
}

func TestLoadSaveTaskSummary(t *testing.T) {
	sessionDir := t.TempDir()

	// Load from nonexistent file returns zero
	got := loadTaskSummary(sessionDir)
	if got.Completed != 0 || got.Total != 0 {
		t.Errorf("expected zero summary, got %+v", got)
	}

	// Save and reload
	s := &taskSummary{Completed: 3, Total: 7}
	saveTaskSummary(sessionDir, s)

	got = loadTaskSummary(sessionDir)
	if got.Completed != 3 || got.Total != 7 {
		t.Errorf("expected {3,7}, got %+v", got)
	}
}

func TestTaskTrackerTaskCreate(t *testing.T) {
	sessionDir := t.TempDir()

	// Simulate three TaskCreate calls.
	for i := 0; i < 3; i++ {
		summary := loadTaskSummary(sessionDir)
		summary.Total++
		saveTaskSummary(sessionDir, &summary)
	}

	got := readSummary(t, sessionDir)
	if got.Total != 3 {
		t.Errorf("total = %d, want 3", got.Total)
	}
	if got.Completed != 0 {
		t.Errorf("completed = %d, want 0", got.Completed)
	}
}

func TestTaskTrackerTaskUpdateCompleted(t *testing.T) {
	sessionDir := t.TempDir()

	// Set up initial state: 3 tasks, 0 completed
	saveTaskSummary(sessionDir, &taskSummary{Total: 3})

	// Simulate TaskUpdate with status=completed
	summary := loadTaskSummary(sessionDir)
	var tu taskUpdateInput
	json.Unmarshal([]byte(`{"status":"completed"}`), &tu)
	if tu.Status == "completed" {
		summary.Completed++
	}
	saveTaskSummary(sessionDir, &summary)

	got := readSummary(t, sessionDir)
	if got.Total != 3 {
		t.Errorf("total = %d, want 3", got.Total)
	}
	if got.Completed != 1 {
		t.Errorf("completed = %d, want 1", got.Completed)
	}
}

func TestTaskTrackerTaskUpdateDeleted(t *testing.T) {
	sessionDir := t.TempDir()

	// Set up initial state: 3 tasks, 1 completed
	saveTaskSummary(sessionDir, &taskSummary{Completed: 1, Total: 3})

	// Simulate TaskUpdate with status=deleted
	summary := loadTaskSummary(sessionDir)
	var tu taskUpdateInput
	json.Unmarshal([]byte(`{"status":"deleted"}`), &tu)
	if tu.Status == "deleted" && summary.Total > 0 {
		summary.Total--
	}
	saveTaskSummary(sessionDir, &summary)

	got := readSummary(t, sessionDir)
	if got.Total != 2 {
		t.Errorf("total = %d, want 2", got.Total)
	}
	if got.Completed != 1 {
		t.Errorf("completed = %d, want 1", got.Completed)
	}
}
