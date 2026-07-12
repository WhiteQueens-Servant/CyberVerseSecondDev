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

var ErrAttractionInUse = errors.New("attraction is referenced by one or more routes")

// Attraction represents a scenic spot.
type Attraction struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Category      string   `json:"category"`
	Location      string   `json:"location"`
	ImageURL      string   `json:"image_url"`
	Tags          []string `json:"tags"`
	OpeningHours  string   `json:"opening_hours"`
	TicketInfo    string   `json:"ticket_info"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

// AttractionForm is the input for creating or updating an attraction.
type AttractionForm struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Category     string   `json:"category"`
	Location     string   `json:"location"`
	ImageURL     string   `json:"image_url"`
	Tags         []string `json:"tags"`
	OpeningHours string   `json:"opening_hours"`
	TicketInfo   string   `json:"ticket_info"`
}

// ListAttractions returns all attractions ordered by creation time.
func (d *DB) ListAttractions(ctx context.Context) ([]Attraction, error) {
	rows, err := d.conn.QueryContext(ctx,
		`SELECT id, name, description, category, location, image_url, tags, opening_hours, ticket_info, created_at, updated_at
		 FROM attractions ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Attraction
	for rows.Next() {
		a, err := scanAttraction(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, rows.Err()
}

// GetAttraction returns a single attraction by ID.
func (d *DB) GetAttraction(ctx context.Context, id string) (Attraction, error) {
	row := d.conn.QueryRowContext(ctx,
		`SELECT id, name, description, category, location, image_url, tags, opening_hours, ticket_info, created_at, updated_at
		 FROM attractions WHERE id = ?`, id)
	a, err := scanAttractionRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Attraction{}, fmt.Errorf("attraction not found: %s", id)
		}
		return Attraction{}, err
	}
	return a, nil
}

// CreateAttraction inserts a new attraction and returns it.
func (d *DB) CreateAttraction(ctx context.Context, f AttractionForm) (Attraction, error) {
	if f.Tags == nil {
		f.Tags = []string{}
	}
	tagsJSON, err := json.Marshal(f.Tags)
	if err != nil {
		return Attraction{}, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	id := uuid.New().String()
	_, err = d.conn.ExecContext(ctx,
		`INSERT INTO attractions (id, name, description, category, location, image_url, tags, opening_hours, ticket_info, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, f.Name, f.Description, f.Category, f.Location, f.ImageURL, string(tagsJSON), f.OpeningHours, f.TicketInfo, now, now)
	if err != nil {
		return Attraction{}, err
	}
	return d.GetAttraction(ctx, id)
}

// UpdateAttraction updates an existing attraction.
func (d *DB) UpdateAttraction(ctx context.Context, id string, f AttractionForm) (Attraction, error) {
	if f.Tags == nil {
		f.Tags = []string{}
	}
	tagsJSON, err := json.Marshal(f.Tags)
	if err != nil {
		return Attraction{}, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := d.conn.ExecContext(ctx,
		`UPDATE attractions SET name=?, description=?, category=?, location=?, image_url=?, tags=?, opening_hours=?, ticket_info=?, updated_at=?
		 WHERE id = ?`,
		f.Name, f.Description, f.Category, f.Location, f.ImageURL, string(tagsJSON), f.OpeningHours, f.TicketInfo, now, id)
	if err != nil {
		return Attraction{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return Attraction{}, fmt.Errorf("attraction not found: %s", id)
	}
	return d.GetAttraction(ctx, id)
}

// DeleteAttraction removes an attraction. Returns ErrAttractionInUse if referenced by route_steps.
func (d *DB) DeleteAttraction(ctx context.Context, id string) error {
	// Check if any route_steps reference this attraction
	var count int
	err := d.conn.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM route_steps WHERE attraction_id = ?`, id).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrAttractionInUse
	}
	res, err := d.conn.ExecContext(ctx, `DELETE FROM attractions WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("attraction not found: %s", id)
	}
	return nil
}

// ── helpers ──

type scannable interface {
	Scan(dest ...any) error
}

func scanAttraction(row scannable) (Attraction, error) {
	var a Attraction
	var tagsJSON string
	err := row.Scan(&a.ID, &a.Name, &a.Description, &a.Category, &a.Location,
		&a.ImageURL, &tagsJSON, &a.OpeningHours, &a.TicketInfo, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return Attraction{}, err
	}
	if err := json.Unmarshal([]byte(tagsJSON), &a.Tags); err != nil {
		a.Tags = []string{}
	}
	return a, nil
}

func scanAttractionRow(row *sql.Row) (Attraction, error) {
	return scanAttraction(row)
}
