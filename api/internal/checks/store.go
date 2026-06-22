package checks

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Check struct {
	ID               int64           `json:"id"`
	EndpointID       string          `json:"endpointId"`
	StatusCode       *int            `json:"statusCode,omitempty"`
	LatencyMs        *int            `json:"latencyMs,omitempty"`
	OK               bool            `json:"ok"`
	Error            *string         `json:"error,omitempty"`
	FailedAssertions json.RawMessage `json:"failedAssertions,omitempty"`
	CheckedAt        time.Time       `json:"checkedAt"`
}

type Store struct{ Pool *pgxpool.Pool }

func (s *Store) Insert(ctx context.Context, c *Check) error {
	return s.Pool.QueryRow(ctx, `
		INSERT INTO checks (endpoint_id, status_code, latency_ms, ok, error, failed_assertions, checked_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, c.EndpointID, c.StatusCode, c.LatencyMs, c.OK, c.Error, c.FailedAssertions, c.CheckedAt).Scan(&c.ID)
}

func (s *Store) Recent(ctx context.Context, endpointID string, limit int) ([]Check, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.Pool.Query(ctx, `
		SELECT id, endpoint_id, status_code, latency_ms, ok, error, failed_assertions, checked_at
		FROM checks WHERE endpoint_id=$1 ORDER BY checked_at DESC LIMIT $2
	`, endpointID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Check{}
	for rows.Next() {
		var c Check
		if err := rows.Scan(&c.ID, &c.EndpointID, &c.StatusCode, &c.LatencyMs, &c.OK, &c.Error, &c.FailedAssertions, &c.CheckedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// RecentSince returns checks within the given time window (newest first),
// capped at limit to bound the response for high-frequency endpoints.
func (s *Store) RecentSince(ctx context.Context, endpointID string, since time.Time, limit int) ([]Check, error) {
	if limit <= 0 || limit > 2000 {
		limit = 2000
	}
	rows, err := s.Pool.Query(ctx, `
		SELECT id, endpoint_id, status_code, latency_ms, ok, error, failed_assertions, checked_at
		FROM checks WHERE endpoint_id=$1 AND checked_at >= $2 ORDER BY checked_at DESC LIMIT $3
	`, endpointID, since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Check{}
	for rows.Next() {
		var c Check
		if err := rows.Scan(&c.ID, &c.EndpointID, &c.StatusCode, &c.LatencyMs, &c.OK, &c.Error, &c.FailedAssertions, &c.CheckedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// Stats summarises recent checks within the given window for an endpoint.
type Stats struct {
	Total      int      `json:"total"`
	UpCount    int      `json:"upCount"`
	UptimePct  float64  `json:"uptimePct"`
	AvgLatency *float64 `json:"avgLatencyMs"`
	P50Latency *int     `json:"p50LatencyMs"`
	P95Latency *int     `json:"p95LatencyMs"`
	MinLatency *int     `json:"minLatencyMs"`
	MaxLatency *int     `json:"maxLatencyMs"`
}

// StatsSince computes aggregate stats over checks since the given time.
func (s *Store) StatsSince(ctx context.Context, endpointID string, since time.Time) (Stats, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT ok, latency_ms FROM checks
		WHERE endpoint_id=$1 AND checked_at >= $2
	`, endpointID, since)
	if err != nil {
		return Stats{}, err
	}
	defer rows.Close()

	var st Stats
	var latencies []int
	var sum float64
	for rows.Next() {
		var ok bool
		var lat *int
		if err := rows.Scan(&ok, &lat); err != nil {
			return Stats{}, err
		}
		st.Total++
		if ok {
			st.UpCount++
		}
		if lat != nil {
			latencies = append(latencies, *lat)
			sum += float64(*lat)
		}
	}
	if err := rows.Err(); err != nil {
		return Stats{}, err
	}
	if st.Total > 0 {
		st.UptimePct = float64(st.UpCount) / float64(st.Total) * 100
	}
	if n := len(latencies); n > 0 {
		sortInts(latencies)
		avg := sum / float64(n)
		st.AvgLatency = &avg
		p50 := latencies[pctIdx(n, 50)]
		p95 := latencies[pctIdx(n, 95)]
		mn := latencies[0]
		mx := latencies[n-1]
		st.P50Latency = &p50
		st.P95Latency = &p95
		st.MinLatency = &mn
		st.MaxLatency = &mx
	}
	return st, nil
}

func pctIdx(n, pct int) int {
	idx := (pct * n) / 100
	if idx >= n {
		idx = n - 1
	}
	return idx
}

func sortInts(a []int) {
	for i := 1; i < len(a); i++ {
		for j := i; j > 0 && a[j-1] > a[j]; j-- {
			a[j-1], a[j] = a[j], a[j-1]
		}
	}
}
