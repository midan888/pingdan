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

func TestSendPagerDutyEventMapping(t *testing.T) {
	type payload struct {
		Summary  string `json:"summary"`
		Source   string `json:"source"`
		Severity string `json:"severity"`
	}
	var got []struct {
		RoutingKey  string  `json:"routing_key"`
		EventAction string  `json:"event_action"`
		DedupKey    string  `json:"dedup_key"`
		Payload     payload `json:"payload"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var event struct {
			RoutingKey  string  `json:"routing_key"`
			EventAction string  `json:"event_action"`
			DedupKey    string  `json:"dedup_key"`
			Payload     payload `json:"payload"`
		}
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		got = append(got, event)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	d := &Dispatcher{Logger: testLogger(), PagerDutyEventsURL: srv.URL}
	cfg := []byte(`{"routingKey":"routing-key"}`)
	endpoint := AlertEndpoint{ID: "ep_123", Name: "API", URL: "https://example.com"}
	if err := d.send(context.Background(), KindPagerDuty, cfg, Alert{
		Event:    EventEndpointDown,
		Endpoint: endpoint,
		Subject:  "API down",
	}); err != nil {
		t.Fatalf("down send() error = %v", err)
	}
	if err := d.send(context.Background(), KindPagerDuty, cfg, Alert{
		Event:    EventEndpointRecovered,
		Endpoint: endpoint,
		Subject:  "API recovered",
	}); err != nil {
		t.Fatalf("recovered send() error = %v", err)
	}
	if err := d.send(context.Background(), KindPagerDuty, cfg, Alert{
		Event:    EventSSLExpiring,
		Endpoint: endpoint,
		Subject:  "SSL expiring",
	}); err != nil {
		t.Fatalf("ssl send() error = %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("events length = %d, want 3", len(got))
	}
	cases := []struct {
		i        int
		action   string
		dedupKey string
		severity string
	}{
		{0, "trigger", "pingdan-endpoint-ep_123", "critical"},
		{1, "resolve", "pingdan-endpoint-ep_123", "critical"},
		{2, "trigger", "pingdan-ssl-ep_123", "warning"},
	}
	for _, c := range cases {
		if got[c.i].RoutingKey != "routing-key" {
			t.Errorf("event %d routing_key = %q, want routing-key", c.i, got[c.i].RoutingKey)
		}
		if got[c.i].EventAction != c.action {
			t.Errorf("event %d action = %q, want %q", c.i, got[c.i].EventAction, c.action)
		}
		if got[c.i].DedupKey != c.dedupKey {
			t.Errorf("event %d dedup_key = %q, want %q", c.i, got[c.i].DedupKey, c.dedupKey)
		}
		if got[c.i].Payload.Severity != c.severity {
			t.Errorf("event %d severity = %q, want %q", c.i, got[c.i].Payload.Severity, c.severity)
		}
	}
}

func TestSendPagerDutyTestTriggersAndResolves(t *testing.T) {
	var actions []string
	var dedupKeys []string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var event struct {
			EventAction string `json:"event_action"`
			DedupKey    string `json:"dedup_key"`
			Payload     struct {
				Severity string `json:"severity"`
			} `json:"payload"`
		}
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		if event.Payload.Severity != "info" {
			t.Errorf("severity = %q, want info", event.Payload.Severity)
		}
		actions = append(actions, event.EventAction)
		dedupKeys = append(dedupKeys, event.DedupKey)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	d := &Dispatcher{Logger: testLogger(), PagerDutyEventsURL: srv.URL}
	if err := d.SendTest(context.Background(), KindPagerDuty, []byte(`{"routingKey":"routing-key"}`)); err != nil {
		t.Fatalf("SendTest() error = %v", err)
	}
	if strings.Join(actions, ",") != "trigger,resolve" {
		t.Errorf("actions = %#v, want trigger then resolve", actions)
	}
	if len(dedupKeys) != 2 || dedupKeys[0] != "pingdan-test" || dedupKeys[1] != "pingdan-test" {
		t.Errorf("dedupKeys = %#v, want two pingdan-test keys", dedupKeys)
	}
}

func TestSendNtfyHeadersAndBody(t *testing.T) {
	var gotPath string
	var gotTitle string
	var gotPriority string
	var gotAuth string
	var gotBody string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		gotPath = r.URL.Path
		gotTitle = r.Header.Get("Title")
		gotPriority = r.Header.Get("Priority")
		gotAuth = r.Header.Get("Authorization")
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	d := &Dispatcher{Logger: testLogger()}
	cfg := []byte(`{"topic":"ops","server":"` + srv.URL + `","accessToken":"ntfy-token"}`)
	err := d.send(context.Background(), KindNtfy, cfg, Alert{
		Event:   EventEndpointDown,
		Subject: "API down",
		Body:    "Endpoint: API",
	})
	if err != nil {
		t.Fatalf("send() error = %v", err)
	}
	if gotPath != "/ops" {
		t.Errorf("path = %q, want /ops", gotPath)
	}
	if gotTitle != "API down" {
		t.Errorf("Title = %q, want API down", gotTitle)
	}
	if gotPriority != "urgent" {
		t.Errorf("Priority = %q, want urgent", gotPriority)
	}
	if gotAuth != "Bearer ntfy-token" {
		t.Errorf("Authorization = %q, want bearer token", gotAuth)
	}
	if gotBody != "Endpoint: API" {
		t.Errorf("body = %q, want alert body", gotBody)
	}
}

func TestSendPushoverFormEncoding(t *testing.T) {
	var gotPath string
	var gotContentType string
	var gotToken string
	var gotUser string
	var gotTitle string
	var gotMessage string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse form: %v", err)
		}
		gotToken = r.Form.Get("token")
		gotUser = r.Form.Get("user")
		gotTitle = r.Form.Get("title")
		gotMessage = r.Form.Get("message")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	d := &Dispatcher{
		Logger:           testLogger(),
		PushoverAppToken: "app-token",
		PushoverBaseURL:  srv.URL,
	}
	if err := d.SendTest(context.Background(), KindPushover, []byte(`{"userKey":"user-key"}`)); err != nil {
		t.Fatalf("SendTest() error = %v", err)
	}
	if gotPath != "/1/messages.json" {
		t.Errorf("path = %q, want /1/messages.json", gotPath)
	}
	if gotContentType != "application/x-www-form-urlencoded" {
		t.Errorf("content-type = %q, want form encoded", gotContentType)
	}
	if gotToken != "app-token" || gotUser != "user-key" {
		t.Errorf("token/user = %q/%q, want app-token/user-key", gotToken, gotUser)
	}
	if gotTitle != "[pingdan] Test alert" {
		t.Errorf("title = %q, want test subject", gotTitle)
	}
	if !strings.Contains(gotMessage, "This is a test alert from pingdan.") {
		t.Errorf("message = %q, want test body", gotMessage)
	}
}

func TestSendPushoverRequiresAppToken(t *testing.T) {
	d := &Dispatcher{Logger: testLogger()}
	err := d.SendTest(context.Background(), KindPushover, []byte(`{"userKey":"user-key"}`))
	if err == nil {
		t.Fatal("SendTest() error = nil, want app-token error")
	}
	if !strings.Contains(err.Error(), "pushover not configured") {
		t.Errorf("error = %q, want pushover not configured", err.Error())
	}
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
