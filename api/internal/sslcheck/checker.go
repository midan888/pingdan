// Package sslcheck monitors the TLS certificates of HTTPS endpoints. Once a
// day it dials each endpoint's host, reads the leaf certificate's expiry, and —
// once the cert drops to AlertThresholdDays or fewer days remaining — sends one
// warning per day per channel (15 days left, 14 days left, ...) until the cert
// is renewed or has expired.
package sslcheck

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"math"
	"net"
	"net/url"
	"time"

	"github.com/pingdan/api/internal/alerts"
	"github.com/pingdan/api/internal/endpoints"
)

const (
	// AlertThresholdDays is the "X days or fewer left" point at which daily
	// expiry warnings begin.
	AlertThresholdDays = 15

	// dialTimeout caps how long a single TLS handshake may take.
	dialTimeout = 10 * time.Second

	// checkInterval is how often the checker sweeps all endpoints. Certs change
	// rarely, so a daily-ish cadence is plenty and keeps this off the hot path.
	checkInterval = 12 * time.Hour
)

type Checker struct {
	Endpoints *endpoints.Store
	Alerts    *alerts.Dispatcher
	Logger    *slog.Logger
}

func New(e *endpoints.Store, a *alerts.Dispatcher, l *slog.Logger) *Checker {
	return &Checker{Endpoints: e, Alerts: a, Logger: l}
}

// Run sweeps once immediately, then on a fixed interval until ctx is cancelled.
func (c *Checker) Run(ctx context.Context) {
	c.sweep(ctx)
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.sweep(ctx)
		}
	}
}

// sweep checks every enabled HTTPS endpoint once.
func (c *Checker) sweep(ctx context.Context) {
	all, err := c.Endpoints.ListEnabledAll(ctx)
	if err != nil {
		c.Logger.Error("sslcheck: list endpoints", "err", err)
		return
	}
	checked := 0
	for _, e := range all {
		if !isHTTPS(e.URL) {
			continue
		}
		c.checkOne(ctx, e)
		checked++
	}
	c.Logger.Info("sslcheck: sweep complete", "https_endpoints", checked)
}

func isHTTPS(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && u.Scheme == "https"
}

// CheckEndpoint runs an on-demand certificate check for one endpoint (used by
// the manual "check now" API). It is a no-op for non-HTTPS endpoints.
func (c *Checker) CheckEndpoint(ctx context.Context, e endpoints.Endpoint) {
	if !isHTTPS(e.URL) {
		return
	}
	c.checkOne(ctx, e)
}

// checkOne inspects a single endpoint's certificate, persists the result, and
// fires a warning when warranted.
func (c *Checker) checkOne(ctx context.Context, e endpoints.Endpoint) {
	now := time.Now().UTC()

	expiresAt, err := fetchCertExpiry(ctx, e.URL)
	if err != nil {
		msg := err.Error()
		// On failure we keep any prior expiry/alert state untouched except the
		// error + checked timestamp, so a transient network blip doesn't wipe a
		// known-good expiry. We do this by reading current values back in.
		if uerr := c.Endpoints.UpdateSSL(ctx, e.ID, e.SSLExpiresAt, &msg, now, e.SSLAlertedDays); uerr != nil {
			c.Logger.Error("sslcheck: persist error", "err", uerr, "endpoint", e.ID)
		}
		c.Logger.Warn("sslcheck: handshake failed", "endpoint", e.ID, "url", e.URL, "err", msg)
		return
	}

	daysLeft := daysUntil(expiresAt, now)
	alertedDays := e.SSLAlertedDays

	// Decide whether to alert. We send one warning per distinct day-count once
	// the cert is at or below the threshold (and not yet expired). ssl_alerted_days
	// remembers the last bucket warned, so re-running the sweep the same day
	// (the interval is sub-daily) won't double-send.
	if daysLeft <= AlertThresholdDays && daysLeft >= 0 {
		if alertedDays == nil || *alertedDays != daysLeft {
			c.Alerts.NotifySSL(ctx, e, daysLeft, expiresAt)
			d := daysLeft
			alertedDays = &d
			c.Logger.Info("sslcheck: expiry warning sent", "endpoint", e.ID, "days_left", daysLeft)
		}
	} else {
		// Outside the warning window (renewed, or still comfortably valid):
		// clear the bucket so a fresh countdown alerts again if it re-enters.
		alertedDays = nil
	}

	if err := c.Endpoints.UpdateSSL(ctx, e.ID, &expiresAt, nil, now, alertedDays); err != nil {
		c.Logger.Error("sslcheck: persist result", "err", err, "endpoint", e.ID)
	}
}

// daysUntil returns whole days from now until t, floored toward negative
// infinity. A cert expiring in 23h59m reports 0 (expiring today); one that
// expired 2h ago reports -1 (so the warning window's >= 0 guard excludes it).
func daysUntil(t, now time.Time) int {
	return int(math.Floor(t.Sub(now).Hours() / 24))
}

// fetchCertExpiry dials the endpoint host over TLS and returns the leaf
// certificate's NotAfter time.
func fetchCertExpiry(ctx context.Context, raw string) (time.Time, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return time.Time{}, err
	}
	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "443"
	}

	dialer := &net.Dialer{Timeout: dialTimeout}
	conn, err := tls.DialWithDialer(dialer, "tcp", net.JoinHostPort(host, port), &tls.Config{
		ServerName: host,
	})
	if err != nil {
		return time.Time{}, fmt.Errorf("tls handshake: %w", err)
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return time.Time{}, fmt.Errorf("no peer certificates")
	}
	// The leaf certificate is always first.
	return certs[0].NotAfter, nil
}
