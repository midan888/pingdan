package pinger

import (
	"net/http"
	"testing"
)

func TestFlattenHeadersLowercasesAndTakesFirstValue(t *testing.T) {
	h := http.Header{}
	h.Add("Content-Type", "application/json")
	h.Add("Content-Type", "text/plain")
	h.Add("X-Trace-ID", "abc123")

	got := flattenHeaders(h)

	if got["content-type"] != "application/json" {
		t.Errorf("content-type = %q, want first header value", got["content-type"])
	}
	if got["x-trace-id"] != "abc123" {
		t.Errorf("x-trace-id = %q, want abc123", got["x-trace-id"])
	}
	if _, ok := got["Content-Type"]; ok {
		t.Fatal("flattenHeaders kept original case key")
	}
}
