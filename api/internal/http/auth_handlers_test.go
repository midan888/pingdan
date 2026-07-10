package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pingdan/api/internal/auth"
)

func TestAuthStartRejectsUnconfiguredProvider(t *testing.T) {
	h := &AuthHandlers{OAuth: auth.NewOAuth(nil, auth.NewJWT("secret", time.Hour), "https://api.example.com", "", "", "", "")}
	r := chi.NewRouter()
	h.Routes(r)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/github/start", nil)

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "provider not configured") {
		t.Fatalf("body = %q, want provider error", rec.Body.String())
	}
}

func TestAuthStartSetsStateCookieAndRedirects(t *testing.T) {
	h := &AuthHandlers{OAuth: auth.NewOAuth(nil, auth.NewJWT("secret", time.Hour), "https://api.example.com", "google-id", "google-secret", "", "")}
	r := chi.NewRouter()
	h.Routes(r)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/google/start", nil)

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("status = %d, want 307", rec.Code)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "oauth_state" || cookies[0].Value == "" {
		t.Fatalf("cookies = %#v, want oauth_state", cookies)
	}
	location := rec.Header().Get("Location")
	u, err := url.Parse(location)
	if err != nil {
		t.Fatalf("redirect location parse: %v", err)
	}
	if u.Query().Get("state") != cookies[0].Value {
		t.Fatalf("redirect state = %q, cookie = %q", u.Query().Get("state"), cookies[0].Value)
	}
	if u.Query().Get("redirect_uri") != "https://api.example.com/auth/google/callback" {
		t.Fatalf("redirect_uri = %q, want callback URL", u.Query().Get("redirect_uri"))
	}
}

func TestAuthRefreshIssuesFreshToken(t *testing.T) {
	j := auth.NewJWT("secret", time.Hour)
	h := &AuthHandlers{JWT: j}
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(j))
		h.AuthedRoutes(r)
	})

	tok, err := j.Issue("user-1", "u@example.com")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.Header.Set("Authorization", "Bearer "+tok)

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body = %q", rec.Code, rec.Body.String())
	}
	var out struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	claims, err := j.Parse(out.Token)
	if err != nil {
		t.Fatalf("parse issued token: %v", err)
	}
	if claims.UserID != "user-1" || claims.Email != "u@example.com" {
		t.Fatalf("claims = %+v, want original user", claims)
	}
}

func TestAuthRefreshRejectsMissingToken(t *testing.T) {
	j := auth.NewJWT("secret", time.Hour)
	h := &AuthHandlers{JWT: j}
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(j))
		h.AuthedRoutes(r)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestAuthCallbackRejectsMissingOrMismatchedState(t *testing.T) {
	h := &AuthHandlers{OAuth: auth.NewOAuth(nil, auth.NewJWT("secret", time.Hour), "https://api.example.com", "google-id", "google-secret", "", "")}
	r := chi.NewRouter()
	h.Routes(r)

	for _, tc := range []struct {
		name   string
		cookie *http.Cookie
	}{
		{"missing cookie", nil},
		{"mismatched cookie", &http.Cookie{Name: "oauth_state", Value: "cookie-state"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?state=query-state&code=abc", nil)
			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}

			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", rec.Code)
			}
			if !strings.Contains(rec.Body.String(), "invalid state") {
				t.Fatalf("body = %q, want invalid state", rec.Body.String())
			}
		})
	}
}
