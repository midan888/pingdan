package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

type Provider string

const (
	ProviderGoogle Provider = "google"
	ProviderGitHub Provider = "github"
)

type OAuth struct {
	configs map[Provider]*oauth2.Config
	pool    *pgxpool.Pool
	jwt     *JWT
}

func NewOAuth(pool *pgxpool.Pool, jwt *JWT, publicURL, googleID, googleSecret, ghID, ghSecret string) *OAuth {
	o := &OAuth{pool: pool, jwt: jwt, configs: map[Provider]*oauth2.Config{}}
	if googleID != "" {
		o.configs[ProviderGoogle] = &oauth2.Config{
			ClientID:     googleID,
			ClientSecret: googleSecret,
			Endpoint:     google.Endpoint,
			RedirectURL:  publicURL + "/auth/google/callback",
			Scopes:       []string{"openid", "email", "profile"},
		}
	}
	if ghID != "" {
		o.configs[ProviderGitHub] = &oauth2.Config{
			ClientID:     ghID,
			ClientSecret: ghSecret,
			Endpoint:     github.Endpoint,
			RedirectURL:  publicURL + "/auth/github/callback",
			Scopes:       []string{"read:user", "user:email"},
		}
	}
	return o
}

func (o *OAuth) Config(p Provider) (*oauth2.Config, bool) {
	c, ok := o.configs[p]
	return c, ok
}

func RandomState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

type providerUser struct {
	ProviderID string
	Email      string
	Name       string
	AvatarURL  string
}

func (o *OAuth) FetchUser(ctx context.Context, p Provider, tok *oauth2.Token) (*providerUser, error) {
	client := o.configs[p].Client(ctx, tok)
	switch p {
	case ProviderGoogle:
		return fetchGoogle(ctx, client)
	case ProviderGitHub:
		return fetchGitHub(ctx, client)
	}
	return nil, fmt.Errorf("unsupported provider: %s", p)
}

func fetchGoogle(ctx context.Context, c *http.Client) (*providerUser, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://openidconnect.googleapis.com/v1/userinfo", nil)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var body struct {
		Sub     string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	return &providerUser{ProviderID: body.Sub, Email: body.Email, Name: body.Name, AvatarURL: body.Picture}, nil
}

func fetchGitHub(ctx context.Context, c *http.Client) (*providerUser, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var u struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, err
	}
	if u.Email == "" {
		// GitHub may not return primary email in /user — fetch /user/emails
		req2, _ := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user/emails", nil)
		r2, err := c.Do(req2)
		if err == nil {
			defer r2.Body.Close()
			body, _ := io.ReadAll(r2.Body)
			var emails []struct {
				Email    string `json:"email"`
				Primary  bool   `json:"primary"`
				Verified bool   `json:"verified"`
			}
			_ = json.Unmarshal(body, &emails)
			for _, e := range emails {
				if e.Primary && e.Verified {
					u.Email = e.Email
					break
				}
			}
		}
	}
	name := u.Name
	if name == "" {
		name = u.Login
	}
	return &providerUser{ProviderID: fmt.Sprintf("%d", u.ID), Email: u.Email, Name: name, AvatarURL: u.AvatarURL}, nil
}

// UpsertAndIssue finds-or-creates a user and returns a signed JWT.
func (o *OAuth) UpsertAndIssue(ctx context.Context, p Provider, pu *providerUser) (string, error) {
	if pu.Email == "" {
		return "", fmt.Errorf("provider returned no email")
	}
	var id string
	err := o.pool.QueryRow(ctx, `
		INSERT INTO users (email, name, avatar_url, provider, provider_id)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (provider, provider_id) DO UPDATE
		   SET email = EXCLUDED.email, name = EXCLUDED.name, avatar_url = EXCLUDED.avatar_url
		RETURNING id
	`, pu.Email, pu.Name, pu.AvatarURL, string(p), pu.ProviderID).Scan(&id)
	if err != nil {
		return "", err
	}
	return o.jwt.Issue(id, pu.Email)
}
