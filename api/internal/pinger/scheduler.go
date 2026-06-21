package pinger

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pingdan/api/internal/alerts"
	"github.com/pingdan/api/internal/assertions"
	"github.com/pingdan/api/internal/checks"
	"github.com/pingdan/api/internal/endpoints"
)

// maxBodyRead caps how much of a response body we read for assertion evaluation.
const maxBodyRead = 1 << 20 // 1 MiB

type Scheduler struct {
	Endpoints  *endpoints.Store
	Checks     *checks.Store
	Assertions *assertions.Store
	Alerts     *alerts.Dispatcher
	Logger     *slog.Logger

	mu      sync.Mutex
	workers map[string]*worker // endpointID -> worker
	parent  context.Context
}

func NewScheduler(parent context.Context, e *endpoints.Store, c *checks.Store, as *assertions.Store, a *alerts.Dispatcher, l *slog.Logger) *Scheduler {
	return &Scheduler{
		Endpoints: e, Checks: c, Assertions: as, Alerts: a, Logger: l,
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

	// Load assertions fresh so edits take effect without restarting the worker.
	asserts, err := w.s.Assertions.ListForEndpoint(ctx, w.ep.ID)
	if err != nil {
		w.s.Logger.Error("load assertions failed", "err", err, "endpoint", w.ep.ID)
	}

	req, err := http.NewRequestWithContext(rctx, w.ep.Method, w.ep.URL, nil)
	if err != nil {
		w.record(ctx, nil, nil, false, err.Error(), nil)
		return
	}
	req.Header.Set("User-Agent", "pingdan/1.0")

	start := time.Now()
	resp, err := client.Do(req)
	latency := int(time.Since(start).Milliseconds())

	if err != nil {
		msg := err.Error()
		w.record(ctx, nil, &latency, false, msg, nil)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxBodyRead))

	// Always evaluate the expected status code, plus any user-defined assertions.
	ar := assertions.Response{
		StatusCode: resp.StatusCode,
		LatencyMs:  latency,
		Headers:    flattenHeaders(resp.Header),
		Body:       body,
	}

	statusOK := resp.StatusCode == w.ep.ExpectedStatus
	var failed []assertions.Result
	if !statusOK {
		failed = append(failed, assertions.Result{
			Source: assertions.SourceStatusCode, Comparison: assertions.CmpEquals,
			Target: strconv.Itoa(w.ep.ExpectedStatus), Actual: strconv.Itoa(resp.StatusCode), Passed: false,
		})
	}
	for _, a := range asserts {
		if res := a.Evaluate(ar); !res.Passed {
			failed = append(failed, res)
		}
	}

	ok := len(failed) == 0
	var errMsg string
	if !ok {
		if !statusOK {
			errMsg = "unexpected status"
		} else {
			errMsg = "assertion failed"
		}
	}
	w.record(ctx, &resp.StatusCode, &latency, ok, errMsg, failed)
}

func flattenHeaders(h http.Header) map[string]string {
	out := make(map[string]string, len(h))
	for k, v := range h {
		if len(v) > 0 {
			out[strings.ToLower(k)] = v[0]
		}
	}
	return out
}

func (w *worker) record(ctx context.Context, status, latency *int, ok bool, errMsg string, failed []assertions.Result) {
	now := time.Now().UTC()
	var errPtr *string
	if errMsg != "" {
		errPtr = &errMsg
	}
	var failedJSON json.RawMessage
	if len(failed) > 0 {
		if b, err := json.Marshal(failed); err == nil {
			failedJSON = b
		}
	}
	c := &checks.Check{
		EndpointID:       w.ep.ID,
		StatusCode:       status,
		LatencyMs:        latency,
		OK:               ok,
		Error:            errPtr,
		FailedAssertions: failedJSON,
		CheckedAt:        now,
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
