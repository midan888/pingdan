package pinger

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pingdan/api/internal/endpoints"
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

func TestHTTPProbeCapturesResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got != "pingdan/1.0" {
			t.Errorf("User-Agent = %q, want pingdan/1.0", got)
		}
		w.Header().Set("X-Probe", "ok")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("healthy"))
	}))
	defer srv.Close()

	result, err := (httpProbe{client: srv.Client()}).Run(context.Background(), endpoints.Endpoint{
		Method: http.MethodGet,
		URL:    srv.URL,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.StatusCode == nil || *result.StatusCode != http.StatusCreated {
		t.Fatalf("StatusCode = %v, want 201", result.StatusCode)
	}
	if result.Headers["x-probe"] != "ok" {
		t.Errorf("X-Probe = %q, want ok", result.Headers["x-probe"])
	}
	if string(result.Body) != "healthy" {
		t.Errorf("Body = %q, want healthy", result.Body)
	}
	if result.LatencyMs == nil {
		t.Fatal("LatencyMs = nil")
	}
}

func TestTCPProbeConnectsAndReportsFailure(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	accepted := make(chan struct{})
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr == nil {
			_ = conn.Close()
			close(accepted)
		}
	}()

	target := "tcp://" + listener.Addr().String()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	result, err := (tcpProbe{}).Run(ctx, endpoints.Endpoint{URL: target})
	if err != nil {
		t.Fatalf("Run(%q) error = %v", target, err)
	}
	if result.LatencyMs == nil {
		t.Fatal("LatencyMs = nil")
	}
	select {
	case <-accepted:
	case <-time.After(time.Second):
		t.Fatal("listener did not accept the probe connection")
	}

	_ = listener.Close()
	_, err = (tcpProbe{}).Run(ctx, endpoints.Endpoint{URL: target})
	if err == nil || !strings.Contains(err.Error(), "TCP connect") {
		t.Fatalf("closed-port error = %v, want TCP connect error", err)
	}
}

func TestResolvePingIP(t *testing.T) {
	ip, err := resolvePingIP(context.Background(), "127.0.0.1")
	if err != nil {
		t.Fatalf("resolvePingIP() error = %v", err)
	}
	if got := ip.String(); got != "127.0.0.1" {
		t.Fatalf("resolved IP = %q, want 127.0.0.1", got)
	}
}
