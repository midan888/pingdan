package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr    string
	DatabaseURL string
	JWTSecret   string
	JWTTTL      time.Duration
	PublicURL   string
	FrontendURL string

	GoogleClientID     string
	GoogleClientSecret string
	GitHubClientID     string
	GitHubClientSecret string

	ResendAPIKey string
	EmailFrom    string

	TelegramBotToken string

	PushoverAppToken string

	TwilioAccountSID string
	TwilioAuthToken  string
	TwilioFrom       string

	// AdminEmails is the set of accounts allowed to use the admin API,
	// keyed by lowercased email.
	AdminEmails map[string]bool
}

// IsAdmin reports whether the given account email is on the admin allowlist.
func (c *Config) IsAdmin(email string) bool {
	return c.AdminEmails[strings.ToLower(strings.TrimSpace(email))]
}

func Load() (*Config, error) {
	c := &Config{
		HTTPAddr:           getenv("HTTP_ADDR", ":8080"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		JWTTTL:             durationEnv("JWT_TTL", 168*time.Hour),
		PublicURL:          getenv("PUBLIC_URL", "http://localhost:8080"),
		FrontendURL:        getenv("FRONTEND_URL", "http://localhost:3000"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GitHubClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		GitHubClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		ResendAPIKey:       os.Getenv("RESEND_API_KEY"),
		EmailFrom:          getenv("EMAIL_FROM", "alerts@pingdan.local"),
		TelegramBotToken:   os.Getenv("TELEGRAM_BOT_TOKEN"),
		PushoverAppToken:   os.Getenv("PUSHOVER_APP_TOKEN"),
		TwilioAccountSID:   os.Getenv("TWILIO_ACCOUNT_SID"),
		TwilioAuthToken:    os.Getenv("TWILIO_AUTH_TOKEN"),
		TwilioFrom:         os.Getenv("TWILIO_FROM"),
		AdminEmails:        parseEmailSet(os.Getenv("ADMIN_EMAILS")),
	}

	if c.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if c.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	return c, nil
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func parseEmailSet(v string) map[string]bool {
	set := map[string]bool{}
	for _, e := range strings.Split(v, ",") {
		e = strings.ToLower(strings.TrimSpace(e))
		if e != "" {
			set[e] = true
		}
	}
	return set
}

func durationEnv(k string, def time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
