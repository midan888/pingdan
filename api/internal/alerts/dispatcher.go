package alerts

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pingdan/api/internal/checks"
	"github.com/pingdan/api/internal/endpoints"
)

const (
	KindEmail     = "email"
	KindTelegram  = "telegram"
	KindSlack     = "slack"
	KindDiscord   = "discord"
	KindTeams     = "teams"
	KindWebhook   = "webhook"
	KindPagerDuty = "pagerduty"
	KindNtfy      = "ntfy"
	KindPushover  = "pushover"
	KindTwilioSMS = "twilio_sms"
	KindOpsgenie  = "opsgenie"

	EventEndpointDown      = "endpoint.down"
	EventEndpointRecovered = "endpoint.recovered"
	EventSSLExpiring       = "ssl.expiring"
	EventTest              = "test"

	defaultResendBaseURL      = "https://api.resend.com"
	defaultTelegramBaseURL    = "https://api.telegram.org"
	defaultPagerDutyEventsURL = "https://events.pagerduty.com/v2/enqueue"
	defaultPushoverBaseURL    = "https://api.pushover.net"
	defaultNtfyServer         = "https://ntfy.sh"
	defaultTwilioBaseURL      = "https://api.twilio.com"
	defaultOpsgenieUSBaseURL  = "https://api.opsgenie.com"
	defaultOpsgenieEUBaseURL  = "https://api.eu.opsgenie.com"
	postJSONTimeout           = 10 * time.Second
)

var ValidKinds = []string{
	KindEmail,
	KindTelegram,
	KindSlack,
	KindDiscord,
	KindTeams,
	KindWebhook,
	KindPagerDuty,
	KindNtfy,
	KindPushover,
	KindTwilioSMS,
	KindOpsgenie,
}

func IsValidKind(kind string) bool {
	for _, valid := range ValidKinds {
		if kind == valid {
			return true
		}
	}
	return false
}

func (d *Dispatcher) KindCapabilities() map[string]bool {
	out := map[string]bool{}
	for _, kind := range ValidKinds {
		out[kind] = true
	}
	out[KindEmail] = d.ResendAPIKey != ""
	out[KindTelegram] = d.TelegramBotToken != ""
	out[KindPushover] = d.PushoverAppToken != ""
	out[KindTwilioSMS] = d.TwilioAccountSID != "" && d.TwilioAuthToken != "" && d.TwilioFrom != ""
	return out
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

	PagerDutyEventsURL string

	PushoverAppToken string
	PushoverBaseURL  string

	TwilioAccountSID string
	TwilioAuthToken  string
	TwilioFrom       string
	TwilioBaseURL    string

	OpsgenieBaseURL string
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
		KindEmail:     d.sendEmail,
		KindTelegram:  d.sendTelegram,
		KindSlack:     d.sendSlack,
		KindDiscord:   d.sendDiscord,
		KindTeams:     d.sendTeams,
		KindWebhook:   d.sendWebhook,
		KindPagerDuty: d.sendPagerDuty,
		KindNtfy:      d.sendNtfy,
		KindPushover:  d.sendPushover,
		KindTwilioSMS: d.sendTwilioSMS,
		KindOpsgenie:  d.sendOpsgenie,
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

type webhookURLConfig struct {
	WebhookURL string `json:"webhookUrl"`
}

func (d *Dispatcher) sendSlack(ctx context.Context, cfg []byte, a Alert) error {
	var sc webhookURLConfig
	if err := json.Unmarshal(cfg, &sc); err != nil || sc.WebhookURL == "" {
		d.logger().Error("alerts: bad slack config")
		return fmt.Errorf("bad slack config")
	}
	return d.postJSONUserURL(ctx, sc.WebhookURL, nil, map[string]string{
		"text": alertText(a),
	})
}

func (d *Dispatcher) sendDiscord(ctx context.Context, cfg []byte, a Alert) error {
	var dc webhookURLConfig
	if err := json.Unmarshal(cfg, &dc); err != nil || dc.WebhookURL == "" {
		d.logger().Error("alerts: bad discord config")
		return fmt.Errorf("bad discord config")
	}
	return d.postJSONUserURL(ctx, dc.WebhookURL, nil, map[string]string{
		"content": truncateRunes(alertText(a), 2000),
	})
}

func (d *Dispatcher) sendTeams(ctx context.Context, cfg []byte, a Alert) error {
	var tc webhookURLConfig
	if err := json.Unmarshal(cfg, &tc); err != nil || tc.WebhookURL == "" {
		d.logger().Error("alerts: bad teams config")
		return fmt.Errorf("bad teams config")
	}
	payload := map[string]any{
		"type": "message",
		"attachments": []map[string]any{
			{
				"contentType": "application/vnd.microsoft.card.adaptive",
				"content": map[string]any{
					"$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
					"type":    "AdaptiveCard",
					"version": "1.4",
					"body": []map[string]any{
						{
							"type":   "TextBlock",
							"text":   a.Subject,
							"weight": "Bolder",
							"wrap":   true,
						},
						{
							"type":  "FactSet",
							"facts": teamsFacts(a),
						},
					},
				},
			},
		},
	}
	return d.postJSONUserURL(ctx, tc.WebhookURL, nil, payload)
}

type genericWebhookConfig struct {
	URL    string `json:"url"`
	Secret string `json:"secret,omitempty"`
}

func (d *Dispatcher) sendWebhook(ctx context.Context, cfg []byte, a Alert) error {
	var wc genericWebhookConfig
	if err := json.Unmarshal(cfg, &wc); err != nil || wc.URL == "" {
		d.logger().Error("alerts: bad webhook config")
		return fmt.Errorf("bad webhook config")
	}
	body, err := json.Marshal(a)
	if err != nil {
		return err
	}
	headers := map[string]string{}
	if wc.Secret != "" {
		headers["X-Pingdan-Signature"] = "sha256=" + hmacSHA256Hex(wc.Secret, body)
	}
	return d.postJSONBody(ctx, wc.URL, headers, body, true)
}

func alertText(a Alert) string {
	return a.Subject + "\n" + a.Body
}

func truncateRunes(s string, limit int) string {
	r := []rune(s)
	if len(r) <= limit {
		return s
	}
	return string(r[:limit])
}

func teamsFacts(a Alert) []map[string]string {
	var facts []map[string]string
	for _, line := range strings.Split(a.Body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		title, value, ok := strings.Cut(line, ":")
		if !ok {
			title, value = "Message", line
		}
		facts = append(facts, map[string]string{
			"title": strings.TrimSpace(title) + ":",
			"value": strings.TrimSpace(value),
		})
	}
	if len(facts) == 0 {
		facts = append(facts, map[string]string{"title": "Message:", "value": a.Subject})
	}
	return facts
}

func hmacSHA256Hex(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

type pagerDutyConfig struct {
	RoutingKey string `json:"routingKey"`
}

type pagerDutyEvent struct {
	RoutingKey  string             `json:"routing_key"`
	EventAction string             `json:"event_action"`
	DedupKey    string             `json:"dedup_key"`
	Payload     pagerDutyEventBody `json:"payload"`
}

type pagerDutyEventBody struct {
	Summary  string `json:"summary"`
	Source   string `json:"source"`
	Severity string `json:"severity"`
}

func (d *Dispatcher) sendPagerDuty(ctx context.Context, cfg []byte, a Alert) error {
	var pc pagerDutyConfig
	if err := json.Unmarshal(cfg, &pc); err != nil || pc.RoutingKey == "" {
		d.logger().Error("alerts: bad pagerduty config")
		return fmt.Errorf("bad pagerduty config")
	}
	events := pagerDutyEvents(pc.RoutingKey, a)
	for _, event := range events {
		if err := d.postJSON(ctx, d.pagerDutyEventsURL(), nil, event); err != nil {
			return fmt.Errorf("pagerduty send: %w", err)
		}
	}
	return nil
}

func pagerDutyEvents(routingKey string, a Alert) []pagerDutyEvent {
	event := pagerDutyEvent{
		RoutingKey: routingKey,
		DedupKey:   pagerDutyDedupKey(a),
		Payload: pagerDutyEventBody{
			Summary:  a.Subject,
			Source:   pagerDutySource(a),
			Severity: pagerDutySeverity(a),
		},
	}
	switch a.Event {
	case EventEndpointRecovered:
		event.EventAction = "resolve"
	case EventTest:
		event.EventAction = "trigger"
		resolve := event
		resolve.EventAction = "resolve"
		return []pagerDutyEvent{event, resolve}
	default:
		event.EventAction = "trigger"
	}
	return []pagerDutyEvent{event}
}

func pagerDutyDedupKey(a Alert) string {
	switch a.Event {
	case EventSSLExpiring:
		return "pingdan-ssl-" + a.Endpoint.ID
	case EventTest:
		return "pingdan-test"
	default:
		return "pingdan-endpoint-" + a.Endpoint.ID
	}
}

func pagerDutySeverity(a Alert) string {
	switch a.Event {
	case EventEndpointDown:
		return "critical"
	case EventSSLExpiring:
		return "warning"
	case EventTest:
		return "info"
	default:
		return "critical"
	}
}

func pagerDutySource(a Alert) string {
	if a.Endpoint.URL != "" {
		return a.Endpoint.URL
	}
	if a.Endpoint.Name != "" {
		return a.Endpoint.Name
	}
	return "pingdan"
}

type ntfyConfig struct {
	Topic       string `json:"topic"`
	Server      string `json:"server,omitempty"`
	AccessToken string `json:"accessToken,omitempty"`
}

func (d *Dispatcher) sendNtfy(ctx context.Context, cfg []byte, a Alert) error {
	var nc ntfyConfig
	if err := json.Unmarshal(cfg, &nc); err != nil || nc.Topic == "" {
		d.logger().Error("alerts: bad ntfy config")
		return fmt.Errorf("bad ntfy config")
	}
	server := nc.Server
	if server == "" {
		server = defaultNtfyServer
	}
	if err := validateUserURL(server); err != nil {
		return err
	}
	target, err := url.JoinPath(server, nc.Topic)
	if err != nil {
		return fmt.Errorf("bad ntfy URL: %w", err)
	}
	headers := map[string]string{
		"Title":    a.Subject,
		"Priority": ntfyPriority(a),
	}
	if nc.AccessToken != "" {
		headers["Authorization"] = "Bearer " + nc.AccessToken
	}
	return d.postTextUserURL(ctx, target, headers, a.Body)
}

func ntfyPriority(a Alert) string {
	if a.Event == EventEndpointDown {
		return "urgent"
	}
	return "default"
}

type pushoverConfig struct {
	UserKey string `json:"userKey"`
}

func (d *Dispatcher) sendPushover(ctx context.Context, cfg []byte, a Alert) error {
	if d.PushoverAppToken == "" {
		d.logger().Warn("alerts: pushover app token not configured")
		return fmt.Errorf("pushover not configured")
	}
	var pc pushoverConfig
	if err := json.Unmarshal(cfg, &pc); err != nil || pc.UserKey == "" {
		d.logger().Error("alerts: bad pushover config")
		return fmt.Errorf("bad pushover config")
	}
	form := url.Values{}
	form.Set("token", d.PushoverAppToken)
	form.Set("user", pc.UserKey)
	form.Set("title", a.Subject)
	form.Set("message", a.Body)
	return d.postForm(ctx, d.pushoverBaseURL()+"/1/messages.json", nil, form)
}

type twilioSMSConfig struct {
	To string `json:"to"`
}

func (d *Dispatcher) sendTwilioSMS(ctx context.Context, cfg []byte, a Alert) error {
	if d.TwilioAccountSID == "" || d.TwilioAuthToken == "" || d.TwilioFrom == "" {
		d.logger().Warn("alerts: twilio not configured")
		return fmt.Errorf("twilio not configured")
	}
	var tc twilioSMSConfig
	if err := json.Unmarshal(cfg, &tc); err != nil || tc.To == "" {
		d.logger().Error("alerts: bad twilio sms config")
		return fmt.Errorf("bad twilio sms config")
	}
	form := url.Values{}
	form.Set("From", d.TwilioFrom)
	form.Set("To", tc.To)
	form.Set("Body", smsBody(a))
	headers := map[string]string{
		"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(d.TwilioAccountSID+":"+d.TwilioAuthToken)),
	}
	endpointURL := fmt.Sprintf("%s/2010-04-01/Accounts/%s/Messages.json", d.twilioBaseURL(), url.PathEscape(d.TwilioAccountSID))
	return d.postForm(ctx, endpointURL, headers, form)
}

func smsBody(a Alert) string {
	switch a.Event {
	case EventEndpointDown:
		return fmt.Sprintf("pingdan: %s DOWN", alertEndpointName(a))
	case EventEndpointRecovered:
		return fmt.Sprintf("pingdan: %s RECOVERED", alertEndpointName(a))
	case EventSSLExpiring:
		if a.SSL != nil {
			return fmt.Sprintf("pingdan: %s SSL expires in %d days", alertEndpointName(a), a.SSL.DaysLeft)
		}
		return fmt.Sprintf("pingdan: %s SSL expiring", alertEndpointName(a))
	case EventTest:
		return "pingdan: test alert"
	default:
		return truncateRunes("pingdan: "+a.Subject, 160)
	}
}

func alertEndpointName(a Alert) string {
	if a.Endpoint.Name != "" {
		return a.Endpoint.Name
	}
	return "endpoint"
}

type opsgenieConfig struct {
	APIKey string `json:"apiKey"`
	Region string `json:"region,omitempty"`
}

type opsgenieAlertPayload struct {
	Message     string            `json:"message"`
	Alias       string            `json:"alias"`
	Description string            `json:"description,omitempty"`
	Source      string            `json:"source"`
	Priority    string            `json:"priority"`
	Details     map[string]string `json:"details,omitempty"`
}

type opsgenieClosePayload struct {
	User   string `json:"user"`
	Source string `json:"source"`
	Note   string `json:"note"`
}

func (d *Dispatcher) sendOpsgenie(ctx context.Context, cfg []byte, a Alert) error {
	var oc opsgenieConfig
	if err := json.Unmarshal(cfg, &oc); err != nil || oc.APIKey == "" {
		d.logger().Error("alerts: bad opsgenie config")
		return fmt.Errorf("bad opsgenie config")
	}
	baseURL, err := d.opsgenieBaseURL(oc.Region)
	if err != nil {
		return err
	}
	headers := map[string]string{"Authorization": "GenieKey " + oc.APIKey}

	switch a.Event {
	case EventEndpointRecovered:
		return d.closeOpsgenieAlert(ctx, baseURL, headers, opsgenieEndpointAlias(a), a)
	case EventTest:
		if err := d.createOpsgenieAlert(ctx, baseURL, headers, opsgenieTestAlias(), a); err != nil {
			return err
		}
		return d.closeOpsgenieAlert(ctx, baseURL, headers, opsgenieTestAlias(), a)
	default:
		return d.createOpsgenieAlert(ctx, baseURL, headers, opsgenieAlias(a), a)
	}
}

func (d *Dispatcher) createOpsgenieAlert(ctx context.Context, baseURL string, headers map[string]string, alias string, a Alert) error {
	payload := opsgenieAlertPayload{
		Message:     truncateRunes(a.Subject, 130),
		Alias:       alias,
		Description: a.Body,
		Source:      "pingdan",
		Priority:    opsgeniePriority(a),
		Details:     opsgenieDetails(a),
	}
	if payload.Message == "" {
		payload.Message = "pingdan alert"
	}
	if err := d.postJSON(ctx, baseURL+"/v2/alerts", headers, payload); err != nil {
		return fmt.Errorf("opsgenie create: %w", err)
	}
	return nil
}

func (d *Dispatcher) closeOpsgenieAlert(ctx context.Context, baseURL string, headers map[string]string, alias string, a Alert) error {
	escapedAlias := url.PathEscape(alias)
	payload := opsgenieClosePayload{
		User:   "pingdan",
		Source: "pingdan",
		Note:   a.Subject,
	}
	if payload.Note == "" {
		payload.Note = "pingdan resolved alert"
	}
	if err := d.postJSON(ctx, baseURL+"/v2/alerts/"+escapedAlias+"/close?identifierType=alias", headers, payload); err != nil {
		return fmt.Errorf("opsgenie close: %w", err)
	}
	return nil
}

func opsgenieAlias(a Alert) string {
	if a.Event == EventSSLExpiring {
		return "pingdan-ssl-" + a.Endpoint.ID
	}
	if a.Event == EventTest {
		return opsgenieTestAlias()
	}
	return opsgenieEndpointAlias(a)
}

func opsgenieEndpointAlias(a Alert) string {
	return "pingdan-endpoint-" + a.Endpoint.ID
}

func opsgenieTestAlias() string {
	return "pingdan-test"
}

func opsgeniePriority(a Alert) string {
	switch a.Event {
	case EventEndpointDown:
		return "P1"
	case EventSSLExpiring:
		return "P3"
	case EventTest:
		return "P5"
	default:
		return "P3"
	}
}

func opsgenieDetails(a Alert) map[string]string {
	details := map[string]string{"event": a.Event}
	if a.Endpoint.ID != "" {
		details["endpointId"] = a.Endpoint.ID
	}
	if a.Endpoint.Name != "" {
		details["endpointName"] = a.Endpoint.Name
	}
	if a.Endpoint.URL != "" {
		details["endpointUrl"] = a.Endpoint.URL
	}
	if a.Check != nil {
		if a.Check.StatusCode != nil {
			details["statusCode"] = fmt.Sprintf("%d", *a.Check.StatusCode)
		}
		if a.Check.Error != nil && *a.Check.Error != "" {
			details["error"] = *a.Check.Error
		}
		details["checkedAt"] = a.Check.CheckedAt.Format(time.RFC3339)
	}
	if a.SSL != nil {
		details["sslDaysLeft"] = fmt.Sprintf("%d", a.SSL.DaysLeft)
		details["sslExpiresAt"] = a.SSL.ExpiresAt.Format(time.RFC3339)
	}
	return details
}

func (d *Dispatcher) postJSON(ctx context.Context, url string, headers map[string]string, payload any) error {
	return d.postJSONPayload(ctx, url, headers, payload, false)
}

func (d *Dispatcher) postJSONUserURL(ctx context.Context, url string, headers map[string]string, payload any) error {
	return d.postJSONPayload(ctx, url, headers, payload, true)
}

func (d *Dispatcher) postTextUserURL(ctx context.Context, url string, headers map[string]string, body string) error {
	return d.postBody(ctx, url, headers, []byte(body), "text/plain; charset=utf-8", true)
}

func (d *Dispatcher) postForm(ctx context.Context, rawURL string, headers map[string]string, form url.Values) error {
	return d.postBody(ctx, rawURL, headers, []byte(form.Encode()), "application/x-www-form-urlencoded", false)
}

func (d *Dispatcher) postJSONPayload(ctx context.Context, url string, headers map[string]string, payload any, userSuppliedURL bool) error {
	body, err := json.Marshal(payload)
	if err != nil {
		d.logger().Error("alerts: marshal json payload", "err", err)
		return err
	}
	return d.postJSONBody(ctx, url, headers, body, userSuppliedURL)
}

func (d *Dispatcher) postJSONBody(ctx context.Context, rawURL string, headers map[string]string, body []byte, userSuppliedURL bool) error {
	return d.postBody(ctx, rawURL, headers, body, "application/json", userSuppliedURL)
}

func (d *Dispatcher) postBody(ctx context.Context, rawURL string, headers map[string]string, body []byte, contentType string, userSuppliedURL bool) error {
	if userSuppliedURL {
		if err := validateUserURL(rawURL); err != nil {
			return err
		}
	}

	ctx, cancel := context.WithTimeout(ctx, postJSONTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, bytes.NewReader(body))
	if err != nil {
		d.logger().Error("alerts: json request", "err", err)
		return err
	}
	req.Header.Set("Content-Type", contentType)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := d.httpClientForRequest(userSuppliedURL).Do(req)
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

func validateUserURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid webhook url: %w", err)
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("webhook url must use http or https")
	}
	if u.Host == "" {
		return fmt.Errorf("webhook url must include a host")
	}
	return nil
}

func (d *Dispatcher) httpClient() *http.Client {
	if d.HTTPClient != nil {
		return d.HTTPClient
	}
	return http.DefaultClient
}

func (d *Dispatcher) httpClientForRequest(userSuppliedURL bool) *http.Client {
	client := d.httpClient()
	if !userSuppliedURL {
		return client
	}
	c := *client
	c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return fmt.Errorf("stopped after 10 redirects")
		}
		if req.URL.Scheme != "https" {
			return fmt.Errorf("redirect to non-https URL blocked")
		}
		return nil
	}
	return &c
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

func (d *Dispatcher) pagerDutyEventsURL() string {
	if d.PagerDutyEventsURL == "" {
		return defaultPagerDutyEventsURL
	}
	return strings.TrimRight(d.PagerDutyEventsURL, "/")
}

func (d *Dispatcher) pushoverBaseURL() string {
	if d.PushoverBaseURL == "" {
		return defaultPushoverBaseURL
	}
	return strings.TrimRight(d.PushoverBaseURL, "/")
}

func (d *Dispatcher) twilioBaseURL() string {
	if d.TwilioBaseURL == "" {
		return defaultTwilioBaseURL
	}
	return strings.TrimRight(d.TwilioBaseURL, "/")
}

func (d *Dispatcher) opsgenieBaseURL(region string) (string, error) {
	if d.OpsgenieBaseURL != "" {
		return strings.TrimRight(d.OpsgenieBaseURL, "/"), nil
	}
	switch strings.ToLower(strings.TrimSpace(region)) {
	case "", "us":
		return defaultOpsgenieUSBaseURL, nil
	case "eu":
		return defaultOpsgenieEUBaseURL, nil
	default:
		return "", fmt.Errorf("opsgenie region must be us or eu")
	}
}

func (d *Dispatcher) logger() *slog.Logger {
	if d.Logger != nil {
		return d.Logger
	}
	return slog.Default()
}
