package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pingdan/api/internal/checks"
	"github.com/pingdan/api/internal/endpoints"
)

type Dispatcher struct {
	Pool   *pgxpool.Pool
	Logger *slog.Logger

	ResendAPIKey string
	EmailFrom    string

	TelegramBotToken string
}

// Notify looks up alert channels attached to the endpoint and dispatches a message per channel.
func (d *Dispatcher) Notify(ctx context.Context, e endpoints.Endpoint, newState string, c *checks.Check) {
	rows, err := d.Pool.Query(ctx, `
		SELECT ac.kind, ac.config
		FROM endpoint_alert_channels eac
		JOIN alert_channels ac ON ac.id = eac.channel_id
		WHERE eac.endpoint_id = $1
	`, e.ID)
	if err != nil {
		d.Logger.Error("alerts: query channels", "err", err)
		return
	}
	defer rows.Close()

	subject, body := renderMessage(e, newState, c)

	for rows.Next() {
		var kind string
		var cfg []byte
		if err := rows.Scan(&kind, &cfg); err != nil {
			continue
		}
		switch kind {
		case "email":
			_ = d.sendEmail(ctx, cfg, subject, body)
		case "telegram":
			_ = d.sendTelegram(ctx, cfg, body)
		}
	}
}

// SendTest dispatches a test notification to a single channel config and returns
// an error if delivery fails, so callers can surface the result to the user.
func (d *Dispatcher) SendTest(ctx context.Context, kind string, cfg []byte) error {
	subject := "[pingdan] Test alert"
	body := fmt.Sprintf("This is a test alert from pingdan.\nIf you received this, the channel is configured correctly.\nAt: %s",
		time.Now().UTC().Format(time.RFC3339))
	switch kind {
	case "email":
		return d.sendEmail(ctx, cfg, subject, body)
	case "telegram":
		return d.sendTelegram(ctx, cfg, body)
	default:
		return fmt.Errorf("unsupported channel kind: %s", kind)
	}
}

func renderMessage(e endpoints.Endpoint, newState string, c *checks.Check) (string, string) {
	status := "DOWN"
	if newState == "up" {
		status = "RECOVERED"
	}
	subject := fmt.Sprintf("[pingdan] %s — %s", e.Name, status)
	var detail string
	if c.Error != nil && *c.Error != "" {
		detail = fmt.Sprintf("Error: %s", *c.Error)
	} else if c.StatusCode != nil {
		detail = fmt.Sprintf("Status: %d", *c.StatusCode)
	}
	body := fmt.Sprintf("Endpoint: %s\nURL: %s\nState: %s\n%s\nAt: %s",
		e.Name, e.URL, status, detail, c.CheckedAt.Format(time.RFC3339))
	return subject, body
}

type emailConfig struct {
	To string `json:"to"`
}

// sendEmail delivers the alert via the Resend HTTP API (https://resend.com/docs/api-reference/emails/send-email).
func (d *Dispatcher) sendEmail(ctx context.Context, cfg []byte, subject, body string) error {
	if d.ResendAPIKey == "" {
		d.Logger.Warn("alerts: email not configured (set RESEND_API_KEY)")
		return fmt.Errorf("email not configured")
	}
	var ec emailConfig
	if err := json.Unmarshal(cfg, &ec); err != nil || ec.To == "" {
		d.Logger.Error("alerts: bad email config")
		return fmt.Errorf("bad email config")
	}
	payload, _ := json.Marshal(map[string]any{
		"from":    d.EmailFrom,
		"to":      []string{ec.To},
		"subject": subject,
		"text":    body,
	})
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.resend.com/emails", bytes.NewReader(payload))
	if err != nil {
		d.Logger.Error("alerts: resend request", "err", err)
		return err
	}
	req.Header.Set("Authorization", "Bearer "+d.ResendAPIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		d.Logger.Error("alerts: resend send", "err", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		d.Logger.Error("alerts: resend send", "status", resp.StatusCode, "body", string(respBody))
		return fmt.Errorf("email provider returned status %d", resp.StatusCode)
	}
	return nil
}

type telegramConfig struct {
	ChatID string `json:"chatId"`
}

func (d *Dispatcher) sendTelegram(ctx context.Context, cfg []byte, text string) error {
	if d.TelegramBotToken == "" {
		d.Logger.Warn("alerts: telegram bot token not configured")
		return fmt.Errorf("telegram not configured")
	}
	var tc telegramConfig
	if err := json.Unmarshal(cfg, &tc); err != nil || tc.ChatID == "" {
		d.Logger.Error("alerts: bad telegram config")
		return fmt.Errorf("bad telegram config")
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", d.TelegramBotToken)
	payload, _ := json.Marshal(map[string]string{"chat_id": tc.ChatID, "text": text})
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		d.Logger.Error("alerts: telegram send", "err", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		d.Logger.Error("alerts: telegram send", "status", resp.StatusCode, "body", string(respBody))
		return fmt.Errorf("telegram returned status %d", resp.StatusCode)
	}
	return nil
}
