package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pingdan/api/internal/auth"
)

type AuthHandlers struct {
	OAuth       *auth.OAuth
	Email       *auth.Email
	JWT         *auth.JWT
	FrontendURL string
}

func (h *AuthHandlers) Routes(r chi.Router) {
	r.Get("/auth/{provider}/start", h.start)
	r.Get("/auth/{provider}/callback", h.callback)
	r.Post("/auth/email/register", h.emailRegister)
	r.Post("/auth/email/login", h.emailLogin)
}

// AuthedRoutes registers auth routes that require a valid token.
func (h *AuthHandlers) AuthedRoutes(r chi.Router) {
	r.Post("/auth/refresh", h.refresh)
}

// refresh re-issues a token for the already-authenticated user, giving the
// session a sliding expiry: any visit within the TTL extends it by a full TTL.
func (h *AuthHandlers) refresh(w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	tok, err := h.JWT.Issue(u.ID, u.Email)
	if err != nil {
		http.Error(w, "issue failed", http.StatusInternalServerError)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"token": tok})
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

type emailCreds struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func (h *AuthHandlers) emailRegister(w http.ResponseWriter, r *http.Request) {
	var in emailCreds
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	tok, err := h.Email.Register(r.Context(), in.Email, in.Password, in.Name)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrEmailTaken):
			http.Error(w, err.Error(), http.StatusConflict)
		case errors.Is(err, auth.ErrWeakPassword):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	WriteJSON(w, http.StatusCreated, map[string]string{"token": tok})
}

func (h *AuthHandlers) emailLogin(w http.ResponseWriter, r *http.Request) {
	var in emailCreds
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	tok, err := h.Email.Login(r.Context(), in.Email, in.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		http.Error(w, "login failed", http.StatusInternalServerError)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"token": tok})
}
