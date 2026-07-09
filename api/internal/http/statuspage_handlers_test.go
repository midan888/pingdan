package httpx

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"  Public API  ":     "public-api",
		"API / EU + US":      "api-eu-us",
		"Already-valid-123":  "already-valid-123",
		"--- surrounded ---": "surrounded",
		"symbols only !!!":   "symbols-only",
		"multiple    spaces": "multiple-spaces",
	}
	for in, want := range cases {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestIsUniqueViolation(t *testing.T) {
	if !isUniqueViolation(&pgconn.PgError{Code: "23505"}) {
		t.Fatal("isUniqueViolation(unique) = false, want true")
	}
	if isUniqueViolation(&pgconn.PgError{Code: "23503"}) {
		t.Fatal("isUniqueViolation(foreign key) = true, want false")
	}
	if isUniqueViolation(errors.New("users_email_key")) {
		t.Fatal("isUniqueViolation(plain error) = true, want false")
	}
}

func TestStatusPageCreateRejectsInvalidInputBeforeDatabase(t *testing.T) {
	h := &StatusPageHandlers{}

	cases := []struct {
		name string
		body string
		want string
	}{
		{"bad json", "{", "bad json"},
		{"blank title", `{"title":"   "}`, "title required"},
		{"no slug material", `{"title":"!!!"}`, "invalid slug"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/status-pages", strings.NewReader(tc.body))

			h.create(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", rec.Code)
			}
			if !strings.Contains(rec.Body.String(), tc.want) {
				t.Fatalf("body = %q, want %q", rec.Body.String(), tc.want)
			}
		})
	}
}
