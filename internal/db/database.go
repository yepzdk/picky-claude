// Package db provides SQLite persistence for observations, sessions, summaries,
// plans, and prompts.
package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps a SQLite database connection.
type DB struct {
	conn   *sql.DB
	logger *slog.Logger
}

// Open opens (or creates) the SQLite database at the given path and runs migrations.
func Open(path string, logger *slog.Logger) (*DB, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create db directory %s: %w", dir, err)
	}

	conn, err := sql.Open("sqlite", path+"?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("open database %s: %w", path, err)
	}

	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	db := &DB{conn: conn, logger: logger}
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return db, nil
}

// OpenInMemory creates an in-memory SQLite database. Useful for tests.
func OpenInMemory(logger *slog.Logger) (*DB, error) {
	conn, err := sql.Open("sqlite", ":memory:?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("open in-memory database: %w", err)
	}

	db := &DB{conn: conn, logger: logger}
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return db, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// Conn returns the underlying *sql.DB for direct queries.
func (db *DB) Conn() *sql.DB {
	return db.conn
}
