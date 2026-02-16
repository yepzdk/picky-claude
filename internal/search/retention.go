package search

import (
	"fmt"
	"time"

	"github.com/jesperpedersen/picky-claude/internal/db"
)

// RetentionConfig controls the retention scheduler behavior.
type RetentionConfig struct {
	MaxAgeDays        int           // Delete observations older than this
	StaleSessionHours int           // End sessions older than this
	Interval          time.Duration // How often to run cleanup
}

// DefaultRetentionConfig returns the default retention settings.
func DefaultRetentionConfig() RetentionConfig {
	return RetentionConfig{
		MaxAgeDays:        90,
		StaleSessionHours: 24,
		Interval:          6 * time.Hour,
	}
}

// Retention handles database cleanup and maintenance.
type Retention struct {
	db *db.DB
}

// NewRetention creates a retention manager.
func NewRetention(database *db.DB) *Retention {
	return &Retention{db: database}
}

// DeleteOldObservations removes observations older than maxAgeDays.
// Returns the number of observations deleted.
func (r *Retention) DeleteOldObservations(maxAgeDays int) (int, error) {
	res, err := r.db.Conn().Exec(
		`DELETE FROM observations WHERE created_at < datetime('now', ? || ' days')`,
		fmt.Sprintf("-%d", maxAgeDays),
	)
	if err != nil {
		return 0, fmt.Errorf("delete old observations: %w", err)
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

// CleanupStaleSessions ends sessions that have been active too long.
func (r *Retention) CleanupStaleSessions(maxAgeHours int) (int, error) {
	return r.db.CleanupStaleSessions(maxAgeHours)
}

// Vacuum runs VACUUM on the database to reclaim space.
func (r *Retention) Vacuum() error {
	_, err := r.db.Conn().Exec("VACUUM")
	if err != nil {
		return fmt.Errorf("vacuum: %w", err)
	}
	return nil
}

// RunOnce performs a single retention cycle: delete old data, cleanup sessions,
// and vacuum.
func (r *Retention) RunOnce(cfg RetentionConfig) error {
	if _, err := r.DeleteOldObservations(cfg.MaxAgeDays); err != nil {
		return err
	}
	if _, err := r.CleanupStaleSessions(cfg.StaleSessionHours); err != nil {
		return err
	}
	return r.Vacuum()
}

// StartScheduler starts a background goroutine that periodically runs
// retention cleanup. Returns a stop function.
func (r *Retention) StartScheduler(cfg RetentionConfig) func() {
	done := make(chan struct{})

	go func() {
		ticker := time.NewTicker(cfg.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				r.RunOnce(cfg)
			case <-done:
				return
			}
		}
	}()

	return func() {
		close(done)
	}
}
