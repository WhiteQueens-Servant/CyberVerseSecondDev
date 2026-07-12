package scenic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

// SessionRecord represents a conversation session.
type SessionRecord struct {
	ID            string `json:"id"`
	CharacterID   string `json:"character_id"`
	CharacterName string `json:"character_name,omitempty"`
	StartedAt     string `json:"started_at"`
	EndedAt       string `json:"ended_at"`
	DurationS     int    `json:"duration_s"`
	TurnCount     int    `json:"turn_count"`
	Sentiment     string `json:"sentiment"`
}

// ConversationLog represents one message in a conversation.
type ConversationLog struct {
	ID        int    `json:"id"`
	SessionID string `json:"session_id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

// SessionLogDetail bundles a session's metadata with its conversation messages.
type SessionLogDetail struct {
	Session  SessionRecord      `json:"session"`
	Messages []ConversationLog  `json:"messages"`
}

// ListSessionsParams holds optional filters for ListSessions.
type ListSessionsParams struct {
	CharacterID string
	DateFrom    string
	DateTo      string
	Sentiment   string
}

// CreateSession inserts a new session record. Sentiment defaults to 'neutral'.
func (d *DB) CreateSession(ctx context.Context, characterID string) (SessionRecord, error) {
	id := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := d.conn.ExecContext(ctx,
		`INSERT INTO sessions (id, character_id, started_at, sentiment)
		 VALUES (?, ?, ?, 'neutral')`,
		id, characterID, now)
	if err != nil {
		return SessionRecord{}, err
	}
	return d.GetSession(ctx, id)
}

// EndSession updates the session's ended_at, duration_s, and turn_count.
func (d *DB) EndSession(ctx context.Context, id string) error {
	now := time.Now().UTC()
	// Get started_at to calculate duration
	var startedAt string
	err := d.conn.QueryRowContext(ctx,
		`SELECT started_at FROM sessions WHERE id = ?`, id).Scan(&startedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("session not found: %s", id)
		}
		return err
	}
	start, err := time.Parse(time.RFC3339, startedAt)
	if err != nil {
		return fmt.Errorf("parse started_at: %w", err)
	}
	durationS := int(now.Sub(start).Seconds())

	// Count conversation turns (user messages = turn count)
	var turnCount int
	err = d.conn.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM conversation_logs WHERE session_id = ? AND role = 'user'`,
		id).Scan(&turnCount)
	if err != nil {
		return err
	}

	_, err = d.conn.ExecContext(ctx,
		`UPDATE sessions SET ended_at = ?, duration_s = ?, turn_count = ? WHERE id = ?`,
		now.Format(time.RFC3339), durationS, turnCount, id)
	return err
}

// UpdateSentiment overwrites the sentiment field for a session.
// Called asynchronously after session end; best-effort.
func (d *DB) UpdateSentiment(ctx context.Context, sessionID, sentiment string) error {
	valid := map[string]bool{"positive": true, "neutral": true, "negative": true}
	if !valid[sentiment] {
		sentiment = "neutral"
	}
	_, err := d.conn.ExecContext(ctx,
		`UPDATE sessions SET sentiment = ? WHERE id = ?`,
		sentiment, sessionID)
	return err
}

// AppendLog inserts a conversation message into an existing session.
// For user messages, normalized_content is computed via normalizeQuestion for hot_questions aggregation.
// Original content is always preserved.
func (d *DB) AppendLog(ctx context.Context, sessionID, role, content string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	normalized := ""
	if role == "user" {
		normalized = normalizeQuestion(content)
	}
	_, err := d.conn.ExecContext(ctx,
		`INSERT INTO conversation_logs (session_id, role, content, normalized_content, timestamp)
		 VALUES (?, ?, ?, ?, ?)`,
		sessionID, role, content, normalized, now)
	return err
}

// normalizeQuestion preprocesses a user question for aggregation:
// strips punctuation, removes common filler words, collapses whitespace, lowercases.
// This is a rule-based approach (Plan B). LLM-based semantic grouping is deferred to T-24.
func normalizeQuestion(s string) string {
	// 1. Remove common Chinese/English question prefixes
	prefixes := []string{
		"请问", "你好请问", "你好，", "你好,", "你好 ",
		"能不能", "可以", "能不能帮我", "帮我看一下", "帮我查一下",
		"我想问", "我想知道", "我想了解", "请帮我",
		"could you", "can you", "please", "hey", "hi", "hello",
	}
	lower := strings.ToLower(strings.TrimSpace(s))
	for _, p := range prefixes {
		if strings.HasPrefix(lower, p) {
			lower = strings.TrimPrefix(lower, p)
			break
		}
	}

	// 2. Strip all punctuation (Chinese + English)
	var b strings.Builder
	for _, r := range lower {
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			continue
		}
		if !unicode.IsSpace(r) {
			b.WriteRune(r)
		} else {
			b.WriteRune(' ')
		}
	}
	result := b.String()

	// 3. Collapse multiple spaces
	for strings.Contains(result, "  ") {
		result = strings.ReplaceAll(result, "  ", " ")
	}
	result = strings.TrimSpace(result)

	// 4. Truncate to 100 chars (avoid very long unique keys)
	runes := []rune(result)
	if len(runes) > 100 {
		result = string(runes[:100])
	}
	return result
}

// ListSessions returns sessions matching the given filters, ordered by most recent first.
func (d *DB) ListSessions(ctx context.Context, params ListSessionsParams) ([]SessionRecord, error) {
	query := `SELECT id, character_id, started_at, ended_at, duration_s, turn_count, sentiment
		      FROM sessions WHERE 1=1`
	var args []any
	if params.CharacterID != "" {
		query += ` AND character_id = ?`
		args = append(args, params.CharacterID)
	}
	if params.DateFrom != "" {
		query += ` AND started_at >= ?`
		args = append(args, params.DateFrom)
	}
	if params.DateTo != "" {
		query += ` AND started_at <= ?`
		args = append(args, params.DateTo+"T23:59:59Z")
	}
	if params.Sentiment != "" {
		query += ` AND sentiment = ?`
		args = append(args, params.Sentiment)
	}
	query += ` ORDER BY started_at DESC`

	rows, err := d.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []SessionRecord
	for rows.Next() {
		var s SessionRecord
		if err := rows.Scan(&s.ID, &s.CharacterID, &s.StartedAt, &s.EndedAt,
			&s.DurationS, &s.TurnCount, &s.Sentiment); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	if result == nil {
		result = []SessionRecord{}
	}
	return result, rows.Err()
}

// GetSession returns a single session by ID.
func (d *DB) GetSession(ctx context.Context, id string) (SessionRecord, error) {
	row := d.conn.QueryRowContext(ctx,
		`SELECT id, character_id, started_at, ended_at, duration_s, turn_count, sentiment
		 FROM sessions WHERE id = ?`, id)
	var s SessionRecord
	if err := row.Scan(&s.ID, &s.CharacterID, &s.StartedAt, &s.EndedAt,
		&s.DurationS, &s.TurnCount, &s.Sentiment); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SessionRecord{}, fmt.Errorf("session not found: %s", id)
		}
		return SessionRecord{}, err
	}
	return s, nil
}

// GetSessionDetail returns a session with all its conversation messages.
func (d *DB) GetSessionDetail(ctx context.Context, id string) (SessionLogDetail, error) {
	session, err := d.GetSession(ctx, id)
	if err != nil {
		return SessionLogDetail{}, err
	}
	rows, err := d.conn.QueryContext(ctx,
		`SELECT id, session_id, role, content, timestamp
		 FROM conversation_logs WHERE session_id = ? ORDER BY timestamp`, id)
	if err != nil {
		return SessionLogDetail{}, err
	}
	defer rows.Close()

	var messages []ConversationLog
	for rows.Next() {
		var m ConversationLog
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &m.Timestamp); err != nil {
			return SessionLogDetail{}, err
		}
		messages = append(messages, m)
	}
	if messages == nil {
		messages = []ConversationLog{}
	}
	return SessionLogDetail{Session: session, Messages: messages}, rows.Err()
}
