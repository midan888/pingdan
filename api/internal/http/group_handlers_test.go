package httpx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGroupCreateRejectsInvalidInputBeforeDatabase(t *testing.T) {
	h := &GroupHandlers{}

	cases := []struct {
		name string
		body string
		want string
	}{
		{"bad json", "{", "bad json"},
		{"blank name", `{"name":"   "}`, "name required"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/groups", strings.NewReader(tc.body))

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

func TestGroupUpdateRejectsInvalidInputBeforeDatabase(t *testing.T) {
	h := &GroupHandlers{}

	for _, body := range []string{"{", `{"name":""}`} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/groups/group_123", strings.NewReader(body))

		h.update(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400 for body %q", rec.Code, body)
		}
	}
}
