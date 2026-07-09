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

func TestSendTwilioSMSFormEncoding(t *testing.T) {
	var gotPath string
	var gotUser string
	var gotPass string
	var gotFrom string
	var gotTo string
	var gotBody string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		var ok bool
		gotUser, gotPass, ok = r.BasicAuth()
		if !ok {
			t.Errorf("missing basic auth")
		}
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse form: %v", err)
		}
		gotFrom = r.Form.Get("From")
		gotTo = r.Form.Get("To")
		gotBody = r.Form.Get("Body")
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	d := &Dispatcher{
		Logger:           testLogger(),
		TwilioAccountSID: "AC123",
		TwilioAuthToken:  "auth-token",
		TwilioFrom:       "+15550000000",
		TwilioBaseURL:    srv.URL,
	}
	if err := d.SendTest(context.Background(), KindTwilioSMS, []byte(`{"to":"+15551112222"}`)); err != nil {
		t.Fatalf("SendTest() error = %v", err)
	}
	if gotPath != "/2010-04-01/Accounts/AC123/Messages.json" {
		t.Errorf("path = %q, want Twilio Messages path", gotPath)
	}
	if gotUser != "AC123" || gotPass != "auth-token" {
		t.Errorf("basic auth = %q/%q, want account SID/auth token", gotUser, gotPass)
	}
	if gotFrom != "+15550000000" || gotTo != "+15551112222" {
		t.Errorf("from/to = %q/%q, want configured numbers", gotFrom, gotTo)
	}
	if gotBody != "pingdan: test alert" {
		t.Errorf("body = %q, want terse test body", gotBody)
	}
}

func TestSendTwilioSMSDownAndRecoveredBodies(t *testing.T) {
	var bodies []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse form: %v", err)
		}
		bodies = append(bodies, r.Form.Get("Body"))
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	d := &Dispatcher{
		Logger:           testLogger(),
		TwilioAccountSID: "AC123",
		TwilioAuthToken:  "auth-token",
		TwilioFrom:       "+15550000000",
		TwilioBaseURL:    srv.URL,
	}
	cfg := []byte(`{"to":"+15551112222"}`)
	endpoint := AlertEndpoint{ID: "ep_123", Name: "API", URL: "https://example.com"}
	if err := d.send(context.Background(), KindTwilioSMS, cfg, Alert{Event: EventEndpointDown, Endpoint: endpoint}); err != nil {
		t.Fatalf("down send() error = %v", err)
	}
	if err := d.send(context.Background(), KindTwilioSMS, cfg, Alert{Event: EventEndpointRecovered, Endpoint: endpoint}); err != nil {
		t.Fatalf("recovered send() error = %v", err)
	}
	if strings.Join(bodies, ",") != "pingdan: API DOWN,pingdan: API RECOVERED" {
		t.Errorf("bodies = %#v, want terse state bodies", bodies)
	}
}

func TestSendTwilioSMSRequiresEnv(t *testing.T) {
	d := &Dispatcher{Logger: testLogger()}
	err := d.SendTest(context.Background(), KindTwilioSMS, []byte(`{"to":"+15551112222"}`))
	if err == nil {
		t.Fatal("SendTest() error = nil, want twilio config error")
	}
	if !strings.Contains(err.Error(), "twilio not configured") {
		t.Errorf("error = %q, want twilio not configured", err.Error())
	}
}

func TestSendOpsgenieCreateAndClose(t *testing.T) {
	type request struct {
		Method string
		Path   string
		Query  string
		Auth   string
		Body   map[string]any
	}
	var requests []request

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		requests = append(requests, request{
			Method: r.Method,
			Path:   r.URL.Path,
			Query:  r.URL.RawQuery,
			Auth:   r.Header.Get("Authorization"),
			Body:   body,
		})
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	d := &Dispatcher{Logger: testLogger(), OpsgenieBaseURL: srv.URL}
	cfg := []byte(`{"apiKey":"ops-key","region":"eu"}`)
	endpoint := AlertEndpoint{ID: "ep_123", Name: "API", URL: "https://example.com"}
	if err := d.send(context.Background(), KindOpsgenie, cfg, Alert{
		Event:    EventEndpointDown,
		Endpoint: endpoint,
		Subject:  "API down",
		Body:     "Endpoint: API",
	}); err != nil {
		t.Fatalf("down send() error = %v", err)
	}
	if err := d.send(context.Background(), KindOpsgenie, cfg, Alert{
		Event:    EventEndpointRecovered,
		Endpoint: endpoint,
		Subject:  "API recovered",
		Body:     "Endpoint: API",
	}); err != nil {
		t.Fatalf("recovered send() error = %v", err)
	}
	if err := d.send(context.Background(), KindOpsgenie, cfg, Alert{
		Event:    EventSSLExpiring,
		Endpoint: endpoint,
		Subject:  "SSL expiring",
		Body:     "Renew cert",
	}); err != nil {
		t.Fatalf("ssl send() error = %v", err)
	}

	if len(requests) != 3 {
		t.Fatalf("requests length = %d, want 3", len(requests))
	}
	if requests[0].Method != http.MethodPost || requests[0].Path != "/v2/alerts" {
		t.Errorf("create request = %s %s, want POST /v2/alerts", requests[0].Method, requests[0].Path)
	}
	if requests[0].Auth != "GenieKey ops-key" {
		t.Errorf("auth = %q, want GenieKey", requests[0].Auth)
	}
	if requests[0].Body["alias"] != "pingdan-endpoint-ep_123" || requests[0].Body["priority"] != "P1" {
		t.Errorf("create body = %#v, want endpoint alias and P1", requests[0].Body)
	}
	if requests[1].Path != "/v2/alerts/pingdan-endpoint-ep_123/close" || requests[1].Query != "identifierType=alias" {
		t.Errorf("close target = %s?%s, want alias close", requests[1].Path, requests[1].Query)
	}
	if requests[1].Body["source"] != "pingdan" || requests[1].Body["user"] != "pingdan" {
		t.Errorf("close body = %#v, want pingdan source/user", requests[1].Body)
	}
	if requests[2].Body["alias"] != "pingdan-ssl-ep_123" || requests[2].Body["priority"] != "P3" {
		t.Errorf("ssl body = %#v, want ssl alias and P3", requests[2].Body)
	}
}

func TestSendOpsgenieTestCreatesAndCloses(t *testing.T) {
	var paths []string
	var aliases []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		paths = append(paths, r.URL.Path)
		if alias, ok := body["alias"].(string); ok {
			aliases = append(aliases, alias)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	d := &Dispatcher{Logger: testLogger(), OpsgenieBaseURL: srv.URL}
	if err := d.SendTest(context.Background(), KindOpsgenie, []byte(`{"apiKey":"ops-key","region":"us"}`)); err != nil {
		t.Fatalf("SendTest() error = %v", err)
	}
	if strings.Join(paths, ",") != "/v2/alerts,/v2/alerts/pingdan-test/close" {
		t.Errorf("paths = %#v, want create then close", paths)
	}
	if len(aliases) != 1 || aliases[0] != "pingdan-test" {
		t.Errorf("aliases = %#v, want pingdan-test create alias", aliases)
	}
}

func TestSendOpsgenieRejectsBadRegion(t *testing.T) {
	d := &Dispatcher{Logger: testLogger()}
	err := d.SendTest(context.Background(), KindOpsgenie, []byte(`{"apiKey":"ops-key","region":"apac"}`))
	if err == nil {
		t.Fatal("SendTest() error = nil, want bad region error")
	}
	if !strings.Contains(err.Error(), "us or eu") {
		t.Errorf("error = %q, want region message", err.Error())
	}
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
