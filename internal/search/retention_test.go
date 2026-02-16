package search

import (
	"testing"
	"time"

	"github.com/jesperpedersen/picky-claude/internal/db"
)

func TestDeleteOldObservations(t *testing.T) {
	database := testDB(t)

	// Insert observations — the default created_at is "now"
	for i := 0; i < 3; i++ {
		database.InsertObservation(&db.Observation{
			SessionID: "s1", Title: "recent", Text: "text",
		})
	}

	// Insert old observations by setting created_at directly
	for i := 0; i < 2; i++ {
		database.Conn().Exec(
			`INSERT INTO observations (session_id, title, text, created_at)
			 VALUES (?, ?, ?, datetime('now', '-100 days'))`,
			"s1", "old", "old text",
		)
	}

	ret := NewRetention(database)
	deleted, err := ret.DeleteOldObservations(90)
	if err != nil {
		t.Fatalf("DeleteOldObservations: %v", err)
	}
	if deleted != 2 {
		t.Errorf("deleted = %d, want 2", deleted)
	}

	// Verify only recent ones remain
	obs, err := database.RecentObservations("", 100)
	if err != nil {
		t.Fatalf("RecentObservations: %v", err)
	}
	if len(obs) != 3 {
		t.Errorf("remaining observations = %d, want 3", len(obs))
	}
}

func TestDeleteOldObservationsNothingToDelete(t *testing.T) {
	database := testDB(t)

	database.InsertObservation(&db.Observation{
		SessionID: "s1", Title: "recent", Text: "text",
	})

	ret := NewRetention(database)
	deleted, err := ret.DeleteOldObservations(90)
	if err != nil {
		t.Fatalf("DeleteOldObservations: %v", err)
	}
	if deleted != 0 {
		t.Errorf("deleted = %d, want 0", deleted)
	}
}

func TestCleanupStaleSessions(t *testing.T) {
	database := testDB(t)

	database.InsertSession(&db.Session{ID: "s1", Project: "test"})

	ret := NewRetention(database)
	cleaned, err := ret.CleanupStaleSessions(24)
	if err != nil {
		t.Fatalf("CleanupStaleSessions: %v", err)
	}
	// The session was just created, so it shouldn't be stale
	if cleaned != 0 {
		t.Errorf("cleaned = %d, want 0", cleaned)
	}
}

func TestVacuum(t *testing.T) {
	database := testDB(t)

	ret := NewRetention(database)
	err := ret.Vacuum()
	if err != nil {
		t.Fatalf("Vacuum: %v", err)
	}
}

func TestRetentionConfig(t *testing.T) {
	cfg := DefaultRetentionConfig()
	if cfg.MaxAgeDays <= 0 {
		t.Error("MaxAgeDays should be positive")
	}
	if cfg.Interval <= 0 {
		t.Error("Interval should be positive")
	}
	if cfg.StaleSessionHours <= 0 {
		t.Error("StaleSessionHours should be positive")
	}
}

func TestSchedulerStartStop(t *testing.T) {
	database := testDB(t)
	ret := NewRetention(database)

	cfg := RetentionConfig{
		MaxAgeDays:        90,
		StaleSessionHours: 24,
		Interval:          50 * time.Millisecond,
	}

	stop := ret.StartScheduler(cfg)

	// Let it tick at least once
	time.Sleep(100 * time.Millisecond)

	stop()
}
