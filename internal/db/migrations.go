package db

import "fmt"

// migrations is an ordered list of SQL statements. Each entry runs once.
// New migrations are appended at the end; never modify existing entries.
var migrations = []string{
	// 0: schema versioning table
	`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at TEXT NOT NULL DEFAULT (datetime('now'))
	)`,

	// 1: observations — individual discoveries, changes, decisions
	`CREATE TABLE IF NOT EXISTS observations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		type TEXT NOT NULL DEFAULT 'discovery',
		title TEXT NOT NULL DEFAULT '',
		text TEXT NOT NULL,
		project TEXT NOT NULL DEFAULT '',
		metadata TEXT NOT NULL DEFAULT '{}',
		created_at TEXT NOT NULL DEFAULT (datetime('now'))
	)`,

	// 2: FTS5 index for observations
	`CREATE VIRTUAL TABLE IF NOT EXISTS observations_fts USING fts5(
		title,
		text,
		content=observations,
		content_rowid=id,
		tokenize='porter unicode61'
	)`,

	// 3: triggers to keep FTS in sync
	`CREATE TRIGGER IF NOT EXISTS observations_ai AFTER INSERT ON observations BEGIN
		INSERT INTO observations_fts(rowid, title, text) VALUES (new.id, new.title, new.text);
	END`,

	`CREATE TRIGGER IF NOT EXISTS observations_ad AFTER DELETE ON observations BEGIN
		INSERT INTO observations_fts(observations_fts, rowid, title, text) VALUES('delete', old.id, old.title, old.text);
	END`,

	`CREATE TRIGGER IF NOT EXISTS observations_au AFTER UPDATE ON observations BEGIN
		INSERT INTO observations_fts(observations_fts, rowid, title, text) VALUES('delete', old.id, old.title, old.text);
		INSERT INTO observations_fts(rowid, title, text) VALUES (new.id, new.title, new.text);
	END`,

	// 6: sessions — track each Claude Code session
	`CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		project TEXT NOT NULL DEFAULT '',
		started_at TEXT NOT NULL DEFAULT (datetime('now')),
		ended_at TEXT,
		message_count INTEGER NOT NULL DEFAULT 0,
		metadata TEXT NOT NULL DEFAULT '{}'
	)`,

	// 7: summaries — session-end summaries
	`CREATE TABLE IF NOT EXISTS summaries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		text TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT (datetime('now')),
		FOREIGN KEY (session_id) REFERENCES sessions(id)
	)`,

	// 8: plans — plan file metadata and status tracking
	`CREATE TABLE IF NOT EXISTS plans (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		path TEXT NOT NULL,
		session_id TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL DEFAULT 'PENDING',
		created_at TEXT NOT NULL DEFAULT (datetime('now')),
		updated_at TEXT NOT NULL DEFAULT (datetime('now'))
	)`,

	// 9: prompts — stored prompts for context injection
	`CREATE TABLE IF NOT EXISTS prompts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL DEFAULT '',
		role TEXT NOT NULL DEFAULT 'system',
		text TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT (datetime('now'))
	)`,

	// 10: observation embeddings for vector search
	`CREATE TABLE IF NOT EXISTS observation_embeddings (
		observation_id INTEGER PRIMARY KEY,
		embedding BLOB NOT NULL,
		FOREIGN KEY (observation_id) REFERENCES observations(id) ON DELETE CASCADE
	)`,

	// 11: indexes
	`CREATE INDEX IF NOT EXISTS idx_observations_session ON observations(session_id)`,
	`CREATE INDEX IF NOT EXISTS idx_observations_type ON observations(type)`,
	`CREATE INDEX IF NOT EXISTS idx_observations_project ON observations(project)`,
	`CREATE INDEX IF NOT EXISTS idx_observations_created ON observations(created_at)`,
	`CREATE INDEX IF NOT EXISTS idx_summaries_session ON summaries(session_id)`,
	`CREATE INDEX IF NOT EXISTS idx_plans_session ON plans(session_id)`,
	`CREATE INDEX IF NOT EXISTS idx_plans_status ON plans(status)`,
}

// migrate runs all pending migrations in order.
func (db *DB) migrate() error {
	// Ensure the migrations table exists (migration 0).
	if _, err := db.conn.Exec(migrations[0]); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	var current int
	row := db.conn.QueryRow("SELECT COALESCE(MAX(version), -1) FROM schema_migrations")
	if err := row.Scan(&current); err != nil {
		return fmt.Errorf("read migration version: %w", err)
	}

	for i := current + 1; i < len(migrations); i++ {
		db.logger.Debug("running migration", "version", i)
		if _, err := db.conn.Exec(migrations[i]); err != nil {
			return fmt.Errorf("migration %d: %w", i, err)
		}
		if _, err := db.conn.Exec("INSERT INTO schema_migrations (version) VALUES (?)", i); err != nil {
			return fmt.Errorf("record migration %d: %w", i, err)
		}
	}

	return nil
}
