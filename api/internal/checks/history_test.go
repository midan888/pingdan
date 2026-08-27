package checks

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/pingdan/api/internal/db"
)

// TestHistorySinceBuckets exercises the bucketing SQL against a real Postgres.
// Set TEST_DATABASE_URL (e.g. the local docker-compose instance) to run it;
// without it the test is skipped so `go test ./...` stays hermetic.
func TestHistorySinceBuckets(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	if err := db.Migrate(dsn); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	ctx := context.Background()
	pool, err := db.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	var userID, endpointID string
	if err := pool.QueryRow(ctx, `
		INSERT INTO users (email, provider, provider_id)
		VALUES ('history-test@example.com', 'test', 'history-test')
		ON CONFLICT (email) DO UPDATE SET email=EXCLUDED.email RETURNING id
	`).Scan(&userID); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		INSERT INTO endpoints (user_id, name, url) VALUES ($1, 'history-test', 'http://example.com')
		RETURNING id
	`, userID).Scan(&endpointID); err != nil {
		t.Fatalf("seed endpoint: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id=$1`, userID)
	})

	// 24 hourly checks ending now. Hours 10-12 (counting from the oldest)
	// failed; everything else passed.
	until := time.Now().UTC().Truncate(time.Hour)
	since := until.Add(-24 * time.Hour)
	s := &Store{Pool: pool}
	for i := 0; i < 24; i++ {
		at := since.Add(time.Duration(i)*time.Hour + 30*time.Minute)
		ok := i < 10 || i > 12
		lat := 30
		if err := s.Insert(ctx, &Check{EndpointID: endpointID, OK: ok, LatencyMs: &lat, CheckedAt: at}); err != nil {
			t.Fatalf("insert check %d: %v", i, err)
		}
	}

	h, err := s.HistorySince(ctx, endpointID, since, until, 24)
	if err != nil {
		t.Fatalf("HistorySince: %v", err)
	}
	if len(h.Buckets) != 24 {
		t.Fatalf("len(Buckets) = %d, want 24", len(h.Buckets))
	}
	if h.Total != 24 || h.Failed != 3 {
		t.Errorf("Total/Failed = %d/%d, want 24/3", h.Total, h.Failed)
	}
	if got, want := h.UptimePct, 87.5; got != want {
		t.Errorf("UptimePct = %v, want %v", got, want)
	}
	if h.BucketSec != 3600 {
		t.Errorf("BucketSec = %d, want 3600", h.BucketSec)
	}
	for i, b := range h.Buckets {
		wantFailed := 0
		if i >= 10 && i <= 12 {
			wantFailed = 1
		}
		if b.Total != 1 || b.Failed != wantFailed {
			t.Errorf("bucket %d = %d/%d total/failed, want 1/%d", i, b.Total, b.Failed, wantFailed)
		}
	}

	// A coarser bar over the same window must still surface the outage rather
	// than averaging it away. 6 buckets of 4h: the failures at hours 10.5 and
	// 11.5 land in bucket 2, the one at 12.5 in bucket 3.
	coarse, err := s.HistorySince(ctx, endpointID, since, until, 6)
	if err != nil {
		t.Fatalf("HistorySince(6): %v", err)
	}
	if len(coarse.Buckets) != 6 || coarse.Failed != 3 {
		t.Fatalf("coarse = %d buckets / %d failed, want 6 / 3", len(coarse.Buckets), coarse.Failed)
	}
	wantCoarse := []int{0, 0, 2, 1, 0, 0}
	for i, want := range wantCoarse {
		if coarse.Buckets[i].Failed != want {
			t.Errorf("coarse bucket %d failed = %d, want %d", i, coarse.Buckets[i].Failed, want)
		}
		if coarse.Buckets[i].Total != 4 {
			t.Errorf("coarse bucket %d total = %d, want 4", i, coarse.Buckets[i].Total)
		}
	}

	// An endpoint with no checks in the window yields empty buckets, not an error.
	empty, err := s.HistorySince(ctx, endpointID, since.Add(-48*time.Hour), since.Add(-24*time.Hour), 5)
	if err != nil {
		t.Fatalf("HistorySince(empty): %v", err)
	}
	if empty.Total != 0 || len(empty.Buckets) != 5 || empty.UptimePct != 0 {
		t.Errorf("empty window = %+v, want 5 zero buckets", empty)
	}
}
