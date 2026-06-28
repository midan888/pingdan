package httpx

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/pingdan/api/internal/groups"
)

type GroupHandlers struct {
	Store *groups.Store
}

func (h *GroupHandlers) Routes(r chi.Router) {
	r.Get("/groups", h.list)
	r.Post("/groups", h.create)
	r.Put("/groups/{id}", h.update)
	r.Delete("/groups/{id}", h.delete)
}

type groupInput struct {
	Name string `json:"name"`
}

func (h *GroupHandlers) list(w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	out, err := h.Store.List(r.Context(), u.ID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	WriteJSON(w, 200, out)
}

func (h *GroupHandlers) create(w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	var in groupInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		http.Error(w, "name required", 400)
		return
	}
	g := &groups.Group{UserID: u.ID, Name: in.Name}
	if err := h.Store.Create(r.Context(), g); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	WriteJSON(w, 201, g)
}

func (h *GroupHandlers) update(w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	id := chi.URLParam(r, "id")
	var in groupInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		http.Error(w, "name required", 400)
		return
	}
	if err := h.Store.Update(r.Context(), u.ID, id, in.Name); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	g, err := h.Store.GetByID(r.Context(), id)
	if err != nil || g == nil || g.UserID != u.ID {
		http.Error(w, "not found", 404)
		return
	}
	WriteJSON(w, 200, g)
}

func (h *GroupHandlers) delete(w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	id := chi.URLParam(r, "id")
	if err := h.Store.Delete(r.Context(), u.ID, id); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(204)
}
