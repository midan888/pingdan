package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

const emailProvider = "email"

var (
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrWeakPassword       = errors.New("password must be at least 8 characters")
)

type Email struct {
	pool *pgxpool.Pool
	jwt  *JWT
}

func NewEmail(pool *pgxpool.Pool, jwt *JWT) *Email {
	return &Email{pool: pool, jwt: jwt}
}

func normalizeEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func (e *Email) Register(ctx context.Context, email, password, name string) (string, error) {
	email = normalizeEmail(email)
	if email == "" || !strings.Contains(email, "@") {
		return "", errors.New("invalid email")
	}
	if len(password) < 8 {
		return "", ErrWeakPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	var id string
	err = e.pool.QueryRow(ctx, `
		INSERT INTO users (email, name, password_hash, provider, provider_id)
		VALUES ($1, $2, $3, $4, $1)
		RETURNING id
	`, email, name, string(hash), emailProvider).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), "users_email_key") || strings.Contains(err.Error(), "users_provider_provider_id_key") {
			return "", ErrEmailTaken
		}
		return "", err
	}
	return e.jwt.Issue(id, email)
}

func (e *Email) Login(ctx context.Context, email, password string) (string, error) {
	email = normalizeEmail(email)
	var id, hash string
	err := e.pool.QueryRow(ctx, `
		SELECT id, COALESCE(password_hash, '') FROM users WHERE email = $1
	`, email).Scan(&id, &hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}
	if hash == "" {
		return "", ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}
	return e.jwt.Issue(id, email)
}
