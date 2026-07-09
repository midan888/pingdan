package statuspages

import "testing"

func TestNullIfEmpty(t *testing.T) {
	if got := nullIfEmpty(nil); got != nil {
		t.Fatalf("nullIfEmpty(nil) = %v, want nil", got)
	}

	empty := ""
	if got := nullIfEmpty(&empty); got != nil {
		t.Fatalf("nullIfEmpty(empty) = %v, want nil", got)
	}

	name := "Public API"
	if got := nullIfEmpty(&name); got == nil || *got != "Public API" {
		t.Fatalf("nullIfEmpty(name) = %v, want original value", got)
	}
}
