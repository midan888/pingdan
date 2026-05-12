package httpx

import (
	"context"
	"net/http"
	"strings"

	"github.com/pingdan/api/internal/auth"
)

type ctxKey string

const userCtxKey ctxKey = "user"

type AuthUser struct {
	ID    string
	Email string
}

func AuthMiddleware(j *auth.JWT) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if !strings.HasPrefix(h, "Bearer ") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			tok := strings.TrimPrefix(h, "Bearer ")
			claims, err := j.Parse(tok)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), userCtxKey, &AuthUser{ID: claims.UserID, Email: claims.Email})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserFrom(ctx context.Context) *AuthUser {
	u, _ := ctx.Value(userCtxKey).(*AuthUser)
	return u
}
