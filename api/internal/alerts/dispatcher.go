package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pingdan/api/internal/checks"
	"github.com/pingdan/api/internal/endpoints"
)

const (
	KindEmail    = "email"
	KindTelegram = "telegram"

	EventEndpointDown      = "endpoint.down"
	EventEndpointRecovered = "endpoint.recovered"
	EventSSLExpiring       = "ssl.expiring"
	EventTest              = "test"

	defaultResendBaseURL   = "https://api.resend.com"
	defaultTelegramBaseURL = "https://api.telegram.org"
	postJSONTimeout        = 10 * time.Second
)

var ValidKinds = []string{KindEmail, KindTelegram}

func IsValidKind(kind string) bool {
	for _, valid := range ValidKinds {
		if kind == valid {
			return true
		}
	}
	return false
}

type Dispatcher struct {
	Pool   *pgxpool.Pool
	Logger *slog.Logger

	HTTPClient *http.Client

	ResendAPIKey  string
	EmailFrom     string
	ResendBaseURL string

	TelegramBotToken string
	TelegramBaseURL  string
}

type Alert struct {
	Event    string        `json:"event"`
	Endpoint AlertEndpoint `json:"endpoint"`
	Check    *AlertCheck   `json:"check,omitempty"`
	SSL      *AlertSSL     `json:"ssl,omitempty"`
	Subject  string        `json:"subject"`
	Body     string        `json:"body"`
}

type AlertEndpoint struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type AlertCheck struct {
	StatusCode *int      `json:"statusCode,omitempty"`
	Error      *string   `json:"error,omitempty"`
	CheckedAt  time.Time `json:"checkedAt"`
}

type AlertSSL struct {
	DaysLeft  int       `json:"daysLeft"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type alertSender func(context.Context, []byte, Alert) error

func endpointAlert(e endpoints.Endpoint) AlertEndpoint {
	return AlertEndpoint{ID: e.ID, Name: e.Name, URL: e.URL}
}

func checkAlert(c *checks.Check) *AlertCheck {
	if c == nil {
		return nil
	}
	return &AlertCheck{
		StatusCode: c.StatusCode,
		Error:      c.Error,
		CheckedAt:  c.CheckedAt,
	}
}

func (d *Dispatcher) senders() map[string]alertSender {
	return map[string]alertSender{
		KindEmail:    d.sendEmail,
		KindTelegram: d.sendTelegram,
	}
}

func (d *Dispatcher) send(ctx context.Context, kind string, cfg []byte, a Alert) error {
	sender, ok := d.senders()[kind]
	if !ok {
		return fmt.Errorf("unsupported channel kind: %s", kind)
	}
	return sender(ctx, cfg, a)
}

// Notify looks up alert channels attached to the endpoint and dispatches a message per channel.
func (d *Dispatcher) Notify(ctx context.Context, e endpoints.Endpoint, newState string, c *checks.Check) {
	subject, body := renderMessage(e, newState, c)
	event := EventEndpointDown
	if newState == "up" {
		event = EventEndpointRecovered
	}
	d.notifyChannels(ctx, e.ID, Alert{
		Event:    event,
		Endpoint: endpointAlert(e),
		Check:    checkAlert(c),
		Subject:  subject,
		Body:     body,
	})
}

// NotifySSL dispatches a TLS-certificate-expiry warning to every alert channel
// attached to the endpoint. daysLeft is the whole number of days until expiry.
func (d *Dispatcher) NotifySSL(ctx context.Context, e endpoints.Endpoint, daysLeft int, expiresAt time.Time) {
	subject, body := renderSSLMessage(e, daysLeft, expiresAt)
	d.notifyChannels(ctx, e.ID, Alert{
		Event:    EventSSLExpiring,
		Endpoint: endpointAlert(e),
		SSL: &AlertSSL{
			DaysLeft:  daysLeft,
			ExpiresAt: expiresAt,
		},
		Subject: subject,
		Body:    body,
	})
}

func (d *Dispatcher) notifyChannels(ctx context.Context, endpointID string, a Alert) {
	rows, err := d.Pool.Query(ctx, `
		SELECT ac.kind, ac.config
		FROM endpoint_alert_channels eac
		JOIN alert_channels ac ON ac.id = eac.channel_id
		WHERE eac.endpoint_id = $1
	`, endpointID)
	if err != nil {
		d.logger().Error("alerts: query channels", "err", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var kind string
		var cfg []byte
		if err := rows.Scan(&kind, &cfg); err != nil {
			d.logger().Error("alerts: scan channel", "err", err)
			continue
		}
		if err := d.send(ctx, kind, cfg, a); err != nil {
			d.logger().Error("alerts: send", "kind", kind, "event", a.Event, "err", err)
		}
	}
	if err := rows.Err(); err != nil {
		d.logger().Error("alerts: channel rows", "err", err)
	}
}

// SendTest dispatches a test notification to a single channel config and returns
// an error if delivery fails, so callers can surface the result to the user.
func (d *Dispatcher) SendTest(ctx context.Context, kind string, cfg []byte) error {
	subject := "[pingdan] Test alert"
	body := fmt.Sprintf("This is a test alert from pingdan.\nIf you received this, the channel is configured correctly.\nAt: %s",
		time.Now().UTC().Format(time.RFC3339))
	return d.send(ctx, kind, cfg, Alert{
		Event:   EventTest,
		Subject: subject,
		Body:    body,
	})
}

func renderMessage(e endpoints.Endpoint, newState string, c *checks.Check) (string, string) {
	status := "DOWN"
	if newState == "up" {
		status = "RECOVERED"
	}
	subject := fmt.Sprintf("[pingdan] %s — %s", e.Name, status)
	var detail string
	var checkedAt string
	if c != nil {
		if c.Error != nil && *c.Error != "" {
			detail = fmt.Sprintf("Error: %s", *c.Error)
		} else if c.StatusCode != nil {
			detail = fmt.Sprintf("Status: %d", *c.StatusCode)
		}
		checkedAt = c.CheckedAt.Format(time.RFC3339)
	}
	body := fmt.Sprintf("Endpoint: %s\nURL: %s\nState: %s\n%s\nAt: %s",
		e.Name, e.URL, status, detail, checkedAt)
	return subject, body
}

func renderSSLMessage(e endpoints.Endpoint, daysLeft int, expiresAt time.Time) (string, string) {
	dayWord := "days"
	if daysLeft == 1 {
		dayWord = "day"
	}
	subject := fmt.Sprintf("[pingdan] SSL expires in %d %s — %s", daysLeft, dayWord, e.Name)
	body := fmt.Sprintf("The SSL certificate for %s is expiring soon.\nEndpoint: %s\nURL: %s\nDays left: %d\nExpires: %s\n\nRenew the certificate to avoid an outage.",
		e.Name, e.Name, e.URL, daysLeft, expiresAt.Format(time.RFC1123))
	return subject, body
}

type emailConfig struct {
	To string `json:"to"`
}

func (d *Dispatcher) sendEmail(ctx context.Context, cfg []byte, a Alert) error {
	if d.ResendAPIKey == "" {
		d.logger().Warn("alerts: email not configured (set RESEND_API_KEY)")
		return fmt.Errorf("email not configured")
	}
	var ec emailConfig
	if err := json.Unmarshal(cfg, &ec); err != nil || ec.To == "" {
		d.logger().Error("alerts: bad email config")
		return fmt.Errorf("bad email config")
	}
	payload := map[string]any{
		"from":    d.EmailFrom,
		"to":      []string{ec.To},
		"subject": a.Subject,
		"text":    a.Body,
	}
	if err := d.postJSON(ctx, d.resendBaseURL()+"/emails", map[string]string{
		"Authorization": "Bearer " + d.ResendAPIKey,
	}, payload); err != nil {
		return fmt.Errorf("email send: %w", err)
	}
	return nil
}

type telegramConfig struct {
	ChatID string `json:"chatId"`
}

func (d *Dispatcher) sendTelegram(ctx context.Context, cfg []byte, a Alert) error {
	if d.TelegramBotToken == "" {
		d.logger().Warn("alerts: telegram bot token not configured")
		return fmt.Errorf("telegram not configured")
	}
	var tc telegramConfig
	if err := json.Unmarshal(cfg, &tc); err != nil || tc.ChatID == "" {
		d.logger().Error("alerts: bad telegram config")
		return fmt.Errorf("bad telegram config")
	}
	url := fmt.Sprintf("%s/bot%s/sendMessage", d.telegramBaseURL(), d.TelegramBotToken)
	if err := d.postJSON(ctx, url, nil, map[string]string{"chat_id": tc.ChatID, "text": a.Body}); err != nil {
		return fmt.Errorf("telegram send: %w", err)
	}
	return nil
}

func (d *Dispatcher) postJSON(ctx context.Context, url string, headers map[string]string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		d.logger().Error("alerts: marshal json payload", "err", err)
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, postJSONTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		d.logger().Error("alerts: json request", "err", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := d.httpClient().Do(req)
	if err != nil {
		d.logger().Error("alerts: post json", "err", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		d.logger().Error("alerts: post json", "status", resp.StatusCode, "body", string(respBody))
		return fmt.Errorf("notification provider returned status %d", resp.StatusCode)
	}
	return nil
}

func (d *Dispatcher) httpClient() *http.Client {
	if d.HTTPClient != nil {
		return d.HTTPClient
	}
	return http.DefaultClient
}

func (d *Dispatcher) resendBaseURL() string {
	if d.ResendBaseURL == "" {
		return defaultResendBaseURL
	}
	return strings.TrimRight(d.ResendBaseURL, "/")
}

func (d *Dispatcher) telegramBaseURL() string {
	if d.TelegramBaseURL == "" {
		return defaultTelegramBaseURL
	}
	return strings.TrimRight(d.TelegramBaseURL, "/")
}

func (d *Dispatcher) logger() *slog.Logger {
	if d.Logger != nil {
		return d.Logger
	}
	return slog.Default()
}
