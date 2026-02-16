package db

import (
	"fmt"
	"time"
)

// Prompt represents a stored prompt for context injection.
type Prompt struct {
	ID        int64
	SessionID string
	Role      string
	Text      string
	CreatedAt time.Time
}

// InsertPrompt stores a prompt.
func (db *DB) InsertPrompt(p *Prompt) (int64, error) {
	res, err := db.conn.Exec(
		`INSERT INTO prompts (session_id, role, text) VALUES (?, ?, ?)`,
		p.SessionID, p.Role, p.Text,
	)
	if err != nil {
		return 0, fmt.Errorf("insert prompt: %w", err)
	}
	return res.LastInsertId()
}

// RecentPrompts returns the N most recent prompts for a session.
func (db *DB) RecentPrompts(sessionID string, limit int) ([]*Prompt, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := db.conn.Query(
		`SELECT id, session_id, role, text, created_at
		 FROM prompts WHERE session_id = ?
		 ORDER BY created_at DESC LIMIT ?`,
		sessionID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("recent prompts for %s: %w", sessionID, err)
	}
	defer rows.Close()

	var results []*Prompt
	for rows.Next() {
		p := &Prompt{}
		var createdAt string
		if err := rows.Scan(&p.ID, &p.SessionID, &p.Role, &p.Text, &createdAt); err != nil {
			return nil, fmt.Errorf("scan prompt: %w", err)
		}
		p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		results = append(results, p)
	}
	return results, rows.Err()
}
