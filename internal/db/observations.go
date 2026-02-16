package db

import (
	"database/sql"
	"fmt"
	"time"
)

// Observation represents a single discovery, change, or decision.
type Observation struct {
	ID        int64
	SessionID string
	Type      string
	Title     string
	Text      string
	Project   string
	Metadata  string
	CreatedAt time.Time
}

// InsertObservation stores a new observation and returns its ID.
func (db *DB) InsertObservation(o *Observation) (int64, error) {
	res, err := db.conn.Exec(
		`INSERT INTO observations (session_id, type, title, text, project, metadata)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		o.SessionID, o.Type, o.Title, o.Text, o.Project, o.Metadata,
	)
	if err != nil {
		return 0, fmt.Errorf("insert observation: %w", err)
	}
	return res.LastInsertId()
}

// GetObservation retrieves an observation by ID.
func (db *DB) GetObservation(id int64) (*Observation, error) {
	o := &Observation{}
	var createdAt string
	err := db.conn.QueryRow(
		`SELECT id, session_id, type, title, text, project, metadata, created_at
		 FROM observations WHERE id = ?`, id,
	).Scan(&o.ID, &o.SessionID, &o.Type, &o.Title, &o.Text, &o.Project, &o.Metadata, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get observation %d: %w", id, err)
	}
	o.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	return o, nil
}

// GetObservations retrieves multiple observations by their IDs.
func (db *DB) GetObservations(ids []int64) ([]*Observation, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	query := "SELECT id, session_id, type, title, text, project, metadata, created_at FROM observations WHERE id IN ("
	args := make([]any, len(ids))
	for i, id := range ids {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i] = id
	}
	query += ") ORDER BY id"

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("get observations: %w", err)
	}
	defer rows.Close()

	var results []*Observation
	for rows.Next() {
		o := &Observation{}
		var createdAt string
		if err := rows.Scan(&o.ID, &o.SessionID, &o.Type, &o.Title, &o.Text, &o.Project, &o.Metadata, &createdAt); err != nil {
			return nil, fmt.Errorf("scan observation: %w", err)
		}
		o.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		results = append(results, o)
	}
	return results, rows.Err()
}

// SearchObservations performs full-text search against the observations FTS index.
func (db *DB) SearchObservations(query string, limit int) ([]*Observation, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := db.conn.Query(
		`SELECT o.id, o.session_id, o.type, o.title, o.text, o.project, o.metadata, o.created_at
		 FROM observations o
		 JOIN observations_fts fts ON o.id = fts.rowid
		 WHERE observations_fts MATCH ?
		 ORDER BY fts.rank
		 LIMIT ?`,
		query, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("search observations: %w", err)
	}
	defer rows.Close()

	var results []*Observation
	for rows.Next() {
		o := &Observation{}
		var createdAt string
		if err := rows.Scan(&o.ID, &o.SessionID, &o.Type, &o.Title, &o.Text, &o.Project, &o.Metadata, &createdAt); err != nil {
			return nil, fmt.Errorf("scan observation: %w", err)
		}
		o.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		results = append(results, o)
	}
	return results, rows.Err()
}

// RecentObservations returns the N most recent observations, optionally filtered
// by project. Results are ordered most recent first.
func (db *DB) RecentObservations(project string, limit int) ([]*Observation, error) {
	if limit <= 0 {
		limit = 50
	}

	var query string
	var args []any

	if project != "" {
		query = `SELECT id, session_id, type, title, text, project, metadata, created_at
			 FROM observations WHERE project = ? ORDER BY created_at DESC LIMIT ?`
		args = []any{project, limit}
	} else {
		query = `SELECT id, session_id, type, title, text, project, metadata, created_at
			 FROM observations ORDER BY created_at DESC LIMIT ?`
		args = []any{limit}
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("recent observations: %w", err)
	}
	defer rows.Close()

	var results []*Observation
	for rows.Next() {
		o := &Observation{}
		var createdAt string
		if err := rows.Scan(&o.ID, &o.SessionID, &o.Type, &o.Title, &o.Text, &o.Project, &o.Metadata, &createdAt); err != nil {
			return nil, fmt.Errorf("scan observation: %w", err)
		}
		o.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		results = append(results, o)
	}
	return results, rows.Err()
}

// SearchFilter defines parameters for filtered full-text search.
type SearchFilter struct {
	Query     string
	Type      string
	Project   string
	DateStart string // YYYY-MM-DD
	DateEnd   string // YYYY-MM-DD
	Limit     int
}

// FilteredSearch performs full-text search with optional type, project, and date filters.
func (db *DB) FilteredSearch(f SearchFilter) ([]*Observation, error) {
	if f.Limit <= 0 {
		f.Limit = 20
	}

	query := `SELECT o.id, o.session_id, o.type, o.title, o.text, o.project, o.metadata, o.created_at
		 FROM observations o
		 JOIN observations_fts fts ON o.id = fts.rowid
		 WHERE observations_fts MATCH ?`
	args := []any{f.Query}

	if f.Type != "" {
		query += " AND o.type = ?"
		args = append(args, f.Type)
	}
	if f.Project != "" {
		query += " AND o.project = ?"
		args = append(args, f.Project)
	}
	if f.DateStart != "" {
		query += " AND o.created_at >= ?"
		args = append(args, f.DateStart)
	}
	if f.DateEnd != "" {
		query += " AND o.created_at <= ?"
		args = append(args, f.DateEnd)
	}

	query += " ORDER BY fts.rank LIMIT ?"
	args = append(args, f.Limit)

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("filtered search: %w", err)
	}
	defer rows.Close()

	var results []*Observation
	for rows.Next() {
		o := &Observation{}
		var createdAt string
		if err := rows.Scan(&o.ID, &o.SessionID, &o.Type, &o.Title, &o.Text, &o.Project, &o.Metadata, &createdAt); err != nil {
			return nil, fmt.Errorf("scan observation: %w", err)
		}
		o.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		results = append(results, o)
	}
	return results, rows.Err()
}

// TimelineAround returns observations around a given observation ID.
func (db *DB) TimelineAround(anchorID int64, before, after int) ([]*Observation, error) {
	if before <= 0 {
		before = 5
	}
	if after <= 0 {
		after = 5
	}

	rows, err := db.conn.Query(
		`SELECT id, session_id, type, title, text, project, metadata, created_at
		 FROM observations
		 WHERE id >= (SELECT id FROM observations WHERE id <= ? ORDER BY id DESC LIMIT 1 OFFSET ?)
		   AND id <= (SELECT id FROM observations WHERE id >= ? ORDER BY id ASC LIMIT 1 OFFSET ?)
		 ORDER BY id`,
		anchorID, before, anchorID, after,
	)
	if err != nil {
		return nil, fmt.Errorf("timeline around %d: %w", anchorID, err)
	}
	defer rows.Close()

	var results []*Observation
	for rows.Next() {
		o := &Observation{}
		var createdAt string
		if err := rows.Scan(&o.ID, &o.SessionID, &o.Type, &o.Title, &o.Text, &o.Project, &o.Metadata, &createdAt); err != nil {
			return nil, fmt.Errorf("scan observation: %w", err)
		}
		o.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		results = append(results, o)
	}
	return results, rows.Err()
}
