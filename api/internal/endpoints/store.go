package endpoints

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	CheckTypeHTTP = "http"
	CheckTypeTCP  = "tcp"
	CheckTypeICMP = "icmp"
)

type Endpoint struct {
	ID                  string     `json:"id"`
	UserID              string     `json:"-"`
	GroupID             *string    `json:"groupId"`
	Name                string     `json:"name"`
	CheckType           string     `json:"checkType"`
	URL                 string     `json:"url"`
	Method              string     `json:"method"`
	ExpectedStatus      int        `json:"expectedStatus"`
	IntervalSec         int        `json:"intervalSec"`
	TimeoutSec          int        `json:"timeoutSec"`
	FailureThreshold    int        `json:"failureThreshold"`
	Enabled             bool       `json:"enabled"`
	CurrentState        string     `json:"currentState"`
	ConsecutiveFailures int        `json:"consecutiveFailures"`
	LastCheckedAt       *time.Time `json:"lastCheckedAt"`
	CreatedAt           time.Time  `json:"createdAt"`

	// SSL/TLS certificate monitoring (populated by the daily ssl checker;
	// NULL for non-HTTPS endpoints or before the first check).
	SSLExpiresAt     *time.Time `json:"sslExpiresAt"`
	SSLLastCheckedAt *time.Time `json:"sslLastCheckedAt"`
	SSLLastError     *string    `json:"sslLastError"`
	SSLAlertedDays   *int       `json:"-"`
}

type Store struct{ Pool *pgxpool.Pool }

func (s *Store) List(ctx context.Context, userID string) ([]Endpoint, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT id, user_id, group_id, name, check_type, url, method, expected_status, interval_sec, timeout_sec,
		       failure_threshold, enabled, current_state, consecutive_failures, last_checked_at, created_at,
		       ssl_expires_at, ssl_last_checked_at, ssl_last_error
		FROM endpoints WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Endpoint{}
	for rows.Next() {
		var e Endpoint
		if err := rows.Scan(&e.ID, &e.UserID, &e.GroupID, &e.Name, &e.CheckType, &e.URL, &e.Method, &e.ExpectedStatus,
			&e.IntervalSec, &e.TimeoutSec, &e.FailureThreshold, &e.Enabled, &e.CurrentState,
			&e.ConsecutiveFailures, &e.LastCheckedAt, &e.CreatedAt,
			&e.SSLExpiresAt, &e.SSLLastCheckedAt, &e.SSLLastError); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Store) Create(ctx context.Context, e *Endpoint) error {
	// group_id is resolved through an ownership-guarded subselect so a group
	// belonging to another user (or an unknown id) lands as NULL.
	return s.Pool.QueryRow(ctx, `
		INSERT INTO endpoints (user_id, group_id, name, check_type, url, method, expected_status, interval_sec, timeout_sec, failure_threshold, enabled)
		VALUES ($1, (SELECT id FROM groups WHERE id=$2 AND user_id=$1), $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, current_state, consecutive_failures, created_at
	`, e.UserID, e.GroupID, e.Name, e.CheckType, e.URL, e.Method, e.ExpectedStatus, e.IntervalSec, e.TimeoutSec, e.FailureThreshold, e.Enabled).
		Scan(&e.ID, &e.CurrentState, &e.ConsecutiveFailures, &e.CreatedAt)
}

func (s *Store) Update(ctx context.Context, userID, id string, e *Endpoint) error {
	// group_id is set via an ownership-guarded subselect: a group from another
	// user (or an unknown id) resolves to NULL rather than being trusted.
	_, err := s.Pool.Exec(ctx, `
		UPDATE endpoints SET name=$1, check_type=$2, url=$3, method=$4, expected_status=$5, interval_sec=$6,
		    timeout_sec=$7, failure_threshold=$8, enabled=$9,
		    group_id=(SELECT id FROM groups WHERE id=$10 AND user_id=$11),
		    updated_at=now()
		WHERE id=$12 AND user_id=$11
	`, e.Name, e.CheckType, e.URL, e.Method, e.ExpectedStatus, e.IntervalSec, e.TimeoutSec, e.FailureThreshold, e.Enabled, e.GroupID, userID, id)
	return err
}

func (s *Store) Delete(ctx context.Context, userID, id string) error {
	_, err := s.Pool.Exec(ctx, `DELETE FROM endpoints WHERE id=$1 AND user_id=$2`, id, userID)
	return err
}

// ListEnabledAll returns every enabled endpoint across all users — used by the pinger scheduler at startup.
func (s *Store) ListEnabledAll(ctx context.Context) ([]Endpoint, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT id, user_id, group_id, name, check_type, url, method, expected_status, interval_sec, timeout_sec,
		       failure_threshold, enabled, current_state, consecutive_failures, last_checked_at, created_at,
		       ssl_expires_at, ssl_last_checked_at, ssl_last_error, ssl_alerted_days
		FROM endpoints WHERE enabled = TRUE
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Endpoint{}
	for rows.Next() {
		var e Endpoint
		if err := rows.Scan(&e.ID, &e.UserID, &e.GroupID, &e.Name, &e.CheckType, &e.URL, &e.Method, &e.ExpectedStatus,
			&e.IntervalSec, &e.TimeoutSec, &e.FailureThreshold, &e.Enabled, &e.CurrentState,
			&e.ConsecutiveFailures, &e.LastCheckedAt, &e.CreatedAt,
			&e.SSLExpiresAt, &e.SSLLastCheckedAt, &e.SSLLastError, &e.SSLAlertedDays); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Store) GetByID(ctx context.Context, id string) (*Endpoint, error) {
	var e Endpoint
	err := s.Pool.QueryRow(ctx, `
		SELECT id, user_id, group_id, name, check_type, url, method, expected_status, interval_sec, timeout_sec,
		       failure_threshold, enabled, current_state, consecutive_failures, last_checked_at, created_at,
		       ssl_expires_at, ssl_last_checked_at, ssl_last_error, ssl_alerted_days
		FROM endpoints WHERE id=$1
	`, id).Scan(&e.ID, &e.UserID, &e.GroupID, &e.Name, &e.CheckType, &e.URL, &e.Method, &e.ExpectedStatus,
		&e.IntervalSec, &e.TimeoutSec, &e.FailureThreshold, &e.Enabled, &e.CurrentState,
		&e.ConsecutiveFailures, &e.LastCheckedAt, &e.CreatedAt,
		&e.SSLExpiresAt, &e.SSLLastCheckedAt, &e.SSLLastError, &e.SSLAlertedDays)
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

// UpdateSSL records the result of a TLS certificate check. expiresAt is nil
// when the check failed (errMsg explains why); alertedDays carries the last
// "days remaining" bucket we warned about, or nil to clear it.
func (s *Store) UpdateSSL(ctx context.Context, id string, expiresAt *time.Time, errMsg *string, checkedAt time.Time, alertedDays *int) error {
	_, err := s.Pool.Exec(ctx, `
		UPDATE endpoints SET ssl_expires_at=$1, ssl_last_error=$2, ssl_last_checked_at=$3,
		    ssl_alerted_days=$4, updated_at=now()
		WHERE id=$5
	`, expiresAt, errMsg, checkedAt, alertedDays, id)
	return err
}
