package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/pingdan/api/internal/checks"
	"github.com/pingdan/api/internal/endpoints"
	"github.com/pingdan/api/internal/statuspages"
)

// isUniqueViolation reports whether err is a Postgres unique-constraint error.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// StatusPageHandlers serves the authenticated CRUD for managing status pages.
type StatusPageHandlers struct {
	Store     *statuspages.Store
	Endpoints *endpoints.Store
}

func (h *StatusPageHandlers) Routes(r chi.Router) {
	r.Get("/status-pages", h.list)
	r.Post("/status-pages", h.create)
	r.Get("/status-pages/{id}", h.get)
	r.Put("/status-pages/{id}", h.update)
	r.Delete("/status-pages/{id}", h.delete)
	r.Put("/status-pages/{id}/items", h.setItems)
}

var slugRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

type statusPageInput struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type statusPageItemInput struct {
	EndpointID  string  `json:"endpointId"`
	DisplayName *string `json:"displayName"`
}

func (h *StatusPageHandlers) list(w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	out, err := h.Store.List(r.Context(), u.ID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	WriteJSON(w, 200, out)
}

func (h *StatusPageHandlers) create(w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	var in statusPageInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		http.Error(w, "title required", 400)
		return
	}
	slug := slugify(in.Slug)
	if slug == "" {
		slug = slugify(in.Title)
	}
	if !slugRe.MatchString(slug) {
		http.Error(w, "invalid slug", 400)
		return
	}
	p := &statuspages.Page{UserID: u.ID, Slug: slug, Title: in.Title, Description: strings.TrimSpace(in.Description)}
	if err := h.Store.Create(r.Context(), p); err != nil {
		if isUniqueViolation(err) {
			http.Error(w, "slug already taken", 409)
			return
		}
		http.Error(w, err.Error(), 500)
		return
	}
	WriteJSON(w, 201, p)
}

func (h *StatusPageHandlers) get(w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	id := chi.URLParam(r, "id")
	p, err := h.Store.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if p == nil || p.UserID != u.ID {
		http.Error(w, "not found", 404)
		return
	}
	items, err := h.Store.Items(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	WriteJSON(w, 200, map[string]any{"page": p, "items": items})
}

func (h *StatusPageHandlers) update(w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	id := chi.URLParam(r, "id")
	existing, err := h.Store.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if existing == nil || existing.UserID != u.ID {
		http.Error(w, "not found", 404)
		return
	}
	var in statusPageInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		http.Error(w, "title required", 400)
		return
	}
	slug := slugify(in.Slug)
	if slug == "" {
		slug = slugify(in.Title)
	}
	if !slugRe.MatchString(slug) {
		http.Error(w, "invalid slug", 400)
		return
	}
	if err := h.Store.Update(r.Context(), u.ID, id, slug, in.Title, strings.TrimSpace(in.Description)); err != nil {
		if isUniqueViolation(err) {
			http.Error(w, "slug already taken", 409)
			return
		}
		http.Error(w, err.Error(), 500)
		return
	}
	p, _ := h.Store.GetByID(r.Context(), id)
	WriteJSON(w, 200, p)
}

func (h *StatusPageHandlers) delete(w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	id := chi.URLParam(r, "id")
	if err := h.Store.Delete(r.Context(), u.ID, id); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(204)
}

func (h *StatusPageHandlers) setItems(w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	id := chi.URLParam(r, "id")
	p, err := h.Store.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if p == nil || p.UserID != u.ID {
		http.Error(w, "not found", 404)
		return
	}
	var in []statusPageItemInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	items := make([]statuspages.Item, 0, len(in))
	for _, it := range in {
		if strings.TrimSpace(it.EndpointID) == "" {
			continue
		}
		items = append(items, statuspages.Item{EndpointID: it.EndpointID, DisplayName: it.DisplayName})
	}
	if err := h.Store.SetItems(r.Context(), u.ID, id, items); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	saved, err := h.Store.Items(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	WriteJSON(w, 200, saved)
}

// --- Public, unauthenticated status page rendering ---

// PublicStatusHandlers serves the read-only public status page. It deliberately
// exposes only state, uptime %, and a recent state timeline — never URLs,
// status codes, latency, or assertion details.
type PublicStatusHandlers struct {
	Store     *statuspages.Store
	Endpoints *endpoints.Store
	Checks    *checks.Store
}

func (h *PublicStatusHandlers) Routes(r chi.Router) {
	r.Get("/public/status/{slug}", h.show)
}

// publicTick is a single check reduced to what is safe to show publicly.
type publicTick struct {
	OK        bool      `json:"ok"`
	CheckedAt time.Time `json:"checkedAt"`
}

type publicItem struct {
	Name      string       `json:"name"`
	State     string       `json:"state"`
	UptimePct float64      `json:"uptimePct"`
	History   []publicTick `json:"history"`
}

type publicPage struct {
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Overall     string       `json:"overall"`
	UpdatedAt   time.Time    `json:"updatedAt"`
	Items       []publicItem `json:"items"`
}

func (h *PublicStatusHandlers) show(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	p, err := h.Store.GetBySlug(r.Context(), slug)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if p == nil {
		http.Error(w, "not found", 404)
		return
	}
	items, err := h.Store.Items(r.Context(), p.ID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	since := time.Now().Add(-90 * 24 * time.Hour)
	out := publicPage{Title: p.Title, Description: p.Description, UpdatedAt: time.Now(), Overall: "up", Items: []publicItem{}}
	anyDown := false
	anyUnknown := false

	for _, it := range items {
		ep, err := h.Endpoints.GetByID(r.Context(), it.EndpointID)
		if err != nil || ep == nil {
			continue
		}
		name := ep.Name
		if it.DisplayName != nil && *it.DisplayName != "" {
			name = *it.DisplayName
		}

		st, err := h.Checks.StatsSince(r.Context(), ep.ID, since)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		// Recent timeline: latest 60 checks, oldest-first for rendering.
		recent, err := h.Checks.Recent(r.Context(), ep.ID, 60)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		hist := make([]publicTick, 0, len(recent))
		for i := len(recent) - 1; i >= 0; i-- {
			hist = append(hist, publicTick{OK: recent[i].OK, CheckedAt: recent[i].CheckedAt})
		}

		state := ep.CurrentState
		switch state {
		case "down":
			anyDown = true
		case "up":
			// no-op
		default:
			anyUnknown = true
		}

		out.Items = append(out.Items, publicItem{
			Name:      name,
			State:     state,
			UptimePct: st.UptimePct,
			History:   hist,
		})
	}

	if anyDown {
		out.Overall = "down"
	} else if anyUnknown && len(out.Items) > 0 {
		out.Overall = "degraded"
	} else if len(out.Items) == 0 {
		out.Overall = "unknown"
	}

	w.Header().Set("Cache-Control", "public, max-age=30")
	WriteJSON(w, 200, out)
}
