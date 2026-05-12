package httpx

import (
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pingdan/api/internal/auth"
)

type AuthHandlers struct {
	OAuth       *auth.OAuth
	FrontendURL string
}

func (h *AuthHandlers) Routes(r chi.Router) {
	r.Get("/auth/{provider}/start", h.start)
	r.Get("/auth/{provider}/callback", h.callback)
}

func (h *AuthHandlers) start(w http.ResponseWriter, r *http.Request) {
	p := auth.Provider(chi.URLParam(r, "provider"))
	cfg, ok := h.OAuth.Config(p)
	if !ok {
		http.Error(w, "provider not configured", http.StatusBadRequest)
		return
	}
	state := auth.RandomState()
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(10 * time.Minute),
	})
	http.Redirect(w, r, cfg.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

func (h *AuthHandlers) callback(w http.ResponseWriter, r *http.Request) {
	p := auth.Provider(chi.URLParam(r, "provider"))
	cfg, ok := h.OAuth.Config(p)
	if !ok {
		http.Error(w, "provider not configured", http.StatusBadRequest)
		return
	}
	cookie, err := r.Cookie("oauth_state")
	if err != nil || cookie.Value == "" || cookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}
	code := r.URL.Query().Get("code")
	tok, err := cfg.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "exchange failed", http.StatusBadRequest)
		return
	}
	pu, err := h.OAuth.FetchUser(r.Context(), p, tok)
	if err != nil {
		http.Error(w, "userinfo failed", http.StatusBadGateway)
		return
	}
	jwtTok, err := h.OAuth.UpsertAndIssue(r.Context(), p, pu)
	if err != nil {
		http.Error(w, "issue failed", http.StatusInternalServerError)
		return
	}
	redirect := h.FrontendURL + "/auth/callback?token=" + url.QueryEscape(jwtTok)
	http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
}
