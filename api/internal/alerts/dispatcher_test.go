package alerts

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSendTestEmailPayload(t *testing.T) {
	var got struct {
		From    string   `json:"from"`
		To      []string `json:"to"`
		Subject string   `json:"subject"`
		Text    string   `json:"text"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/emails" {
			t.Errorf("path = %s, want /emails", r.URL.Path)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer resend-key" {
			t.Errorf("authorization = %q, want bearer key", auth)
		}
		if ct := r.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
			t.Errorf("content-type = %q, want application/json", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	d := &Dispatcher{
		Logger:        testLogger(),
		ResendAPIKey:  "resend-key",
		EmailFrom:     "alerts@pingdan.test",
		ResendBaseURL: srv.URL,
	}
	if err := d.SendTest(context.Background(), KindEmail, []byte(`{"to":"ops@example.com"}`)); err != nil {
		t.Fatalf("SendTest() error = %v", err)
	}
	if got.From != "alerts@pingdan.test" {
		t.Errorf("from = %q, want alerts@pingdan.test", got.From)
	}
	if len(got.To) != 1 || got.To[0] != "ops@example.com" {
		t.Errorf("to = %#v, want ops@example.com", got.To)
	}
	if got.Subject != "[pingdan] Test alert" {
		t.Errorf("subject = %q, want test subject", got.Subject)
	}
	if !strings.Contains(got.Text, "This is a test alert from pingdan.") {
		t.Errorf("text = %q, want test alert body", got.Text)
	}
}

func TestSendTestTelegramPayload(t *testing.T) {
	var got struct {
		ChatID string `json:"chat_id"`
		Text   string `json:"text"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/bottelegram-token/sendMessage" {
			t.Errorf("path = %s, want /bottelegram-token/sendMessage", r.URL.Path)
		}
		if ct := r.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
			t.Errorf("content-type = %q, want application/json", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	d := &Dispatcher{
		Logger:           testLogger(),
		TelegramBotToken: "telegram-token",
		TelegramBaseURL:  srv.URL,
	}
	if err := d.SendTest(context.Background(), KindTelegram, []byte(`{"chatId":"123456789"}`)); err != nil {
		t.Fatalf("SendTest() error = %v", err)
	}
	if got.ChatID != "123456789" {
		t.Errorf("chat_id = %q, want 123456789", got.ChatID)
	}
	if !strings.Contains(got.Text, "This is a test alert from pingdan.") {
		t.Errorf("text = %q, want test alert body", got.Text)
	}
}

func TestSendTestSlackPayload(t *testing.T) {
	var got struct {
		Text string `json:"text"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	d := &Dispatcher{Logger: testLogger()}
	if err := d.SendTest(context.Background(), KindSlack, []byte(`{"webhookUrl":"`+srv.URL+`"}`)); err != nil {
		t.Fatalf("SendTest() error = %v", err)
	}
	if !strings.HasPrefix(got.Text, "[pingdan] Test alert\nThis is a test alert from pingdan.") {
		t.Errorf("text = %q, want subject and body", got.Text)
	}
}

func TestSendDiscordTruncatesContent(t *testing.T) {
	var got struct {
		Content string `json:"content"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	d := &Dispatcher{Logger: testLogger()}
	err := d.send(context.Background(), KindDiscord, []byte(`{"webhookUrl":"`+srv.URL+`"}`), Alert{
		Subject: "Discord test",
		Body:    strings.Repeat("x", 2500),
	})
	if err != nil {
		t.Fatalf("send() error = %v", err)
	}
	if n := len([]rune(got.Content)); n != 2000 {
		t.Errorf("content length = %d, want 2000", n)
	}
}

func TestSendTeamsPayload(t *testing.T) {
	var got struct {
		Type        string `json:"type"`
		Attachments []struct {
			ContentType string `json:"contentType"`
			Content     struct {
				Type string `json:"type"`
				Body []struct {
					Type  string `json:"type"`
					Text  string `json:"text,omitempty"`
					Facts []struct {
						Title string `json:"title"`
						Value string `json:"value"`
					} `json:"facts,omitempty"`
				} `json:"body"`
			} `json:"content"`
		} `json:"attachments"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	d := &Dispatcher{Logger: testLogger()}
	err := d.send(context.Background(), KindTeams, []byte(`{"webhookUrl":"`+srv.URL+`"}`), Alert{
		Subject: "[pingdan] API - DOWN",
		Body:    "Endpoint: API\nURL: https://example.com\nState: DOWN",
	})
	if err != nil {
		t.Fatalf("send() error = %v", err)
	}
	if got.Type != "message" {
		t.Errorf("type = %q, want message", got.Type)
	}
	if len(got.Attachments) != 1 {
		t.Fatalf("attachments length = %d, want 1", len(got.Attachments))
	}
	if got.Attachments[0].ContentType != "application/vnd.microsoft.card.adaptive" {
		t.Errorf("contentType = %q, want adaptive card", got.Attachments[0].ContentType)
	}
	if got.Attachments[0].Content.Body[0].Text != "[pingdan] API - DOWN" {
		t.Errorf("title = %q, want subject", got.Attachments[0].Content.Body[0].Text)
	}
	facts := got.Attachments[0].Content.Body[1].Facts
	if len(facts) < 3 || facts[0].Title != "Endpoint:" || facts[0].Value != "API" {
		t.Errorf("facts = %#v, want endpoint fact first", facts)
	}
}

func TestSendGenericWebhookPayloadAndSignature(t *testing.T) {
	statusCode := 503
	checkedAt := time.Date(2026, 7, 9, 12, 0, 0, 0, time.UTC)
	var got Alert

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		wantSig := "sha256=" + hmacSHA256Hex("topsecret", body)
		if sig := r.Header.Get("X-Pingdan-Signature"); sig != wantSig {
			t.Errorf("signature = %q, want %q", sig, wantSig)
		}
		if err := json.Unmarshal(body, &got); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	d := &Dispatcher{Logger: testLogger()}
	err := d.send(context.Background(), KindWebhook, []byte(`{"url":"`+srv.URL+`","secret":"topsecret"}`), Alert{
		Event:    EventEndpointDown,
		Endpoint: AlertEndpoint{ID: "ep_123", Name: "API", URL: "https://example.com"},
		Check:    &AlertCheck{StatusCode: &statusCode, CheckedAt: checkedAt},
		Subject:  "[pingdan] API - DOWN",
		Body:     "Endpoint: API\nStatus: 503",
	})
	if err != nil {
		t.Fatalf("send() error = %v", err)
	}
	if got.Event != EventEndpointDown {
		t.Errorf("event = %q, want %q", got.Event, EventEndpointDown)
	}
	if got.Endpoint.ID != "ep_123" || got.Check == nil || got.Check.StatusCode == nil || *got.Check.StatusCode != 503 {
		t.Errorf("payload = %#v, want endpoint and check data", got)
	}
}

func TestUserWebhookRejectsBadScheme(t *testing.T) {
	d := &Dispatcher{Logger: testLogger()}
	err := d.SendTest(context.Background(), KindWebhook, []byte(`{"url":"file:///tmp/nope"}`))
	if err == nil {
		t.Fatal("SendTest() error = nil, want bad scheme error")
	}
	if !strings.Contains(err.Error(), "http or https") {
		t.Errorf("error = %q, want scheme message", err.Error())
	}
}

func TestUserWebhookRejectsRedirectToNonHTTPS(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://example.com/nope", http.StatusFound)
	}))
	defer srv.Close()

	d := &Dispatcher{Logger: testLogger()}
	err := d.SendTest(context.Background(), KindWebhook, []byte(`{"url":"`+srv.URL+`"}`))
	if err == nil {
		t.Fatal("SendTest() error = nil, want redirect error")
	}
	if !strings.Contains(err.Error(), "non-https") {
		t.Errorf("error = %q, want non-https redirect message", err.Error())
	}
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
