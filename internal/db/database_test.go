package db

import (
	"log/slog"
	"os"
	"testing"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func testDB(t *testing.T) *DB {
	t.Helper()
	db, err := OpenInMemory(testLogger())
	if err != nil {
		t.Fatalf("OpenInMemory: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestOpenInMemory(t *testing.T) {
	db := testDB(t)
	if db.Conn() == nil {
		t.Fatal("Conn() returned nil")
	}
}

func TestMigrationsIdempotent(t *testing.T) {
	db := testDB(t)
	// Running migrate again should be a no-op.
	if err := db.migrate(); err != nil {
		t.Fatalf("second migrate call failed: %v", err)
	}
}

func TestObservationCRUD(t *testing.T) {
	db := testDB(t)

	o := &Observation{
		SessionID: "sess-1",
		Type:      "discovery",
		Title:     "Found config bug",
		Text:      "The config loader ignores nested keys",
		Project:   "myproject",
		Metadata:  `{"severity":"high"}`,
	}

	id, err := db.InsertObservation(o)
	if err != nil {
		t.Fatalf("InsertObservation: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}

	got, err := db.GetObservation(id)
	if err != nil {
		t.Fatalf("GetObservation: %v", err)
	}
	if got == nil {
		t.Fatal("GetObservation returned nil")
	}
	if got.Title != o.Title {
		t.Errorf("Title = %q, want %q", got.Title, o.Title)
	}
	if got.Text != o.Text {
		t.Errorf("Text = %q, want %q", got.Text, o.Text)
	}
	if got.SessionID != o.SessionID {
		t.Errorf("SessionID = %q, want %q", got.SessionID, o.SessionID)
	}
}

func TestObservationGetMultiple(t *testing.T) {
	db := testDB(t)

	var ids []int64
	for i := 0; i < 3; i++ {
		id, err := db.InsertObservation(&Observation{
			SessionID: "sess-1",
			Title:     "obs",
			Text:      "text",
		})
		if err != nil {
			t.Fatalf("InsertObservation: %v", err)
		}
		ids = append(ids, id)
	}

	got, err := db.GetObservations(ids)
	if err != nil {
		t.Fatalf("GetObservations: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("got %d observations, want 3", len(got))
	}
}

func TestObservationSearch(t *testing.T) {
	db := testDB(t)

	db.InsertObservation(&Observation{
		SessionID: "sess-1",
		Title:     "authentication bug",
		Text:      "JWT tokens expire too quickly in production",
	})
	db.InsertObservation(&Observation{
		SessionID: "sess-1",
		Title:     "database issue",
		Text:      "Connection pool exhaustion under load",
	})

	results, err := db.SearchObservations("authentication", 10)
	if err != nil {
		t.Fatalf("SearchObservations: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Title != "authentication bug" {
		t.Errorf("Title = %q, want %q", results[0].Title, "authentication bug")
	}
}

func TestFilteredSearch(t *testing.T) {
	db := testDB(t)

	db.InsertObservation(&Observation{
		SessionID: "sess-1",
		Type:      "bugfix",
		Title:     "auth bug",
		Text:      "Fixed authentication token expiry",
		Project:   "backend",
	})
	db.InsertObservation(&Observation{
		SessionID: "sess-1",
		Type:      "feature",
		Title:     "auth feature",
		Text:      "Added authentication rate limiting",
		Project:   "backend",
	})
	db.InsertObservation(&Observation{
		SessionID: "sess-2",
		Type:      "bugfix",
		Title:     "ui bug",
		Text:      "Fixed CSS authentication dialog alignment",
		Project:   "frontend",
	})

	tests := []struct {
		name    string
		filter  SearchFilter
		wantLen int
	}{
		{
			name:    "query only",
			filter:  SearchFilter{Query: "authentication", Limit: 10},
			wantLen: 3,
		},
		{
			name:    "filter by type",
			filter:  SearchFilter{Query: "authentication", Type: "bugfix", Limit: 10},
			wantLen: 2,
		},
		{
			name:    "filter by project",
			filter:  SearchFilter{Query: "authentication", Project: "frontend", Limit: 10},
			wantLen: 1,
		},
		{
			name:    "filter by type and project",
			filter:  SearchFilter{Query: "authentication", Type: "bugfix", Project: "backend", Limit: 10},
			wantLen: 1,
		},
		{
			name:    "no match",
			filter:  SearchFilter{Query: "authentication", Type: "refactor", Limit: 10},
			wantLen: 0,
		},
		{
			name:    "default limit",
			filter:  SearchFilter{Query: "authentication"},
			wantLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := db.FilteredSearch(tt.filter)
			if err != nil {
				t.Fatalf("FilteredSearch: %v", err)
			}
			if len(results) != tt.wantLen {
				t.Errorf("got %d results, want %d", len(results), tt.wantLen)
			}
		})
	}
}

func TestFilteredSearchDateRange(t *testing.T) {
	db := testDB(t)

	// Insert with explicit timestamps via raw SQL
	db.conn.Exec(
		`INSERT INTO observations (session_id, type, title, text, project, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		"sess-1", "discovery", "old item", "old discovery text", "proj",
		"2025-01-01 00:00:00",
	)
	db.conn.Exec(
		`INSERT INTO observations (session_id, type, title, text, project, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		"sess-1", "discovery", "new item", "new discovery text", "proj",
		"2026-02-15 00:00:00",
	)

	results, err := db.FilteredSearch(SearchFilter{
		Query:     "discovery",
		DateStart: "2026-01-01",
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("FilteredSearch: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Title != "new item" {
		t.Errorf("Title = %q, want %q", results[0].Title, "new item")
	}

	results, err = db.FilteredSearch(SearchFilter{
		Query:   "discovery",
		DateEnd: "2025-06-01",
		Limit:   10,
	})
	if err != nil {
		t.Fatalf("FilteredSearch: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Title != "old item" {
		t.Errorf("Title = %q, want %q", results[0].Title, "old item")
	}
}

func TestObservationNotFound(t *testing.T) {
	db := testDB(t)

	got, err := db.GetObservation(99999)
	if err != nil {
		t.Fatalf("GetObservation: %v", err)
	}
	if got != nil {
		t.Error("expected nil for non-existent observation")
	}
}

func TestSessionCRUD(t *testing.T) {
	db := testDB(t)

	s := &Session{
		ID:       "sess-1",
		Project:  "myproject",
		Metadata: "{}",
	}
	if err := db.InsertSession(s); err != nil {
		t.Fatalf("InsertSession: %v", err)
	}

	got, err := db.GetSession("sess-1")
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if got == nil {
		t.Fatal("GetSession returned nil")
	}
	if got.Project != "myproject" {
		t.Errorf("Project = %q, want %q", got.Project, "myproject")
	}
	if got.EndedAt != nil {
		t.Error("EndedAt should be nil for active session")
	}

	if err := db.IncrementMessageCount("sess-1"); err != nil {
		t.Fatalf("IncrementMessageCount: %v", err)
	}
	got, _ = db.GetSession("sess-1")
	if got.MessageCount != 1 {
		t.Errorf("MessageCount = %d, want 1", got.MessageCount)
	}

	if err := db.EndSession("sess-1"); err != nil {
		t.Fatalf("EndSession: %v", err)
	}
	got, _ = db.GetSession("sess-1")
	if got.EndedAt == nil {
		t.Error("EndedAt should be set after EndSession")
	}
}

func TestListActiveSessions(t *testing.T) {
	db := testDB(t)

	db.InsertSession(&Session{ID: "active-1", Project: "p", Metadata: "{}"})
	db.InsertSession(&Session{ID: "active-2", Project: "p", Metadata: "{}"})
	db.InsertSession(&Session{ID: "ended-1", Project: "p", Metadata: "{}"})
	db.EndSession("ended-1")

	active, err := db.ListActiveSessions()
	if err != nil {
		t.Fatalf("ListActiveSessions: %v", err)
	}
	if len(active) != 2 {
		t.Errorf("got %d active sessions, want 2", len(active))
	}
}

func TestSessionNotFound(t *testing.T) {
	db := testDB(t)

	got, err := db.GetSession("nonexistent")
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if got != nil {
		t.Error("expected nil for non-existent session")
	}
}

func TestCleanupStaleSessions(t *testing.T) {
	db := testDB(t)

	// Insert a session with an old start time
	db.conn.Exec(
		`INSERT INTO sessions (id, project, started_at, metadata) VALUES (?, ?, ?, ?)`,
		"stale-1", "proj", "2025-01-01 00:00:00", "{}",
	)
	// Insert a recent active session
	db.InsertSession(&Session{ID: "active-1", Project: "proj", Metadata: "{}"})

	count, err := db.CleanupStaleSessions(24)
	if err != nil {
		t.Fatalf("CleanupStaleSessions: %v", err)
	}
	if count != 1 {
		t.Errorf("cleaned up %d sessions, want 1", count)
	}

	// Verify stale session is ended
	s, _ := db.GetSession("stale-1")
	if s.EndedAt == nil {
		t.Error("stale session should have EndedAt set")
	}

	// Verify active session is untouched
	s, _ = db.GetSession("active-1")
	if s.EndedAt != nil {
		t.Error("active session should not have EndedAt set")
	}
}

func TestSummaryCRUD(t *testing.T) {
	db := testDB(t)

	db.InsertSession(&Session{ID: "sess-1", Metadata: "{}"})

	id, err := db.InsertSummary(&Summary{
		SessionID: "sess-1",
		Text:      "Implemented auth flow with JWT tokens",
	})
	if err != nil {
		t.Fatalf("InsertSummary: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}

	summaries, err := db.RecentSummaries(10)
	if err != nil {
		t.Fatalf("RecentSummaries: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("got %d summaries, want 1", len(summaries))
	}
	if summaries[0].Text != "Implemented auth flow with JWT tokens" {
		t.Errorf("Text = %q", summaries[0].Text)
	}
}

func TestPlanCRUD(t *testing.T) {
	db := testDB(t)

	id, err := db.InsertPlan(&Plan{
		Path:      "docs/plans/2026-02-16-auth.md",
		SessionID: "sess-1",
		Status:    "PENDING",
	})
	if err != nil {
		t.Fatalf("InsertPlan: %v", err)
	}

	got, err := db.GetPlanByPath("docs/plans/2026-02-16-auth.md")
	if err != nil {
		t.Fatalf("GetPlanByPath: %v", err)
	}
	if got == nil {
		t.Fatal("GetPlanByPath returned nil")
	}
	if got.Status != "PENDING" {
		t.Errorf("Status = %q, want PENDING", got.Status)
	}

	if err := db.UpdatePlanStatus(id, "COMPLETE"); err != nil {
		t.Fatalf("UpdatePlanStatus: %v", err)
	}
	got, _ = db.GetPlanByPath("docs/plans/2026-02-16-auth.md")
	if got.Status != "COMPLETE" {
		t.Errorf("Status = %q, want COMPLETE", got.Status)
	}
}

func TestPlanNotFound(t *testing.T) {
	db := testDB(t)

	got, err := db.GetPlanByPath("nonexistent.md")
	if err != nil {
		t.Fatalf("GetPlanByPath: %v", err)
	}
	if got != nil {
		t.Error("expected nil for non-existent plan")
	}
}

func TestPromptCRUD(t *testing.T) {
	db := testDB(t)

	id, err := db.InsertPrompt(&Prompt{
		SessionID: "sess-1",
		Role:      "system",
		Text:      "You are a helpful assistant",
	})
	if err != nil {
		t.Fatalf("InsertPrompt: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}

	prompts, err := db.RecentPrompts("sess-1", 10)
	if err != nil {
		t.Fatalf("RecentPrompts: %v", err)
	}
	if len(prompts) != 1 {
		t.Fatalf("got %d prompts, want 1", len(prompts))
	}
	if prompts[0].Role != "system" {
		t.Errorf("Role = %q, want system", prompts[0].Role)
	}
}

func TestTimelineAround(t *testing.T) {
	db := testDB(t)

	var ids []int64
	for i := 0; i < 10; i++ {
		id, _ := db.InsertObservation(&Observation{
			SessionID: "sess-1",
			Title:     "obs",
			Text:      "text",
		})
		ids = append(ids, id)
	}

	// Get timeline around the 5th observation
	results, err := db.TimelineAround(ids[4], 2, 2)
	if err != nil {
		t.Fatalf("TimelineAround: %v", err)
	}
	if len(results) == 0 {
		t.Error("TimelineAround returned no results")
	}
}
