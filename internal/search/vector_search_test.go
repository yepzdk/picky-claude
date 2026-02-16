package search

import (
	"log/slog"
	"os"
	"testing"

	"github.com/jesperpedersen/picky-claude/internal/db"
)

func testDB(t *testing.T) *db.DB {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	database, err := db.OpenInMemory(logger)
	if err != nil {
		t.Fatalf("OpenInMemory: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return database
}

func TestVectorStoreAndSearch(t *testing.T) {
	database := testDB(t)

	store, err := NewVectorStore(database)
	if err != nil {
		t.Fatalf("NewVectorStore: %v", err)
	}

	// Insert some observations
	obs := []struct {
		title string
		text  string
	}{
		{"auth bug", "Fixed authentication login flow error"},
		{"db migration", "Database migration schema updated for users table"},
		{"auth token", "Authentication session token expiration fixed"},
		{"css fix", "Fixed CSS flexbox layout issue in dashboard"},
	}

	var ids []int64
	for _, o := range obs {
		id, err := database.InsertObservation(&db.Observation{
			SessionID: "s1",
			Type:      "bugfix",
			Title:     o.title,
			Text:      o.text,
			Project:   "test",
		})
		if err != nil {
			t.Fatalf("InsertObservation: %v", err)
		}
		ids = append(ids, id)
	}

	// Index all observations
	if err := store.IndexAll(); err != nil {
		t.Fatalf("IndexAll: %v", err)
	}

	// Search for authentication-related observations
	results, err := store.Search("authentication login session", 3)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Search returned no results")
	}

	// First result should be an auth-related observation
	first := results[0]
	if first.ID != ids[0] && first.ID != ids[2] {
		t.Errorf("first result ID=%d, expected auth observation (id %d or %d)", first.ID, ids[0], ids[2])
	}
}

func TestVectorStoreIndexSingle(t *testing.T) {
	database := testDB(t)

	store, err := NewVectorStore(database)
	if err != nil {
		t.Fatalf("NewVectorStore: %v", err)
	}

	id, err := database.InsertObservation(&db.Observation{
		SessionID: "s1",
		Title:     "test",
		Text:      "sample observation text",
	})
	if err != nil {
		t.Fatalf("InsertObservation: %v", err)
	}

	if err := store.IndexObservation(id); err != nil {
		t.Fatalf("IndexObservation: %v", err)
	}

	results, err := store.Search("sample observation", 5)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ID != id {
		t.Errorf("result ID = %d, want %d", results[0].ID, id)
	}
}

func TestVectorStoreEmptySearch(t *testing.T) {
	database := testDB(t)

	store, err := NewVectorStore(database)
	if err != nil {
		t.Fatalf("NewVectorStore: %v", err)
	}

	results, err := store.Search("anything", 5)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results from empty store, got %d", len(results))
	}
}
