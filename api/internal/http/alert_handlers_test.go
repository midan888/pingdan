package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pingdan/api/internal/alerts"
)

func TestAlertCapabilitiesHandlerReflectsConfiguredEnvBackedKinds(t *testing.T) {
	h := &AlertHandlers{Dispatcher: &alerts.Dispatcher{
		ResendAPIKey:     "resend-key",
		TelegramBotToken: "",
		PushoverAppToken: "pushover-token",
		TwilioAccountSID: "AC123",
		TwilioAuthToken:  "twilio-token",
		TwilioFrom:       "+15550000000",
	}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/capabilities", nil)

	h.capabilities(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body struct {
		AlertChannelKinds map[string]bool `json:"alertChannelKinds"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if !body.AlertChannelKinds[alerts.KindEmail] {
		t.Fatal("email capability = false, want true")
	}
	if body.AlertChannelKinds[alerts.KindTelegram] {
		t.Fatal("telegram capability = true, want false without token")
	}
	if !body.AlertChannelKinds[alerts.KindWebhook] {
		t.Fatal("webhook capability = false, want always available")
	}
	if !body.AlertChannelKinds[alerts.KindTwilioSMS] {
		t.Fatal("twilio_sms capability = false, want true with all envs")
	}
}

func TestAlertTestRejectsInvalidInputBeforeDispatch(t *testing.T) {
	h := &AlertHandlers{Dispatcher: &alerts.Dispatcher{}}

	cases := []struct {
		name string
		body string
		want string
	}{
		{"bad json", "{", "bad json"},
		{"bad kind", `{"kind":"fax","config":{}}`, "kind must be"},
		{"missing config", `{"kind":"webhook"}`, "config required"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/alert-channels/test", strings.NewReader(tc.body))

			h.test(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", rec.Code)
			}
			if !strings.Contains(rec.Body.String(), tc.want) {
				t.Fatalf("body = %q, want %q", rec.Body.String(), tc.want)
			}
		})
	}
}

func TestAlertTestDispatchesAndReturnsNoContent(t *testing.T) {
	var got struct {
		Text string `json:"text"`
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	h := &AlertHandlers{Dispatcher: &alerts.Dispatcher{}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPost,
		"/alert-channels/test",
		strings.NewReader(`{"kind":"slack","config":{"webhookUrl":"`+srv.URL+`"}}`),
	)

	h.test(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
	if !strings.Contains(got.Text, "This is a test alert from pingdan.") {
		t.Fatalf("slack text = %q, want test alert", got.Text)
	}
}

func TestAlertCreateRejectsInvalidInputBeforeDatabase(t *testing.T) {
	h := &AlertHandlers{}

	cases := []struct {
		name string
		body string
		want string
	}{
		{"bad json", "{", "bad json"},
		{"bad kind", `{"kind":"fax","label":"Fax","config":{"to":"123"}}`, "kind must be"},
		{"missing label", `{"kind":"webhook","config":{"url":"https://example.com"}}`, "label and config required"},
		{"missing config", `{"kind":"webhook","label":"Hook"}`, "label and config required"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/alert-channels", strings.NewReader(tc.body))

			h.create(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", rec.Code)
			}
			if !strings.Contains(rec.Body.String(), tc.want) {
				t.Fatalf("body = %q, want %q", rec.Body.String(), tc.want)
			}
		})
	}
}
