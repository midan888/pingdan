package alerts

import (
	"strings"
	"testing"
	"time"

	"github.com/pingdan/api/internal/checks"
	"github.com/pingdan/api/internal/endpoints"
)

func TestIsValidKind(t *testing.T) {
	for _, kind := range ValidKinds {
		if !IsValidKind(kind) {
			t.Fatalf("IsValidKind(%q) = false, want true", kind)
		}
	}
	if IsValidKind("fax") {
		t.Fatal("IsValidKind(fax) = true, want false")
	}
}

func TestKindCapabilitiesRequiresEnvForEnvBackedKinds(t *testing.T) {
	d := &Dispatcher{}
	got := d.KindCapabilities()

	for _, kind := range []string{KindSlack, KindDiscord, KindTeams, KindWebhook, KindPagerDuty, KindNtfy, KindOpsgenie} {
		if !got[kind] {
			t.Fatalf("%s capability = false, want always available", kind)
		}
	}
	for _, kind := range []string{KindEmail, KindTelegram, KindPushover, KindTwilioSMS} {
		if got[kind] {
			t.Fatalf("%s capability = true, want false without env", kind)
		}
	}

	d.ResendAPIKey = "resend-key"
	d.TelegramBotToken = "telegram-token"
	d.PushoverAppToken = "pushover-token"
	d.TwilioAccountSID = "AC123"
	d.TwilioAuthToken = "twilio-token"
	d.TwilioFrom = "+15550000000"
	got = d.KindCapabilities()

	for _, kind := range []string{KindEmail, KindTelegram, KindPushover, KindTwilioSMS} {
		if !got[kind] {
			t.Fatalf("%s capability = false, want true with env", kind)
		}
	}
}

func TestRenderMessageIncludesStatusOrError(t *testing.T) {
	statusCode := 503
	errMsg := "connection refused"
	checkedAt := time.Date(2026, 7, 10, 10, 30, 0, 0, time.UTC)
	e := endpoints.Endpoint{Name: "API", URL: "https://example.com"}

	subject, body := renderMessage(e, "down", &checks.Check{StatusCode: &statusCode, CheckedAt: checkedAt})
	if subject != "[pingdan] API — DOWN" {
		t.Fatalf("subject = %q, want down subject", subject)
	}
	if !strings.Contains(body, "Status: 503") || !strings.Contains(body, checkedAt.Format(time.RFC3339)) {
		t.Fatalf("body = %q, want status and timestamp", body)
	}

	subject, body = renderMessage(e, "up", &checks.Check{Error: &errMsg, CheckedAt: checkedAt})
	if subject != "[pingdan] API — RECOVERED" {
		t.Fatalf("subject = %q, want recovered subject", subject)
	}
	if !strings.Contains(body, "Error: connection refused") || !strings.Contains(body, "State: RECOVERED") {
		t.Fatalf("body = %q, want error detail and recovered state", body)
	}
}

func TestRenderSSLMessageSingularAndPluralDays(t *testing.T) {
	expiresAt := time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC)
	e := endpoints.Endpoint{Name: "API", URL: "https://example.com"}

	subject, body := renderSSLMessage(e, 1, expiresAt)
	if !strings.Contains(subject, "1 day") {
		t.Fatalf("subject = %q, want singular day", subject)
	}
	if !strings.Contains(body, "Days left: 1") || !strings.Contains(body, expiresAt.Format(time.RFC1123)) {
		t.Fatalf("body = %q, want days and expiry", body)
	}

	subject, _ = renderSSLMessage(e, 12, expiresAt)
	if !strings.Contains(subject, "12 days") {
		t.Fatalf("subject = %q, want plural days", subject)
	}
}
