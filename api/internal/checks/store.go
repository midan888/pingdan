package checks

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Check struct {
	ID         int64     `json:"id"`
	EndpointID string    `json:"endpointId"`
	StatusCode *int      `json:"statusCode,omitempty"`
	LatencyMs  *int      `json:"latencyMs,omitempty"`
	OK         bool      `json:"ok"`
	Error      *string   `json:"error,omitempty"`
	CheckedAt  time.Time `json:"checkedAt"`
}

type Store struct{ Pool *pgxpool.Pool }

func (s *Store) Insert(ctx context.Context, c *Check) error {
	return s.Pool.QueryRow(ctx, `
		INSERT INTO checks (endpoint_id, status_code, latency_ms, ok, error, checked_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, c.EndpointID, c.StatusCode, c.LatencyMs, c.OK, c.Error, c.CheckedAt).Scan(&c.ID)
}

func (s *Store) Recent(ctx context.Context, endpointID string, limit int) ([]Check, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.Pool.Query(ctx, `
		SELECT id, endpoint_id, status_code, latency_ms, ok, error, checked_at
		FROM checks WHERE endpoint_id=$1 ORDER BY checked_at DESC LIMIT $2
	`, endpointID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Check{}
	for rows.Next() {
		var c Check
		if err := rows.Scan(&c.ID, &c.EndpointID, &c.StatusCode, &c.LatencyMs, &c.OK, &c.Error, &c.CheckedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
