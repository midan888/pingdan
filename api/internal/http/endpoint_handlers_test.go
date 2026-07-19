package httpx

import (
	"strings"
	"testing"

	"github.com/pingdan/api/internal/assertions"
	"github.com/pingdan/api/internal/endpoints"
)

func strptr(s string) *string { return &s }

// TestNormalizeGroupID covers how normalize() resolves the optional group id:
// a real id is kept verbatim, while nil/empty/whitespace all mean "ungrouped".
func TestNormalizeGroupID(t *testing.T) {
	cases := []struct {
		name string
		in   *string
		want *string
	}{
		{"nil stays nil", nil, nil},
		{"empty becomes nil", strptr(""), nil},
		{"whitespace becomes nil", strptr("   "), nil},
		{"real id preserved", strptr("a1b2c3"), strptr("a1b2c3")},
		{"untrimmed id preserved as-is", strptr("  a1b2c3  "), strptr("  a1b2c3  ")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			in := &endpointInput{Name: "x", URL: "https://e.com", GroupID: c.in}
			in.normalize()
			switch {
			case c.want == nil && in.GroupID != nil:
				t.Fatalf("want nil group id, got %q", *in.GroupID)
			case c.want != nil && in.GroupID == nil:
				t.Fatalf("want %q, got nil", *c.want)
			case c.want != nil && *in.GroupID != *c.want:
				t.Fatalf("want %q, got %q", *c.want, *in.GroupID)
			}
		})
	}
}

func TestClampInterval(t *testing.T) {
	cases := []struct {
		name string
		in   int
		want int
	}{
		{"too low", 1, minIntervalSec},
		{"one minute", 60, 60},
		{"rounds down to whole minute", 125, 120},
		{"too high", 9999999, maxIntervalSec},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := clampInterval(c.in); got != c.want {
				t.Fatalf("clampInterval(%d) = %d, want %d", c.in, got, c.want)
			}
		})
	}
}

func TestEndpointInputNormalizeDefaults(t *testing.T) {
	in := &endpointInput{
		Name:        "  API  ",
		URL:         "  https://example.com  ",
		IntervalSec: 125,
	}

	in.normalize()

	if in.Name != "API" {
		t.Errorf("Name = %q, want trimmed API", in.Name)
	}
	if in.CheckType != endpoints.CheckTypeHTTP {
		t.Errorf("CheckType = %q, want HTTP default", in.CheckType)
	}
	if in.URL != "https://example.com" {
		t.Errorf("URL = %q, want trimmed URL", in.URL)
	}
	if in.Method != "GET" {
		t.Errorf("Method = %q, want GET default", in.Method)
	}
	if in.ExpectedStatus != 200 {
		t.Errorf("ExpectedStatus = %d, want 200 default", in.ExpectedStatus)
	}
	if in.IntervalSec != 120 {
		t.Errorf("IntervalSec = %d, want rounded interval", in.IntervalSec)
	}
	if in.TimeoutSec != 10 {
		t.Errorf("TimeoutSec = %d, want 10 default", in.TimeoutSec)
	}
	if in.FailureThreshold != 2 {
		t.Errorf("FailureThreshold = %d, want 2 default", in.FailureThreshold)
	}
}

func TestEndpointInputValidate(t *testing.T) {
	cases := []struct {
		name    string
		in      endpointInput
		wantErr string
	}{
		{"missing name", endpointInput{URL: "https://example.com"}, "name required"},
		{"unknown check type", endpointInput{Name: "API", CheckType: "udp", URL: "udp://example.com:53"}, "checkType must be http, tcp, or icmp"},
		{"HTTP bad scheme", endpointInput{Name: "API", CheckType: endpoints.CheckTypeHTTP, URL: "ftp://example.com"}, "HTTP target must use http:// or https://"},
		{"HTTP missing host", endpointInput{Name: "API", CheckType: endpoints.CheckTypeHTTP, URL: "https://"}, "HTTP target must use http:// or https://"},
		{"valid HTTP", endpointInput{Name: "API", CheckType: endpoints.CheckTypeHTTP, URL: "http://example.com"}, ""},
		{"valid HTTPS", endpointInput{Name: "API", CheckType: endpoints.CheckTypeHTTP, URL: "https://example.com"}, ""},
		{"TCP missing port", endpointInput{Name: "VPN", CheckType: endpoints.CheckTypeTCP, URL: "tcp://vpn.example.com"}, "TCP target must look like tcp://host:port"},
		{"TCP port too high", endpointInput{Name: "VPN", CheckType: endpoints.CheckTypeTCP, URL: "tcp://vpn.example.com:65536"}, "TCP port must be between 1 and 65535"},
		{"TCP path rejected", endpointInput{Name: "VPN", CheckType: endpoints.CheckTypeTCP, URL: "tcp://vpn.example.com:443/path"}, "TCP target must look like tcp://host:port"},
		{"valid TCP", endpointInput{Name: "VPN", CheckType: endpoints.CheckTypeTCP, URL: "tcp://vpn.example.com:443"}, ""},
		{"valid TCP IPv6", endpointInput{Name: "VPN", CheckType: endpoints.CheckTypeTCP, URL: "tcp://[2001:db8::1]:443"}, ""},
		{"ICMP port rejected", endpointInput{Name: "VPN", CheckType: endpoints.CheckTypeICMP, URL: "icmp://vpn.example.com:7"}, "ICMP target must look like icmp://host"},
		{"ICMP path rejected", endpointInput{Name: "VPN", CheckType: endpoints.CheckTypeICMP, URL: "icmp://vpn.example.com/path"}, "ICMP target must look like icmp://host"},
		{"valid ICMP", endpointInput{Name: "VPN", CheckType: endpoints.CheckTypeICMP, URL: "icmp://vpn.example.com"}, ""},
		{"non-HTTP assertions rejected", endpointInput{Name: "VPN", CheckType: endpoints.CheckTypeTCP, URL: "tcp://vpn.example.com:443", Assertions: []assertionInput{{Source: assertions.SourceResponseTime}}}, "assertions are only supported for HTTP checks"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.in.validate()
			if c.wantErr == "" && err != nil {
				t.Fatalf("validate() error = %v, want nil", err)
			}
			if c.wantErr != "" {
				if err == nil {
					t.Fatalf("validate() error = nil, want %q", c.wantErr)
				}
				if !strings.Contains(err.Error(), c.wantErr) {
					t.Fatalf("validate() error = %q, want %q", err.Error(), c.wantErr)
				}
			}
		})
	}
}

func TestEndpointInputToAssertionsTrimsAndValidates(t *testing.T) {
	in := &endpointInput{Assertions: []assertionInput{
		{Source: " header ", Property: " Content-Type ", Comparison: " contains ", Target: "json"},
		{Source: assertions.SourceStatusCode, Comparison: assertions.CmpEquals, Target: "200"},
	}}

	got, err := in.toAssertions()
	if err != nil {
		t.Fatalf("toAssertions() error = %v", err)
	}
	if got[0].Source != assertions.SourceHeader || got[0].Property != "Content-Type" || got[0].Comparison != assertions.CmpContains {
		t.Fatalf("first assertion = %#v, want trimmed fields", got[0])
	}

	in = &endpointInput{Assertions: []assertionInput{{Source: "bogus", Comparison: assertions.CmpEquals, Target: "x"}}}
	if _, err := in.toAssertions(); err == nil {
		t.Fatal("toAssertions() error = nil, want validation error")
	}
}
