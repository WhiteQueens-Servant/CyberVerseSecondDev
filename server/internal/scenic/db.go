package scenic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps the shared cyberverse.db connection used by the scenic guide module.
type DB struct {
	conn *sql.DB
}

// OpenDB opens (or creates) the unified SQLite database and runs idempotent migrations.
func OpenDB(dbPath string) (*DB, error) {
	if dbPath == "" {
		return nil, errors.New("database path is required")
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	conn.SetMaxOpenConns(1)
	d := &DB{conn: conn}
	if err := d.migrate(context.Background()); err != nil {
		conn.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return d, nil
}

// Close closes the underlying database connection.
func (d *DB) Close() error {
	if d == nil || d.conn == nil {
		return nil
	}
	return d.conn.Close()
}

// Conn returns the raw *sql.DB for direct queries if needed.
func (d *DB) Conn() *sql.DB {
	if d == nil {
		return nil
	}
	return d.conn
}

// migrate creates all tables and indexes idempotently.
func (d *DB) migrate(ctx context.Context) error {
	if _, err := d.conn.ExecContext(ctx, `PRAGMA journal_mode=WAL;`); err != nil {
		return err
	}
	stmts := []string{
		// Attractions
		`CREATE TABLE IF NOT EXISTS attractions (
			id            TEXT PRIMARY KEY,
			name          TEXT NOT NULL,
			description   TEXT NOT NULL DEFAULT '',
			category      TEXT NOT NULL DEFAULT '',
			location      TEXT NOT NULL DEFAULT '',
			image_url     TEXT NOT NULL DEFAULT '',
			tags          TEXT NOT NULL DEFAULT '[]',
			opening_hours TEXT NOT NULL DEFAULT '',
			ticket_info   TEXT NOT NULL DEFAULT '',
			created_at    TEXT NOT NULL,
			updated_at    TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_attractions_category ON attractions(category)`,

		// Routes
		`CREATE TABLE IF NOT EXISTS routes (
			id          TEXT PRIMARY KEY,
			name        TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			duration    TEXT NOT NULL DEFAULT '',
			difficulty  TEXT NOT NULL DEFAULT 'easy',
			tags        TEXT NOT NULL DEFAULT '[]',
			created_at  TEXT NOT NULL,
			updated_at  TEXT NOT NULL
		)`,

		// Route steps (junction table)
		`CREATE TABLE IF NOT EXISTS route_steps (
			id               INTEGER PRIMARY KEY AUTOINCREMENT,
			route_id         TEXT NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
			attraction_id    TEXT NOT NULL REFERENCES attractions(id) ON DELETE CASCADE,
			step_order       INTEGER NOT NULL DEFAULT 0,
			duration_minutes INTEGER NOT NULL DEFAULT 30,
			highlight        TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE INDEX IF NOT EXISTS idx_route_steps_route ON route_steps(route_id)`,
		`CREATE INDEX IF NOT EXISTS idx_route_steps_attraction ON route_steps(attraction_id)`,

		// Sessions
		// sentiment defaults to 'neutral' so that ListSessions filtering works
		// before T-24 (LLM sentiment analysis) is implemented. T-24 will overwrite
		// this value with the LLM-determined label when the session ends.
		`CREATE TABLE IF NOT EXISTS sessions (
			id            TEXT PRIMARY KEY,
			character_id  TEXT NOT NULL,
			started_at    TEXT NOT NULL,
			ended_at      TEXT NOT NULL DEFAULT '',
			duration_s    INTEGER NOT NULL DEFAULT 0,
			turn_count    INTEGER NOT NULL DEFAULT 0,
			sentiment     TEXT NOT NULL DEFAULT 'neutral'
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_character ON sessions(character_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_started ON sessions(started_at)`,

		// Conversation logs
		// normalized_content: rule-preprocessed version of user messages for hot_questions aggregation.
		// Original content is preserved for conversation display. Only user messages are normalized;
		// assistant messages keep normalized_content = ''.
		`CREATE TABLE IF NOT EXISTS conversation_logs (
			id                 INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id         TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
			role               TEXT NOT NULL,
			content            TEXT NOT NULL DEFAULT '',
			normalized_content TEXT NOT NULL DEFAULT '',
			timestamp          TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_conv_logs_session ON conversation_logs(session_id)`,

		// Reports
		`CREATE TABLE IF NOT EXISTS reports (
			id           TEXT PRIMARY KEY,
			status       TEXT NOT NULL DEFAULT 'queued',
			date_from    TEXT NOT NULL DEFAULT '',
			date_to      TEXT NOT NULL DEFAULT '',
			character_id TEXT NOT NULL DEFAULT '',
			content      TEXT NOT NULL DEFAULT '{}',
			created_at   TEXT NOT NULL,
			finished_at  TEXT NOT NULL DEFAULT ''
		)`,
	}
	for _, stmt := range stmts {
		if _, err := d.conn.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("exec %q: %w", truncateSQL(stmt), err)
		}
	}
	return nil
}

func truncateSQL(s string) string {
	if len(s) > 60 {
		return s[:60] + "..."
	}
	return s
}
