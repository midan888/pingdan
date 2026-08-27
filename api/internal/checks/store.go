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

// Bucket is one equal-width time slice of an endpoint's check history.
// A bucket with Total == 0 had no checks in it (monitor paused, created
// mid-window, or a gap in collection) and renders as "no data".
type Bucket struct {
	Total  int `json:"total"`
	Failed int `json:"failed"`
}

// History summarises an endpoint's checks over a window, both in aggregate and
// split into equally sized time buckets (oldest first).
//
// Bucketing happens in SQL on purpose: a 24h window at a 60s interval is 1440
// rows per monitor, and the dashboard refreshes every monitor every 15s. The
// grouped query returns one row per bucket instead, so the status bar can span
// the same window as the uptime figure next to it without shipping the raw
// check log for the whole dashboard.
type History struct {
	From      time.Time `json:"from"`
	To        time.Time `json:"to"`
	BucketSec int       `json:"bucketSec"`
	Total     int       `json:"total"`
	Failed    int       `json:"failed"`
	UptimePct float64   `json:"uptimePct"`
	Buckets   []Bucket  `json:"buckets"`
}

// HistorySince buckets [since, until) into n slices of equal duration.
func (s *Store) HistorySince(ctx context.Context, endpointID string, since, until time.Time, n int) (History, error) {
	if n <= 0 {
		n = 30
	}
	h := History{From: since, To: until, Buckets: make([]Bucket, n)}
	span := until.Sub(since)
	if span <= 0 {
		return h, nil
	}
	bucketSec := span.Seconds() / float64(n)
	h.BucketSec = int(bucketSec + 0.5)

	// The index is clamped either side so a row landing exactly on `until`
	// (or on a boundary rounded the wrong way) folds into the edge bucket
	// rather than overflowing the slice.
	rows, err := s.Pool.Query(ctx, `
		SELECT LEAST($4::int - 1, GREATEST(0, floor(EXTRACT(EPOCH FROM (checked_at - $2::timestamptz)) / $3::float8)::int)) AS bucket,
		       COUNT(*)::int AS total,
		       COUNT(*) FILTER (WHERE NOT ok)::int AS failed
		FROM checks
		WHERE endpoint_id=$1 AND checked_at >= $2::timestamptz AND checked_at < $5::timestamptz
		GROUP BY bucket
		ORDER BY bucket
	`, endpointID, since, bucketSec, n, until)
	if err != nil {
		return History{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var idx, total, failed int
		if err := rows.Scan(&idx, &total, &failed); err != nil {
			return History{}, err
		}
		if idx < 0 || idx >= n {
			continue
		}
		h.Buckets[idx] = Bucket{Total: total, Failed: failed}
		h.Total += total
		h.Failed += failed
	}
	if err := rows.Err(); err != nil {
		return History{}, err
	}
	if h.Total > 0 {
		h.UptimePct = float64(h.Total-h.Failed) / float64(h.Total) * 100
	}
	return h, nil
}
