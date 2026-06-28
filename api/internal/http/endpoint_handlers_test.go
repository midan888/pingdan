package httpx

import "testing"

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
