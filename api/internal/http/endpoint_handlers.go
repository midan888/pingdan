package httpx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pingdan/api/internal/assertions"
	"github.com/pingdan/api/internal/checks"
	"github.com/pingdan/api/internal/endpoints"
	"github.com/pingdan/api/internal/pinger"
	"github.com/pingdan/api/internal/sslcheck"
)

type EndpointHandlers struct {
	Store      *endpoints.Store
	Checks     *checks.Store
	Assertions *assertions.Store
	Scheduler  *pinger.Scheduler
	SSL        *sslcheck.Checker
	Pool       *pgxpool.Pool
}

// channelIDs returns the alert channel IDs attached to an endpoint.
func (h *EndpointHandlers) channelIDs(ctx context.Context, endpointID string) ([]string, error) {
	rows, err := h.Pool.Query(ctx, `SELECT channel_id FROM endpoint_alert_channels WHERE endpoint_id=$1`, endpointID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// replaceChannels swaps the set of alert channels attached to an endpoint.
// Only channels owned by the same user are accepted; unknown IDs are ignored.
func (h *EndpointHandlers) replaceChannels(ctx context.Context, userID, endpointID string, channelIDs []string) error {
	tx, err := h.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM endpoint_alert_channels WHERE endpoint_id=$1`, endpointID); err != nil {
		return err
	}
	for _, cid := range channelIDs {
		// INSERT...SELECT guards ownership: nothing is inserted unless the channel belongs to the user.
		if _, err := tx.Exec(ctx, `
			INSERT INTO endpoint_alert_channels (endpoint_id, channel_id)
			SELECT $1, id FROM alert_channels WHERE id=$2 AND user_id=$3
			ON CONFLICT DO NOTHING
		`, endpointID, cid, userID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (h *EndpointHandlers) Routes(r chi.Router) {
	r.Get("/endpoints", h.list)
	r.Post("/endpoints", h.create)
	r.Get("/endpoints/{id}", h.get)
	r.Put("/endpoints/{id}", h.update)
	r.Delete("/endpoints/{id}", h.delete)
	r.Get("/endpoints/{id}/checks", h.listChecks)
	r.Get("/endpoints/{id}/stats", h.stats)
	r.Post("/endpoints/{id}/ssl-check", h.sslCheck)
}

// sslCheck runs an on-demand TLS certificate check and returns the refreshed
// endpoint with its updated ssl fields.
func (h *EndpointHandlers) sslCheck(w http.ResponseWriter, r *http.Request) {
	e := h.owned(w, r)
	if e == nil {
		return
	}
	if h.SSL != nil {
		h.SSL.CheckEndpoint(r.Context(), *e)
	}
	updated, err := h.Store.GetByID(r.Context(), e.ID)
	if err != nil || updated == nil {
		http.Error(w, "not found", 404)
		return
	}
	WriteJSON(w, 200, updated)
}

// owned loads an endpoint and verifies it belongs to the requesting user.
// Writes a 404 and returns nil if not found or not owned.
func (h *EndpointHandlers) owned(w http.ResponseWriter, r *http.Request) *endpoints.Endpoint {
	u := UserFrom(r.Context())
	id := chi.URLParam(r, "id")
	e, err := h.Store.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return nil
	}
	if e == nil || e.UserID != u.ID {
		http.Error(w, "not found", 404)
		return nil
	}
	return e
}

func (h *EndpointHandlers) list(w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	out, err := h.Store.List(r.Context(), u.ID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	WriteJSON(w, 200, out)
}

// get returns one endpoint plus its assertions.
func (h *EndpointHandlers) get(w http.ResponseWriter, r *http.Request) {
	e := h.owned(w, r)
	if e == nil {
		return
	}
	as, err := h.Assertions.ListForEndpoint(r.Context(), e.ID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	chans, err := h.channelIDs(r.Context(), e.ID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	WriteJSON(w, 200, map[string]any{"endpoint": e, "assertions": as, "channelIds": chans})
}

// listChecks returns the most recent checks for an endpoint (newest first).
func (h *EndpointHandlers) listChecks(w http.ResponseWriter, r *http.Request) {
	e := h.owned(w, r)
	if e == nil {
		return
	}
	limit := 100
	if q := r.URL.Query().Get("limit"); q != "" {
		if n, err := strconv.Atoi(q); err == nil {
			limit = n
		}
	}
	// When a time window is given, filter checks to that window so the page's
	// charts/timeline/table track the selected range (matching the stats handler).
	if q := r.URL.Query().Get("hours"); q != "" {
		if n, err := strconv.Atoi(q); err == nil && n > 0 && n <= 24*30 {
			since := time.Now().Add(-time.Duration(n) * time.Hour)
			out, err := h.Checks.RecentSince(r.Context(), e.ID, since, limit)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			WriteJSON(w, 200, out)
			return
		}
	}
	out, err := h.Checks.Recent(r.Context(), e.ID, limit)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	WriteJSON(w, 200, out)
}

// stats returns aggregate uptime/latency stats over a time window (default 24h).
func (h *EndpointHandlers) stats(w http.ResponseWriter, r *http.Request) {
	e := h.owned(w, r)
	if e == nil {
		return
	}
	window := 24 * time.Hour
	if q := r.URL.Query().Get("hours"); q != "" {
		if n, err := strconv.Atoi(q); err == nil && n > 0 && n <= 24*30 {
			window = time.Duration(n) * time.Hour
		}
	}
	since := time.Now().Add(-window)
	st, err := h.Checks.StatsSince(r.Context(), e.ID, since)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	WriteJSON(w, 200, st)
}

type assertionInput struct {
	Source     string `json:"source"`
	Property   string `json:"property"`
	Comparison string `json:"comparison"`
	Target     string `json:"target"`
}

type endpointInput struct {
	GroupID          *string          `json:"groupId"`
	Name             string           `json:"name"`
	CheckType        string           `json:"checkType"`
	URL              string           `json:"url"`
	Method           string           `json:"method"`
	ExpectedStatus   int              `json:"expectedStatus"`
	IntervalSec      int              `json:"intervalSec"`
	TimeoutSec       int              `json:"timeoutSec"`
	FailureThreshold int              `json:"failureThreshold"`
	Enabled          *bool            `json:"enabled"`
	Assertions       []assertionInput `json:"assertions"`
	ChannelIDs       []string         `json:"channelIds"`
}

// toAssertions converts and validates the assertion inputs.
func (in *endpointInput) toAssertions() ([]assertions.Assertion, error) {
	out := make([]assertions.Assertion, 0, len(in.Assertions))
	for _, a := range in.Assertions {
		as := assertions.Assertion{
			Source:     strings.TrimSpace(a.Source),
			Property:   strings.TrimSpace(a.Property),
			Comparison: strings.TrimSpace(a.Comparison),
			Target:     a.Target,
		}
		if err := as.Validate(); err != nil {
			return nil, errBadRequest(err.Error())
		}
		out = append(out, as)
	}
	return out, nil
}

// Supported check interval bounds (seconds).
const (
	minIntervalSec = 60            // 1 minute
	maxIntervalSec = 7 * 24 * 3600 // 7 days
)

// clampInterval bounds an arbitrary interval to the supported range,
// rounding down to a whole minute.
func clampInterval(sec int) int {
	if sec < minIntervalSec {
		return minIntervalSec
	}
	if sec > maxIntervalSec {
		return maxIntervalSec
	}
	return sec - sec%60
}

func (in *endpointInput) normalize() {
	// Treat an empty/blank group id as ungrouped.
	if in.GroupID != nil && strings.TrimSpace(*in.GroupID) == "" {
		in.GroupID = nil
	}
	in.Name = strings.TrimSpace(in.Name)
	in.CheckType = strings.ToLower(strings.TrimSpace(in.CheckType))
	if in.CheckType == "" {
		in.CheckType = endpoints.CheckTypeHTTP
	}
	in.URL = strings.TrimSpace(in.URL)
	if in.Method == "" {
		in.Method = "GET"
	}
	in.Method = strings.ToUpper(in.Method)
	if in.ExpectedStatus == 0 {
		in.ExpectedStatus = 200
	}
	in.IntervalSec = clampInterval(in.IntervalSec)
	if in.TimeoutSec <= 0 {
		in.TimeoutSec = 10
	}
	if in.FailureThreshold <= 0 {
		in.FailureThreshold = 2
	}
}

func (in *endpointInput) validate() error {
	if in.Name == "" {
		return errBadRequest("name required")
	}
	if len(in.Assertions) > 0 && in.CheckType != endpoints.CheckTypeHTTP {
		return errBadRequest("assertions are only supported for HTTP checks")
	}

	u, err := url.Parse(in.URL)
	if err != nil {
		return errBadRequest("invalid target")
	}
	switch in.CheckType {
	case endpoints.CheckTypeHTTP:
		if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return errBadRequest("HTTP target must use http:// or https://")
		}
	case endpoints.CheckTypeTCP:
		if u.Scheme != "tcp" || u.Hostname() == "" || u.Port() == "" || !plainHostTarget(u) {
			return errBadRequest("TCP target must look like tcp://host:port")
		}
		port, err := strconv.Atoi(u.Port())
		if err != nil || port < 1 || port > 65535 {
			return errBadRequest("TCP port must be between 1 and 65535")
		}
	case endpoints.CheckTypeICMP:
		if u.Scheme != "icmp" || u.Hostname() == "" || u.Port() != "" || !plainHostTarget(u) {
			return errBadRequest("ICMP target must look like icmp://host")
		}
	default:
		return errBadRequest("checkType must be http, tcp, or icmp")
	}
	return nil
}

// plainHostTarget keeps TCP and ICMP targets unambiguous: credentials, paths,
// queries, and fragments are not meaningful to either probe.
func plainHostTarget(u *url.URL) bool {
	return u.User == nil && (u.Path == "" || u.Path == "/") && u.RawQuery == "" && u.Fragment == ""
}

func (h *EndpointHandlers) create(w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	var in endpointInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	in.normalize()
	if err := in.validate(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	asserts, err := in.toAssertions()
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	enabled := true
	if in.Enabled != nil {
		enabled = *in.Enabled
	}
	e := &endpoints.Endpoint{
		UserID: u.ID, GroupID: in.GroupID, Name: in.Name, CheckType: in.CheckType, URL: in.URL, Method: in.Method,
		ExpectedStatus: in.ExpectedStatus, IntervalSec: in.IntervalSec, TimeoutSec: in.TimeoutSec,
		FailureThreshold: in.FailureThreshold, Enabled: enabled,
	}
	if err := h.Store.Create(r.Context(), e); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if err := h.Assertions.Replace(r.Context(), e.ID, asserts); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if err := h.replaceChannels(r.Context(), u.ID, e.ID, in.ChannelIDs); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if e.Enabled {
		h.Scheduler.Upsert(*e)
	}
	WriteJSON(w, 201, map[string]any{"endpoint": e, "assertions": asserts, "channelIds": in.ChannelIDs})
}

func (h *EndpointHandlers) update(w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	id := chi.URLParam(r, "id")
	var in endpointInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	in.normalize()
	if err := in.validate(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	asserts, err := in.toAssertions()
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	enabled := true
	if in.Enabled != nil {
		enabled = *in.Enabled
	}
	e := &endpoints.Endpoint{
		GroupID: in.GroupID, Name: in.Name, CheckType: in.CheckType, URL: in.URL, Method: in.Method,
		ExpectedStatus: in.ExpectedStatus, IntervalSec: in.IntervalSec, TimeoutSec: in.TimeoutSec,
		FailureThreshold: in.FailureThreshold, Enabled: enabled,
	}
	if err := h.Store.Update(r.Context(), u.ID, id, e); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	updated, err := h.Store.GetByID(r.Context(), id)
	if err != nil || updated == nil || updated.UserID != u.ID {
		http.Error(w, "not found", 404)
		return
	}
	if err := h.Assertions.Replace(r.Context(), updated.ID, asserts); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if err := h.replaceChannels(r.Context(), u.ID, updated.ID, in.ChannelIDs); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if updated.Enabled {
		h.Scheduler.Upsert(*updated)
	} else {
		h.Scheduler.Remove(updated.ID)
	}
	WriteJSON(w, 200, map[string]any{"endpoint": updated, "assertions": asserts, "channelIds": in.ChannelIDs})
}

func (h *EndpointHandlers) delete(w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	id := chi.URLParam(r, "id")
	if err := h.Store.Delete(r.Context(), u.ID, id); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	h.Scheduler.Remove(id)
	w.WriteHeader(204)
}

type httpError struct {
	code int
	msg  string
}

func (e *httpError) Error() string { return e.msg }

func errBadRequest(msg string) error { return &httpError{code: 400, msg: msg} }

func WriteJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
