package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNormalizeEmail(t *testing.T) {
	if got := normalizeEmail("  OPS@Example.COM  "); got != "ops@example.com" {
		t.Errorf("normalizeEmail() = %q, want ops@example.com", got)
	}
}

func TestRegisterRejectsInvalidInputBeforeDatabase(t *testing.T) {
	e := NewEmail(nil, NewJWT("secret", time.Hour))

	if _, err := e.Register(context.Background(), "not-an-email", "long-enough", "Ops"); err == nil {
		t.Fatal("Register() error = nil, want invalid email error")
	}

	if _, err := e.Register(context.Background(), "ops@example.com", "short", "Ops"); !errors.Is(err, ErrWeakPassword) {
		t.Fatalf("Register() error = %v, want ErrWeakPassword", err)
	}
}
