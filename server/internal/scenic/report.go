package scenic

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Report represents a sentiment analysis report.
type Report struct {
	ID          string                 `json:"id"`
	Status      string                 `json:"status"`
	DateFrom    string                 `json:"date_from"`
	DateTo      string                 `json:"date_to"`
	CharacterID string                 `json:"character_id"`
	Content     map[string]interface{} `json:"content,omitempty"`
	CreatedAt   string                 `json:"created_at"`
	FinishedAt  string                 `json:"finished_at"`
}

// CreateReport inserts a new report in 'queued' status.
func (d *DB) CreateReport(ctx context.Context, dateFrom, dateTo, characterID string) (Report, error) {
	id := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := d.conn.ExecContext(ctx,
		`INSERT INTO reports (id, status, date_from, date_to, character_id, content, created_at)
		 VALUES (?, 'queued', ?, ?, ?, '{}', ?)`,
		id, dateFrom, dateTo, characterID, now)
	if err != nil {
		return Report{}, err
	}
	return d.GetReport(ctx, id)
}

// ListReports returns all reports ordered by creation time (newest first).
func (d *DB) ListReports(ctx context.Context) ([]Report, error) {
	rows, err := d.conn.QueryContext(ctx,
		`SELECT id, status, date_from, date_to, character_id, content, created_at, finished_at
		 FROM reports ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Report
	for rows.Next() {
		r, err := scanReport(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	if result == nil {
		result = []Report{}
	}
	return result, rows.Err()
}

// GetReport returns a single report by ID.
func (d *DB) GetReport(ctx context.Context, id string) (Report, error) {
	row := d.conn.QueryRowContext(ctx,
		`SELECT id, status, date_from, date_to, character_id, content, created_at, finished_at
		 FROM reports WHERE id = ?`, id)
	r, err := scanReportRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Report{}, fmt.Errorf("report not found: %s", id)
		}
		return Report{}, err
	}
	return r, nil
}

// DeleteReport removes a report.
func (d *DB) DeleteReport(ctx context.Context, id string) error {
	res, err := d.conn.ExecContext(ctx, `DELETE FROM reports WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("report not found: %s", id)
	}
	return nil
}

// UpdateReportContent updates a report's content and sets status to 'completed'.
func (d *DB) UpdateReportContent(ctx context.Context, id string, content map[string]interface{}) error {
	contentJSON, err := json.Marshal(content)
	if err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := d.conn.ExecContext(ctx,
		`UPDATE reports SET status = 'completed', content = ?, finished_at = ? WHERE id = ?`,
		string(contentJSON), now, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("report not found: %s", id)
	}
	return nil
}

// UpdateReportStatus updates a report's status.
func (d *DB) UpdateReportStatus(ctx context.Context, id, status string) error {
	_, err := d.conn.ExecContext(ctx,
		`UPDATE reports SET status = ? WHERE id = ?`, status, id)
	return err
}

// ── helpers ──

func scanReport(sc scannable) (Report, error) {
	var r Report
	var contentJSON string
	var finishedAt string
	err := sc.Scan(&r.ID, &r.Status, &r.DateFrom, &r.DateTo, &r.CharacterID,
		&contentJSON, &r.CreatedAt, &finishedAt)
	if err != nil {
		return Report{}, err
	}
	r.FinishedAt = finishedAt
	if contentJSON != "" && contentJSON != "{}" {
		json.Unmarshal([]byte(contentJSON), &r.Content)
	}
	return r, nil
}

func scanReportRow(row *sql.Row) (Report, error) {
	return scanReport(row)
}
