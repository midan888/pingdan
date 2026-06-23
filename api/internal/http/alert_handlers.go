package httpx

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pingdan/api/internal/alerts"
)

type AlertHandlers struct {
	Pool       *pgxpool.Pool
	Dispatcher *alerts.Dispatcher
}

func (h *AlertHandlers) Routes(r chi.Router) {
	r.Get("/alert-channels", h.list)
	r.Post("/alert-channels", h.create)
	r.Post("/alert-channels/test", h.test)
	r.Delete("/alert-channels/{id}", h.delete)
	r.Post("/endpoints/{id}/channels/{channelId}", h.attach)
	r.Delete("/endpoints/{id}/channels/{channelId}", h.detach)
}

type alertChannel struct {
	ID     string          `json:"id"`
	Kind   string          `json:"kind"`
	Label  string          `json:"label"`
	Config json.RawMessage `json:"config"`
}

func (h *AlertHandlers) list(w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	rows, err := h.Pool.Query(r.Context(), `SELECT id, kind, label, config FROM alert_channels WHERE user_id=$1 ORDER BY created_at DESC`, u.ID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	out := []alertChannel{}
	for rows.Next() {
		var a alertChannel
		if err := rows.Scan(&a.ID, &a.Kind, &a.Label, &a.Config); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		out = append(out, a)
	}
	WriteJSON(w, 200, out)
}

func (h *AlertHandlers) create(w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	var in alertChannel
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	if in.Kind != "email" && in.Kind != "telegram" {
		http.Error(w, "kind must be email|telegram", 400)
		return
	}
	if in.Label == "" || len(in.Config) == 0 {
		http.Error(w, "label and config required", 400)
		return
	}
	err := h.Pool.QueryRow(r.Context(), `
		INSERT INTO alert_channels (user_id, kind, label, config) VALUES ($1, $2, $3, $4) RETURNING id
	`, u.ID, in.Kind, in.Label, in.Config).Scan(&in.ID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	WriteJSON(w, 201, in)
}

// test sends a test notification to the given channel config without persisting it,
// so users can verify a channel works from the create/edit form before (or after) saving.
func (h *AlertHandlers) test(w http.ResponseWriter, r *http.Request) {
	var in alertChannel
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	if in.Kind != "email" && in.Kind != "telegram" {
		http.Error(w, "kind must be email|telegram", 400)
		return
	}
	if len(in.Config) == 0 {
		http.Error(w, "config required", 400)
		return
	}
	if err := h.Dispatcher.SendTest(r.Context(), in.Kind, in.Config); err != nil {
		http.Error(w, err.Error(), 502)
		return
	}
	w.WriteHeader(204)
}

func (h *AlertHandlers) delete(w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	id := chi.URLParam(r, "id")
	_, err := h.Pool.Exec(r.Context(), `DELETE FROM alert_channels WHERE id=$1 AND user_id=$2`, id, u.ID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(204)
}

func (h *AlertHandlers) attach(w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	endpointID := chi.URLParam(r, "id")
	channelID := chi.URLParam(r, "channelId")
	// confirm both belong to the user
	var ok bool
	err := h.Pool.QueryRow(r.Context(), `
		SELECT EXISTS(SELECT 1 FROM endpoints WHERE id=$1 AND user_id=$2)
		   AND EXISTS(SELECT 1 FROM alert_channels WHERE id=$3 AND user_id=$2)
	`, endpointID, u.ID, channelID).Scan(&ok)
	if err != nil || !ok {
		http.Error(w, "not found", 404)
		return
	}
	_, err = h.Pool.Exec(r.Context(), `
		INSERT INTO endpoint_alert_channels (endpoint_id, channel_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, endpointID, channelID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(204)
}

func (h *AlertHandlers) detach(w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	endpointID := chi.URLParam(r, "id")
	channelID := chi.URLParam(r, "channelId")
	_, err := h.Pool.Exec(r.Context(), `
		DELETE FROM endpoint_alert_channels eac
		USING endpoints e
		WHERE eac.endpoint_id = e.id AND e.user_id = $1
		  AND eac.endpoint_id = $2 AND eac.channel_id = $3
	`, u.ID, endpointID, channelID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(204)
}
