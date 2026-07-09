package auth

import (
	"testing"
	"time"
)

func TestNewOAuthConfiguresOnlyEnabledProviders(t *testing.T) {
	o := NewOAuth(nil, NewJWT("secret", time.Hour), "https://api.example.com", "google-id", "google-secret", "", "")

	googleCfg, ok := o.Config(ProviderGoogle)
	if !ok {
		t.Fatal("google config missing")
	}
	if googleCfg.RedirectURL != "https://api.example.com/auth/google/callback" {
		t.Errorf("google redirect = %q, want public callback URL", googleCfg.RedirectURL)
	}
	if googleCfg.ClientID != "google-id" || googleCfg.ClientSecret != "google-secret" {
		t.Errorf("google credentials = %q/%q, want configured credentials", googleCfg.ClientID, googleCfg.ClientSecret)
	}

	if _, ok := o.Config(ProviderGitHub); ok {
		t.Fatal("github config present, want disabled provider to be absent")
	}
}

func TestRandomStateShape(t *testing.T) {
	a := RandomState()
	b := RandomState()

	if len(a) != 32 {
		t.Fatalf("RandomState() length = %d, want 32 hex chars", len(a))
	}
	if a == b {
		t.Fatal("RandomState() returned the same value twice")
	}
	for _, r := range a {
		if !(r >= '0' && r <= '9') && !(r >= 'a' && r <= 'f') {
			t.Fatalf("RandomState() = %q, contains non-hex rune %q", a, r)
		}
	}
}
