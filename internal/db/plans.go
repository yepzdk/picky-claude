package db

import (
	"database/sql"
	"fmt"
	"time"
)

// Plan represents a spec plan file tracked in the database.
type Plan struct {
	ID        int64
	Path      string
	SessionID string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// InsertPlan registers a plan file.
func (db *DB) InsertPlan(p *Plan) (int64, error) {
	res, err := db.conn.Exec(
		`INSERT INTO plans (path, session_id, status) VALUES (?, ?, ?)`,
		p.Path, p.SessionID, p.Status,
	)
	if err != nil {
		return 0, fmt.Errorf("insert plan: %w", err)
	}
	return res.LastInsertId()
}

// UpdatePlanStatus changes the status of a plan and updates the timestamp.
func (db *DB) UpdatePlanStatus(id int64, status string) error {
	_, err := db.conn.Exec(
		`UPDATE plans SET status = ?, updated_at = datetime('now') WHERE id = ?`,
		status, id,
	)
	if err != nil {
		return fmt.Errorf("update plan %d status: %w", id, err)
	}
	return nil
}

// GetPlanByPath finds a plan by its file path.
func (db *DB) GetPlanByPath(path string) (*Plan, error) {
	p := &Plan{}
	var createdAt, updatedAt string
	err := db.conn.QueryRow(
		`SELECT id, path, session_id, status, created_at, updated_at
		 FROM plans WHERE path = ? ORDER BY id DESC LIMIT 1`, path,
	).Scan(&p.ID, &p.Path, &p.SessionID, &p.Status, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get plan by path %s: %w", path, err)
	}
	p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	p.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
	return p, nil
}
