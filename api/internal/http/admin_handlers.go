package httpx

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AdminHandlers exposes cross-user aggregates for operators. Every route is
// gated on IsAdmin, so it must be registered inside the authenticated group.
type AdminHandlers struct {
	Pool *pgxpool.Pool
	// IsAdmin decides whether the authenticated account may use these routes.
	IsAdmin func(email string) bool
}

func (h *AdminHandlers) Routes(r chi.Router) {
	r.Route("/admin", func(r chi.Router) {
		r.Use(h.requireAdmin)
		r.Get("/stats", h.stats)
		r.Get("/users", h.users)
	})
}

func (h *AdminHandlers) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := UserFrom(r.Context())
		if u == nil || !h.IsAdmin(u.Email) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *AdminHandlers) stats(w http.ResponseWriter, r *http.Request) {
	var userCount, endpointCount int64
	if err := h.Pool.QueryRow(r.Context(), `SELECT count(*) FROM users`).Scan(&userCount); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if err := h.Pool.QueryRow(r.Context(), `SELECT count(*) FROM endpoints`).Scan(&endpointCount); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	WriteJSON(w, 200, map[string]int64{
		"userCount":     userCount,
		"endpointCount": endpointCount,
	})
}

type adminUser struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	Name          *string   `json:"name"`
	Provider      *string   `json:"provider"`
	CreatedAt     time.Time `json:"createdAt"`
	EndpointCount int64     `json:"endpointCount"`
}

func (h *AdminHandlers) users(w http.ResponseWriter, r *http.Request) {
	rows, err := h.Pool.Query(r.Context(), `
		SELECT u.id, u.email, u.name, u.provider, u.created_at, count(e.id)
		FROM users u
		LEFT JOIN endpoints e ON e.user_id = u.id
		GROUP BY u.id
		ORDER BY u.created_at DESC`)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	out := []adminUser{}
	for rows.Next() {
		var u adminUser
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Provider, &u.CreatedAt, &u.EndpointCount); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		out = append(out, u)
	}
	if rows.Err() != nil {
		http.Error(w, rows.Err().Error(), 500)
		return
	}
	WriteJSON(w, 200, out)
}
