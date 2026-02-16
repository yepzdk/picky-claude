package search

import (
	"testing"

	"github.com/jesperpedersen/picky-claude/internal/db"
)

func seedObservations(t *testing.T, database *db.DB) []int64 {
	t.Helper()
	obs := []struct {
		title string
		text  string
		typ   string
	}{
		{"auth bug", "Fixed authentication login flow error in session handler", "bugfix"},
		{"db migration", "Database migration schema updated for users table", "feature"},
		{"auth token", "Authentication session token expiration handling improved", "bugfix"},
		{"css fix", "Fixed CSS flexbox layout issue in the admin dashboard", "bugfix"},
		{"api rate limit", "Added rate limiting to the REST API endpoints", "feature"},
	}

	var ids []int64
	for _, o := range obs {
		id, err := database.InsertObservation(&db.Observation{
			SessionID: "s1",
			Type:      o.typ,
			Title:     o.title,
			Text:      o.text,
			Project:   "test",
		})
		if err != nil {
			t.Fatalf("InsertObservation: %v", err)
		}
		ids = append(ids, id)
	}
	return ids
}

func TestOrchestratorSearch(t *testing.T) {
	database := testDB(t)
	ids := seedObservations(t, database)

	orch, err := NewOrchestrator(database)
	if err != nil {
		t.Fatalf("NewOrchestrator: %v", err)
	}

	if err := orch.RebuildIndex(); err != nil {
		t.Fatalf("RebuildIndex: %v", err)
	}

	results, err := orch.Search(SearchQuery{
		Text:  "authentication login session",
		Limit: 5,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Search returned no results")
	}

	// Auth-related observations should rank higher
	firstID := results[0].ID
	if firstID != ids[0] && firstID != ids[2] {
		t.Errorf("first result ID=%d, expected an auth observation (id %d or %d)", firstID, ids[0], ids[2])
	}
}

func TestOrchestratorSearchWithTypeFilter(t *testing.T) {
	database := testDB(t)
	seedObservations(t, database)

	orch, err := NewOrchestrator(database)
	if err != nil {
		t.Fatalf("NewOrchestrator: %v", err)
	}

	if err := orch.RebuildIndex(); err != nil {
		t.Fatalf("RebuildIndex: %v", err)
	}

	results, err := orch.Search(SearchQuery{
		Text:  "authentication",
		Type:  "feature",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	// No auth observations have type=feature, so results should not include auth obs
	for _, r := range results {
		if r.ObsType != "feature" {
			t.Errorf("result ID=%d has type=%q, want 'feature'", r.ID, r.ObsType)
		}
	}
}

func TestOrchestratorSearchEmpty(t *testing.T) {
	database := testDB(t)

	orch, err := NewOrchestrator(database)
	if err != nil {
		t.Fatalf("NewOrchestrator: %v", err)
	}

	results, err := orch.Search(SearchQuery{Text: "anything", Limit: 5})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results from empty DB, got %d", len(results))
	}
}

func TestOrchestratorWeightConfig(t *testing.T) {
	cfg := DefaultWeights()
	if cfg.FTS < 0 || cfg.FTS > 1 {
		t.Errorf("FTS weight = %f, want [0, 1]", cfg.FTS)
	}
	if cfg.Vector < 0 || cfg.Vector > 1 {
		t.Errorf("Vector weight = %f, want [0, 1]", cfg.Vector)
	}
}

func TestFormatResult(t *testing.T) {
	r := &HybridResult{
		ID:      1,
		Title:   "Test",
		Score:   0.85,
		ObsType: "bugfix",
	}

	formatted := FormatResult(r)
	if formatted == "" {
		t.Error("FormatResult returned empty string")
	}
}
