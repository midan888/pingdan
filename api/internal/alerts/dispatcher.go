package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/smtp"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pingdan/api/internal/checks"
	"github.com/pingdan/api/internal/endpoints"
)

type Dispatcher struct {
	Pool   *pgxpool.Pool
	Logger *slog.Logger

	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string

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
			d.sendEmail(cfg, subject, body)
		case "telegram":
			d.sendTelegram(ctx, cfg, body)
		}
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

func (d *Dispatcher) sendEmail(cfg []byte, subject, body string) {
	if d.SMTPHost == "" {
		d.Logger.Warn("alerts: smtp not configured")
		return
	}
	var ec emailConfig
	if err := json.Unmarshal(cfg, &ec); err != nil || ec.To == "" {
		d.Logger.Error("alerts: bad email config")
		return
	}
	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", d.SMTPFrom, ec.To, subject, body))
	addr := fmt.Sprintf("%s:%d", d.SMTPHost, d.SMTPPort)
	auth := smtp.PlainAuth("", d.SMTPUser, d.SMTPPassword, d.SMTPHost)
	if err := smtp.SendMail(addr, auth, d.SMTPFrom, []string{ec.To}, msg); err != nil {
		d.Logger.Error("alerts: smtp send", "err", err)
	}
}

type telegramConfig struct {
	ChatID string `json:"chatId"`
}

func (d *Dispatcher) sendTelegram(ctx context.Context, cfg []byte, text string) {
	if d.TelegramBotToken == "" {
		d.Logger.Warn("alerts: telegram bot token not configured")
		return
	}
	var tc telegramConfig
	if err := json.Unmarshal(cfg, &tc); err != nil || tc.ChatID == "" {
		d.Logger.Error("alerts: bad telegram config")
		return
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", d.TelegramBotToken)
	payload, _ := json.Marshal(map[string]string{"chat_id": tc.ChatID, "text": text})
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		d.Logger.Error("alerts: telegram send", "err", err)
		return
	}
	resp.Body.Close()
}
