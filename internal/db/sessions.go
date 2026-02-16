package db

import (
	"database/sql"
	"fmt"
	"time"
)

// Session represents a Claude Code session.
type Session struct {
	ID           string
	Project      string
	StartedAt    time.Time
	EndedAt      *time.Time
	MessageCount int
	Metadata     string
}

// InsertSession creates a new session record.
func (db *DB) InsertSession(s *Session) error {
	_, err := db.conn.Exec(
		`INSERT INTO sessions (id, project, metadata) VALUES (?, ?, ?)`,
		s.ID, s.Project, s.Metadata,
	)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

// GetSession retrieves a session by ID.
func (db *DB) GetSession(id string) (*Session, error) {
	s := &Session{}
	var startedAt string
	var endedAt sql.NullString
	err := db.conn.QueryRow(
		`SELECT id, project, started_at, ended_at, message_count, metadata
		 FROM sessions WHERE id = ?`, id,
	).Scan(&s.ID, &s.Project, &startedAt, &endedAt, &s.MessageCount, &s.Metadata)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get session %s: %w", id, err)
	}
	s.StartedAt, _ = time.Parse("2006-01-02 15:04:05", startedAt)
	if endedAt.Valid {
		t, _ := time.Parse("2006-01-02 15:04:05", endedAt.String)
		s.EndedAt = &t
	}
	return s, nil
}

// EndSession marks a session as ended with the current timestamp.
func (db *DB) EndSession(id string) error {
	_, err := db.conn.Exec(
		`UPDATE sessions SET ended_at = datetime('now') WHERE id = ?`, id,
	)
	if err != nil {
		return fmt.Errorf("end session %s: %w", id, err)
	}
	return nil
}

// IncrementMessageCount bumps the message counter for a session.
func (db *DB) IncrementMessageCount(id string) error {
	_, err := db.conn.Exec(
		`UPDATE sessions SET message_count = message_count + 1 WHERE id = ?`, id,
	)
	if err != nil {
		return fmt.Errorf("increment message count for %s: %w", id, err)
	}
	return nil
}

// CleanupStaleSessions ends sessions that have been active longer than maxAgeHours.
// Returns the number of sessions cleaned up.
func (db *DB) CleanupStaleSessions(maxAgeHours int) (int, error) {
	res, err := db.conn.Exec(
		`UPDATE sessions SET ended_at = datetime('now')
		 WHERE ended_at IS NULL
		   AND started_at < datetime('now', ? || ' hours')`,
		fmt.Sprintf("-%d", maxAgeHours),
	)
	if err != nil {
		return 0, fmt.Errorf("cleanup stale sessions: %w", err)
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

// ListActiveSessions returns sessions that have not ended.
func (db *DB) ListActiveSessions() ([]*Session, error) {
	rows, err := db.conn.Query(
		`SELECT id, project, started_at, ended_at, message_count, metadata
		 FROM sessions WHERE ended_at IS NULL ORDER BY started_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list active sessions: %w", err)
	}
	defer rows.Close()

	var results []*Session
	for rows.Next() {
		s := &Session{}
		var startedAt string
		var endedAt sql.NullString
		if err := rows.Scan(&s.ID, &s.Project, &startedAt, &endedAt, &s.MessageCount, &s.Metadata); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		s.StartedAt, _ = time.Parse("2006-01-02 15:04:05", startedAt)
		results = append(results, s)
	}
	return results, rows.Err()
}
