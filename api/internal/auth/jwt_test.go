package auth

import (
	"testing"
	"time"
)

func TestJWTRoundTrip(t *testing.T) {
	j := NewJWT("secret", time.Hour)

	token, err := j.Issue("user_123", "ops@example.com")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	claims, err := j.Parse(token)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if claims.UserID != "user_123" {
		t.Errorf("UserID = %q, want user_123", claims.UserID)
	}
	if claims.Email != "ops@example.com" {
		t.Errorf("Email = %q, want ops@example.com", claims.Email)
	}
	if claims.ExpiresAt == nil || time.Until(claims.ExpiresAt.Time) <= 0 {
		t.Errorf("ExpiresAt = %v, want a future expiry", claims.ExpiresAt)
	}
}

func TestJWTRejectsWrongSecret(t *testing.T) {
	token, err := NewJWT("right", time.Hour).Issue("user_123", "ops@example.com")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	if _, err := NewJWT("wrong", time.Hour).Parse(token); err == nil {
		t.Fatal("Parse() error = nil, want signature error")
	}
}

func TestJWTRejectsExpiredToken(t *testing.T) {
	token, err := NewJWT("secret", -time.Hour).Issue("user_123", "ops@example.com")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	if _, err := NewJWT("secret", time.Hour).Parse(token); err == nil {
		t.Fatal("Parse() error = nil, want expired token error")
	}
}
