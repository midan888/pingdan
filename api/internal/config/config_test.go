package config

import (
	"testing"
	"time"
)

func TestLoadRequiresDatabaseAndJWTSecret(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("JWT_SECRET", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want missing DATABASE_URL error")
	}

	t.Setenv("DATABASE_URL", "postgres://example")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want missing JWT_SECRET error")
	}
}

func TestLoadDefaultsAndOptionalAlertEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("JWT_TTL", "2h")
	t.Setenv("PUBLIC_URL", "")
	t.Setenv("FRONTEND_URL", "")
	t.Setenv("EMAIL_FROM", "")
	t.Setenv("RESEND_API_KEY", "resend-key")
	t.Setenv("TELEGRAM_BOT_TOKEN", "telegram-token")
	t.Setenv("PUSHOVER_APP_TOKEN", "pushover-token")
	t.Setenv("TWILIO_ACCOUNT_SID", "AC123")
	t.Setenv("TWILIO_AUTH_TOKEN", "twilio-token")
	t.Setenv("TWILIO_FROM", "+15550000000")

	c, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if c.HTTPAddr != ":8080" {
		t.Errorf("HTTPAddr = %q, want default :8080", c.HTTPAddr)
	}
	if c.JWTTTL != 2*time.Hour {
		t.Errorf("JWTTTL = %v, want 2h", c.JWTTTL)
	}
	if c.PublicURL != "http://localhost:8080" || c.FrontendURL != "http://localhost:3000" {
		t.Errorf("public/frontend URL = %q/%q, want defaults", c.PublicURL, c.FrontendURL)
	}
	if c.EmailFrom != "alerts@pingdan.local" {
		t.Errorf("EmailFrom = %q, want default sender", c.EmailFrom)
	}
	if c.ResendAPIKey != "resend-key" || c.TelegramBotToken != "telegram-token" || c.PushoverAppToken != "pushover-token" {
		t.Errorf("alert envs not loaded: %#v", c)
	}
	if c.TwilioAccountSID != "AC123" || c.TwilioAuthToken != "twilio-token" || c.TwilioFrom != "+15550000000" {
		t.Errorf("twilio envs = %q/%q/%q, want configured values", c.TwilioAccountSID, c.TwilioAuthToken, c.TwilioFrom)
	}
}

func TestDurationEnvFallsBackOnInvalidValue(t *testing.T) {
	t.Setenv("PINGDAN_TEST_DURATION", "not-a-duration")

	if got := durationEnv("PINGDAN_TEST_DURATION", 30*time.Second); got != 30*time.Second {
		t.Errorf("durationEnv() = %v, want fallback", got)
	}
}
