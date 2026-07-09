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

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
