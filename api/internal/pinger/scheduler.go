package pinger

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/pingdan/api/internal/alerts"
	"github.com/pingdan/api/internal/checks"
	"github.com/pingdan/api/internal/endpoints"
)

type Scheduler struct {
	Endpoints *endpoints.Store
	Checks    *checks.Store
	Alerts    *alerts.Dispatcher
	Logger    *slog.Logger

	mu      sync.Mutex
	workers map[string]*worker // endpointID -> worker
	parent  context.Context
}

func NewScheduler(parent context.Context, e *endpoints.Store, c *checks.Store, a *alerts.Dispatcher, l *slog.Logger) *Scheduler {
	return &Scheduler{
		Endpoints: e, Checks: c, Alerts: a, Logger: l,
		workers: map[string]*worker{},
		parent:  parent,
	}
}

// Start loads all enabled endpoints and spawns a worker per endpoint.
func (s *Scheduler) Start(ctx context.Context) error {
	all, err := s.Endpoints.ListEnabledAll(ctx)
	if err != nil {
		return err
	}
	for _, e := range all {
		s.Upsert(e)
	}
	s.Logger.Info("pinger started", "endpoints", len(all))
	return nil
}

func (s *Scheduler) Upsert(e endpoints.Endpoint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if w, ok := s.workers[e.ID]; ok {
		w.update(e)
		return
	}
	w := newWorker(s, e)
	s.workers[e.ID] = w
	go w.run(s.parent)
}

func (s *Scheduler) Remove(endpointID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if w, ok := s.workers[endpointID]; ok {
		w.stop()
		delete(s.workers, endpointID)
	}
}

type worker struct {
	s       *Scheduler
	ep      endpoints.Endpoint
	updates chan endpoints.Endpoint
	done    chan struct{}
}

func newWorker(s *Scheduler, e endpoints.Endpoint) *worker {
	return &worker{
		s:       s,
		ep:      e,
		updates: make(chan endpoints.Endpoint, 4),
		done:    make(chan struct{}),
	}
}

func (w *worker) update(e endpoints.Endpoint) {
	select {
	case w.updates <- e:
	default:
	}
}

func (w *worker) stop() { close(w.done) }

func (w *worker) run(ctx context.Context) {
	interval := time.Duration(w.ep.IntervalSec) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// run an immediate first check
	w.tick(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.done:
			return
		case e := <-w.updates:
			w.ep = e
			if newIv := time.Duration(e.IntervalSec) * time.Second; newIv != interval {
				interval = newIv
				ticker.Reset(interval)
			}
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

var client = &http.Client{
	// Per-request timeout is enforced via context; this guards against absurdly slow connection setup.
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 4,
		IdleConnTimeout:     90 * time.Second,
	},
}

func (w *worker) tick(ctx context.Context) {
	rctx, cancel := context.WithTimeout(ctx, time.Duration(w.ep.TimeoutSec)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(rctx, w.ep.Method, w.ep.URL, nil)
	if err != nil {
		w.record(ctx, nil, nil, false, err.Error())
		return
	}
	req.Header.Set("User-Agent", "pingdan/1.0")

	start := time.Now()
	resp, err := client.Do(req)
	latency := int(time.Since(start).Milliseconds())

	if err != nil {
		msg := err.Error()
		w.record(ctx, nil, &latency, false, msg)
		return
	}
	defer resp.Body.Close()

	ok := resp.StatusCode == w.ep.ExpectedStatus
	var errMsg *string
	if !ok {
		m := "unexpected status"
		errMsg = &m
	}
	w.record(ctx, &resp.StatusCode, &latency, ok, ptrOrNil(errMsg))
}

func ptrOrNil(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (w *worker) record(ctx context.Context, status, latency *int, ok bool, errMsg string) {
	now := time.Now().UTC()
	var errPtr *string
	if errMsg != "" {
		errPtr = &errMsg
	}
	c := &checks.Check{
		EndpointID: w.ep.ID,
		StatusCode: status,
		LatencyMs:  latency,
		OK:         ok,
		Error:      errPtr,
		CheckedAt:  now,
	}
	if err := w.s.Checks.Insert(ctx, c); err != nil {
		w.s.Logger.Error("insert check failed", "err", err, "endpoint", w.ep.ID)
	}

	prevState := w.ep.CurrentState
	prevFails := w.ep.ConsecutiveFailures

	var newState string
	newFails := prevFails
	if ok {
		newFails = 0
		newState = "up"
	} else {
		newFails++
		if newFails >= w.ep.FailureThreshold {
			newState = "down"
		} else {
			newState = prevState
			if prevState == "unknown" {
				newState = "unknown"
			}
		}
	}

	if err := w.s.Endpoints.UpdateState(ctx, w.ep.ID, newState, newFails, now); err != nil {
		w.s.Logger.Error("update endpoint state failed", "err", err, "endpoint", w.ep.ID)
	}
	w.ep.CurrentState = newState
	w.ep.ConsecutiveFailures = newFails

	// fire alert on state transition
	if prevState != newState && (newState == "down" || (prevState == "down" && newState == "up")) {
		w.s.Alerts.Notify(ctx, w.ep, newState, c)
	}
}
