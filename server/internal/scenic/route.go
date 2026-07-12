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

var ErrRouteReferenced = errors.New("route is referenced by one or more characters")

// RouteStep represents one waypoint in a route.
type RouteStep struct {
	AttractionID     string `json:"attraction_id"`
	AttractionName   string `json:"attraction_name,omitempty"`
	Order            int    `json:"order"`
	DurationMinutes  int    `json:"duration_minutes"`
	Highlight        string `json:"highlight"`
}

// Route represents a scenic tour route.
type Route struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Duration    string      `json:"duration"`
	Difficulty  string      `json:"difficulty"`
	Tags        []string    `json:"tags"`
	Steps       []RouteStep `json:"steps"`
	CreatedAt   string      `json:"created_at"`
	UpdatedAt   string      `json:"updated_at"`
}

// RouteForm is the input for creating or updating a route.
type RouteForm struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Duration    string           `json:"duration"`
	Difficulty  string           `json:"difficulty"`
	Tags        []string         `json:"tags"`
	Steps       []RouteStepInput `json:"steps"`
}

// RouteStepInput is one step in the route form.
type RouteStepInput struct {
	AttractionID    string `json:"attraction_id"`
	Order           int    `json:"order"`
	DurationMinutes int    `json:"duration_minutes"`
	Highlight       string `json:"highlight"`
}

// ListRoutes returns all routes with their steps (attraction names resolved).
// Two-phase query to avoid SQLite deadlock with SetMaxOpenConns(1):
// first collect route metadata (close rows), then load steps individually.
func (d *DB) ListRoutes(ctx context.Context) ([]Route, error) {
	rows, err := d.conn.QueryContext(ctx,
		`SELECT id, name, description, duration, difficulty, tags, created_at, updated_at
		 FROM routes ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	var result []Route
	for rows.Next() {
		r, err := scanRouteMeta(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		result = append(result, r)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close() // Release connection before loading steps

	for i := range result {
		steps, err := d.loadSteps(ctx, result[i].ID)
		if err != nil {
			return nil, err
		}
		result[i].Steps = steps
	}
	return result, nil
}

// GetRoute returns a single route with its steps.
func (d *DB) GetRoute(ctx context.Context, id string) (Route, error) {
	row := d.conn.QueryRowContext(ctx,
		`SELECT id, name, description, duration, difficulty, tags, created_at, updated_at
		 FROM routes WHERE id = ?`, id)
	r, err := scanRouteMetaRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Route{}, fmt.Errorf("route not found: %s", id)
		}
		return Route{}, err
	}
	steps, err := d.loadSteps(ctx, r.ID)
	if err != nil {
		return Route{}, err
	}
	r.Steps = steps
	return r, nil
}

// CreateRoute inserts a new route with its steps in a transaction.
func (d *DB) CreateRoute(ctx context.Context, f RouteForm) (Route, error) {
	if f.Tags == nil {
		f.Tags = []string{}
	}
	tagsJSON, err := json.Marshal(f.Tags)
	if err != nil {
		return Route{}, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	id := uuid.New().String()

	tx, err := d.conn.BeginTx(ctx, nil)
	if err != nil {
		return Route{}, err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO routes (id, name, description, duration, difficulty, tags, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, f.Name, f.Description, f.Duration, f.Difficulty, string(tagsJSON), now, now)
	if err != nil {
		return Route{}, err
	}

	for _, step := range f.Steps {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO route_steps (route_id, attraction_id, step_order, duration_minutes, highlight)
			 VALUES (?, ?, ?, ?, ?)`,
			id, step.AttractionID, step.Order, step.DurationMinutes, step.Highlight)
		if err != nil {
			return Route{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return Route{}, err
	}
	return d.GetRoute(ctx, id)
}

// UpdateRoute replaces a route and its steps in a transaction.
func (d *DB) UpdateRoute(ctx context.Context, id string, f RouteForm) (Route, error) {
	if f.Tags == nil {
		f.Tags = []string{}
	}
	tagsJSON, err := json.Marshal(f.Tags)
	if err != nil {
		return Route{}, err
	}
	now := time.Now().UTC().Format(time.RFC3339)

	tx, err := d.conn.BeginTx(ctx, nil)
	if err != nil {
		return Route{}, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`UPDATE routes SET name=?, description=?, duration=?, difficulty=?, tags=?, updated_at=?
		 WHERE id = ?`,
		f.Name, f.Description, f.Duration, f.Difficulty, string(tagsJSON), now, id)
	if err != nil {
		return Route{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return Route{}, fmt.Errorf("route not found: %s", id)
	}

	// Delete old steps and insert new ones
	_, err = tx.ExecContext(ctx, `DELETE FROM route_steps WHERE route_id = ?`, id)
	if err != nil {
		return Route{}, err
	}
	for _, step := range f.Steps {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO route_steps (route_id, attraction_id, step_order, duration_minutes, highlight)
			 VALUES (?, ?, ?, ?, ?)`,
			id, step.AttractionID, step.Order, step.DurationMinutes, step.Highlight)
		if err != nil {
			return Route{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return Route{}, err
	}
	return d.GetRoute(ctx, id)
}

// DeleteRoute removes a route and its steps.
// NOTE: Character reference check (recommended_routes) is NOT done here because
// the characters table is in the file-based store, not in cyberverse.db.
// The API handler layer must check character.Store before calling this method.
func (d *DB) DeleteRoute(ctx context.Context, id string) error {
	tx, err := d.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `DELETE FROM route_steps WHERE route_id = ?`, id)
	if err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM routes WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("route not found: %s", id)
	}
	return tx.Commit()
}

// ── helpers ──

func (d *DB) loadSteps(ctx context.Context, routeID string) ([]RouteStep, error) {
	rows, err := d.conn.QueryContext(ctx,
		`SELECT rs.attraction_id, COALESCE(a.name, ''), rs.step_order, rs.duration_minutes, rs.highlight
		 FROM route_steps rs
		 LEFT JOIN attractions a ON a.id = rs.attraction_id
		 WHERE rs.route_id = ?
		 ORDER BY rs.step_order`, routeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var steps []RouteStep
	for rows.Next() {
		var s RouteStep
		if err := rows.Scan(&s.AttractionID, &s.AttractionName, &s.Order, &s.DurationMinutes, &s.Highlight); err != nil {
			return nil, err
		}
		steps = append(steps, s)
	}
	if steps == nil {
		steps = []RouteStep{}
	}
	return steps, rows.Err()
}

func scanRouteMeta(sc scannable) (Route, error) {
	var r Route
	var tagsJSON string
	err := sc.Scan(&r.ID, &r.Name, &r.Description, &r.Duration, &r.Difficulty,
		&tagsJSON, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return Route{}, err
	}
	if err := json.Unmarshal([]byte(tagsJSON), &r.Tags); err != nil {
		r.Tags = []string{}
	}
	return r, nil
}

func scanRouteMetaRow(row *sql.Row) (Route, error) {
	return scanRouteMeta(row)
}
