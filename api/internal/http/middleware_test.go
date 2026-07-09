package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pingdan/api/internal/auth"
)

func TestAuthMiddlewareRejectsMissingAndInvalidToken(t *testing.T) {
	j := auth.NewJWT("secret", time.Hour)
	handler := AuthMiddleware(j)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for _, tc := range []struct {
		name string
		auth string
	}{
		{"missing", ""},
		{"invalid", "Bearer nope"},
		{"wrong scheme", "Token nope"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.auth != "" {
				req.Header.Set("Authorization", tc.auth)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401", rec.Code)
			}
		})
	}
}

func TestAuthMiddlewareStoresUserInContext(t *testing.T) {
	j := auth.NewJWT("secret", time.Hour)
	token, err := j.Issue("user_123", "ops@example.com")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	handler := AuthMiddleware(j)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := UserFrom(r.Context())
		if u == nil {
			t.Fatal("UserFrom() = nil, want authenticated user")
		}
		if u.ID != "user_123" || u.Email != "ops@example.com" {
			t.Fatalf("user = %#v, want token claims", u)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
}
