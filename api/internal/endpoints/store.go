package endpoints

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Endpoint struct {
	ID                   string     `json:"id"`
	UserID               string     `json:"-"`
	Name                 string     `json:"name"`
	URL                  string     `json:"url"`
	Method               string     `json:"method"`
	ExpectedStatus       int        `json:"expectedStatus"`
	IntervalSec          int        `json:"intervalSec"`
	TimeoutSec           int        `json:"timeoutSec"`
	FailureThreshold     int        `json:"failureThreshold"`
	Enabled              bool       `json:"enabled"`
	CurrentState         string     `json:"currentState"`
	ConsecutiveFailures  int        `json:"consecutiveFailures"`
	LastCheckedAt        *time.Time `json:"lastCheckedAt"`
	CreatedAt            time.Time  `json:"createdAt"`
}

type Store struct{ Pool *pgxpool.Pool }

func (s *Store) List(ctx context.Context, userID string) ([]Endpoint, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT id, user_id, name, url, method, expected_status, interval_sec, timeout_sec,
		       failure_threshold, enabled, current_state, consecutive_failures, last_checked_at, created_at
		FROM endpoints WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Endpoint{}
	for rows.Next() {
		var e Endpoint
		if err := rows.Scan(&e.ID, &e.UserID, &e.Name, &e.URL, &e.Method, &e.ExpectedStatus,
			&e.IntervalSec, &e.TimeoutSec, &e.FailureThreshold, &e.Enabled, &e.CurrentState,
			&e.ConsecutiveFailures, &e.LastCheckedAt, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Store) Create(ctx context.Context, e *Endpoint) error {
	return s.Pool.QueryRow(ctx, `
		INSERT INTO endpoints (user_id, name, url, method, expected_status, interval_sec, timeout_sec, failure_threshold, enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, current_state, consecutive_failures, created_at
	`, e.UserID, e.Name, e.URL, e.Method, e.ExpectedStatus, e.IntervalSec, e.TimeoutSec, e.FailureThreshold, e.Enabled).
		Scan(&e.ID, &e.CurrentState, &e.ConsecutiveFailures, &e.CreatedAt)
}

func (s *Store) Update(ctx context.Context, userID, id string, e *Endpoint) error {
	_, err := s.Pool.Exec(ctx, `
		UPDATE endpoints SET name=$1, url=$2, method=$3, expected_status=$4, interval_sec=$5,
		    timeout_sec=$6, failure_threshold=$7, enabled=$8, updated_at=now()
		WHERE id=$9 AND user_id=$10
	`, e.Name, e.URL, e.Method, e.ExpectedStatus, e.IntervalSec, e.TimeoutSec, e.FailureThreshold, e.Enabled, id, userID)
	return err
}

func (s *Store) Delete(ctx context.Context, userID, id string) error {
	_, err := s.Pool.Exec(ctx, `DELETE FROM endpoints WHERE id=$1 AND user_id=$2`, id, userID)
	return err
}

// ListEnabledAll returns every enabled endpoint across all users — used by the pinger scheduler at startup.
func (s *Store) ListEnabledAll(ctx context.Context) ([]Endpoint, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT id, user_id, name, url, method, expected_status, interval_sec, timeout_sec,
		       failure_threshold, enabled, current_state, consecutive_failures, last_checked_at, created_at
		FROM endpoints WHERE enabled = TRUE
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Endpoint{}
	for rows.Next() {
		var e Endpoint
		if err := rows.Scan(&e.ID, &e.UserID, &e.Name, &e.URL, &e.Method, &e.ExpectedStatus,
			&e.IntervalSec, &e.TimeoutSec, &e.FailureThreshold, &e.Enabled, &e.CurrentState,
			&e.ConsecutiveFailures, &e.LastCheckedAt, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Store) GetByID(ctx context.Context, id string) (*Endpoint, error) {
	var e Endpoint
	err := s.Pool.QueryRow(ctx, `
		SELECT id, user_id, name, url, method, expected_status, interval_sec, timeout_sec,
		       failure_threshold, enabled, current_state, consecutive_failures, last_checked_at, created_at
		FROM endpoints WHERE id=$1
	`, id).Scan(&e.ID, &e.UserID, &e.Name, &e.URL, &e.Method, &e.ExpectedStatus,
		&e.IntervalSec, &e.TimeoutSec, &e.FailureThreshold, &e.Enabled, &e.CurrentState,
		&e.ConsecutiveFailures, &e.LastCheckedAt, &e.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &e, err
}

func (s *Store) UpdateState(ctx context.Context, id, state string, consecFails int, checkedAt time.Time) error {
	_, err := s.Pool.Exec(ctx, `
		UPDATE endpoints SET current_state=$1, consecutive_failures=$2, last_checked_at=$3, updated_at=now()
		WHERE id=$4
	`, state, consecFails, checkedAt, id)
	return err
}
