package httpx

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/pingdan/api/internal/endpoints"
	"github.com/pingdan/api/internal/pinger"
)

type EndpointHandlers struct {
	Store     *endpoints.Store
	Scheduler *pinger.Scheduler
}

func (h *EndpointHandlers) Routes(r chi.Router) {
	r.Get("/endpoints", h.list)
	r.Post("/endpoints", h.create)
	r.Put("/endpoints/{id}", h.update)
	r.Delete("/endpoints/{id}", h.delete)
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

type endpointInput struct {
	Name             string `json:"name"`
	URL              string `json:"url"`
	Method           string `json:"method"`
	ExpectedStatus   int    `json:"expectedStatus"`
	IntervalSec      int    `json:"intervalSec"`
	TimeoutSec       int    `json:"timeoutSec"`
	FailureThreshold int    `json:"failureThreshold"`
	Enabled          *bool  `json:"enabled"`
}

func (in *endpointInput) normalize() {
	in.Name = strings.TrimSpace(in.Name)
	in.URL = strings.TrimSpace(in.URL)
	if in.Method == "" {
		in.Method = "GET"
	}
	in.Method = strings.ToUpper(in.Method)
	if in.ExpectedStatus == 0 {
		in.ExpectedStatus = 200
	}
	if in.IntervalSec < 10 {
		in.IntervalSec = 60
	}
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
	u, err := url.Parse(in.URL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return errBadRequest("url must be http(s)")
	}
	return nil
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
	enabled := true
	if in.Enabled != nil {
		enabled = *in.Enabled
	}
	e := &endpoints.Endpoint{
		UserID: u.ID, Name: in.Name, URL: in.URL, Method: in.Method,
		ExpectedStatus: in.ExpectedStatus, IntervalSec: in.IntervalSec, TimeoutSec: in.TimeoutSec,
		FailureThreshold: in.FailureThreshold, Enabled: enabled,
	}
	if err := h.Store.Create(r.Context(), e); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if e.Enabled {
		h.Scheduler.Upsert(*e)
	}
	WriteJSON(w, 201, e)
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
	enabled := true
	if in.Enabled != nil {
		enabled = *in.Enabled
	}
	e := &endpoints.Endpoint{
		Name: in.Name, URL: in.URL, Method: in.Method,
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
	if updated.Enabled {
		h.Scheduler.Upsert(*updated)
	} else {
		h.Scheduler.Remove(updated.ID)
	}
	WriteJSON(w, 200, updated)
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
