package sslcheck

import (
	"testing"
	"time"
)

func TestDaysUntil(t *testing.T) {
	now := time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name  string
		until time.Time
		want  int
	}{
		{"exactly 15 days", now.Add(15 * 24 * time.Hour), 15},
		{"23h59m rounds to 0", now.Add(23*time.Hour + 59*time.Minute), 0},
		{"expires in 1.5 days", now.Add(36 * time.Hour), 1},
		{"already expired", now.Add(-2 * time.Hour), -1},
		{"45 days out", now.Add(45 * 24 * time.Hour), 45},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := daysUntil(c.until, now); got != c.want {
				t.Errorf("daysUntil = %d, want %d", got, c.want)
			}
		})
	}
}

func TestIsHTTPS(t *testing.T) {
	cases := map[string]bool{
		"https://example.com":      true,
		"https://example.com:8443": true,
		"http://example.com":       false,
		"ftp://example.com":        false,
		"not a url":                false,
	}
	for in, want := range cases {
		if got := isHTTPS(in); got != want {
			t.Errorf("isHTTPS(%q) = %v, want %v", in, got, want)
		}
	}
}

// shouldAlert mirrors the decision made in checkOne: alert once per distinct
// day bucket at or below the threshold (and not yet expired).
func shouldAlert(daysLeft int, alertedDays *int) bool {
	if daysLeft > AlertThresholdDays || daysLeft < 0 {
		return false
	}
	return alertedDays == nil || *alertedDays != daysLeft
}

func TestAlertCadence(t *testing.T) {
	d := func(n int) *int { return &n }

	cases := []struct {
		name     string
		daysLeft int
		alerted  *int
		want     bool
	}{
		{"first entry into window at 15", 15, nil, true},
		{"same day, already alerted 15", 15, d(15), false},
		{"next day ticks to 14", 14, d(15), true},
		{"comfortably valid, 31 days", 31, nil, false},
		{"just above threshold, 16 days", 16, nil, false},
		{"expired, no alert", -1, d(0), false},
		{"last day, 0 left", 0, d(1), true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := shouldAlert(c.daysLeft, c.alerted); got != c.want {
				t.Errorf("shouldAlert(%d, %v) = %v, want %v", c.daysLeft, c.alerted, got, c.want)
			}
		})
	}
}
